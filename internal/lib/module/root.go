package module

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log/slog"
)

type Root struct {
	*Module
}

func NewRoot(log *slog.Logger) *Root {
	return &Root{
		Module: New(log, "root"),
	}
}

func (r *Root) awaiter(ctx context.Context) error {
	r.log.LogAttrs(ctx, slog.LevelInfo, "press CTRL-C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-stop:
		r.log.LogAttrs(ctx, slog.LevelInfo, "received signal", slog.String("signal", s.String()))
		return nil
	case err := <-r.fatal:
		return err
	}
}

func (r *Root) Start(ctx context.Context) error {
	return r.start(ctx, r.awaiter)
}
