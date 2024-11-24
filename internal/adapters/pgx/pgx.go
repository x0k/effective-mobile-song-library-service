package pgx_adapter

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
)

func NewPgx(m module.Interface, connectionURI string) *pgx.Conn {
	conn := &pgx.Conn{}
	m.PreStartFn("pgx_connect", func(ctx context.Context) error {
		c, err := pgx.Connect(ctx, connectionURI)
		if err == nil {
			*conn = *c
		}
		return err
	})
	m.PostStopFn("pgx_close", conn.Close)
	return conn
}
