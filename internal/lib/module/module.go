package module

import (
	"context"
	"sync"
	"sync/atomic"

	"log/slog"
)

type Interface interface {
	Fataler
	PreStarter
	PostStarter
	PreStopper
	PostStopper
	Service
}

type Module struct {
	name      string
	log       *slog.Logger
	wg        sync.WaitGroup
	services  []Service
	preStart  []Hook
	postStart []Hook
	preStop   []Hook
	postStop  []Hook
	fatal     chan error
	stopped   atomic.Bool
}

func New(log *slog.Logger, name string) *Module {
	return &Module{
		log:   log,
		name:  name,
		fatal: make(chan error, 1),
	}
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) Append(services ...Service) {
	m.services = append(m.services, services...)
}

func (m *Module) AppendServiceFn(name string, run func(ctx context.Context) error) {
	m.services = append(m.services, NewService(name, run))
}

func (m *Module) PreStart(hooks ...Hook) {
	m.preStart = append(m.preStart, hooks...)
}

func (m *Module) PreStartFn(name string, run func(ctx context.Context) error) {
	m.preStart = append(m.preStart, NewHook(name, run))
}

func (m *Module) PostStart(hooks ...Hook) {
	m.postStart = append(m.postStart, hooks...)
}

func (m *Module) PreStop(hooks ...Hook) {
	m.preStop = append(m.preStop, hooks...)
}

func (m *Module) PostStop(hooks ...Hook) {
	m.postStop = append(m.postStop, hooks...)
}

func (m *Module) PostStopFn(name string, run func(ctx context.Context) error) {
	m.postStop = append(m.postStop, NewHook(name, run))
}

func (m *Module) Fatal(ctx context.Context, err error) {
	if m.stopped.Swap(true) {
		m.log.LogAttrs(ctx, slog.LevelError, "fatal error", slog.String("error", err.Error()))
		return
	}
	m.fatal <- err
	close(m.fatal)
}

func (m *Module) awaiter(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-m.fatal:
		return err
	}
}

func (m *Module) hooks(ctx context.Context, hooks []Hook, msg string) {
	for _, hook := range hooks {
		m.log.LogAttrs(ctx, slog.LevelInfo, msg, slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.Fatal(ctx, err)
		}
	}
}

func (m *Module) start(ctx context.Context, awaiter func(context.Context) error) error {
	if len(m.services) == 0 && len(m.postStart) == 0 && len(m.preStop) == 0 {
		return nil
	}

	if m.stopped.Load() {
		return <-m.fatal
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	m.hooks(ctx, m.preStart, "run pre start")

	for _, service := range m.services {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.log.LogAttrs(ctx, slog.LevelInfo, "starting", slog.String("service", service.Name()))
			if err := service.Start(ctx); err != nil {
				m.Fatal(ctx, err)
			}
			m.log.LogAttrs(ctx, slog.LevelInfo, "stopped", slog.String("service", service.Name()))
		}()
	}

	m.hooks(ctx, m.postStart, "run post start")

	err := awaiter(ctx)

	m.hooks(ctx, m.preStop, "run pre stop")

	m.log.LogAttrs(ctx, slog.LevelInfo, "stopping")
	m.stopped.Store(true)
	cancel()

	m.hooks(ctx, m.postStop, "run post stop")

	m.wg.Wait()

	return err
}

func (m *Module) Start(ctx context.Context) error {
	return m.start(ctx, m.awaiter)
}
