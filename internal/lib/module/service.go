package module

import "context"

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

type service struct {
	name  string
	start func(ctx context.Context) error
}

func NewService(name string, start func(ctx context.Context) error) Service {
	return &service{name: name, start: start}
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Start(ctx context.Context) error {
	return s.start(ctx)
}
