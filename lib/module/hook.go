package module

import (
	"context"
)

type Hook interface {
	Name() string
	Run(ctx context.Context) error
}

type PreStopper interface {
	PreStop(hooks ...Hook)
}

type PostStopper interface {
	PostStop(hooks ...Hook)
}

type hook struct {
	name string
	run  func(ctx context.Context) error
}

func NewHook(name string, run func(ctx context.Context) error) Hook {
	return &hook{name: name, run: run}
}

func (h *hook) Name() string {
	return h.name
}

func (h *hook) Run(ctx context.Context) error {
	return h.run(ctx)
}
