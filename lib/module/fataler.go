package module

import "context"

type Fataler interface {
	Fatal(ctx context.Context, err error)
}
