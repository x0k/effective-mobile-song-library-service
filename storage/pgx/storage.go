package pgx_storage

import (
	"context"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	connectionString string
	migrations       source.Driver
	conn             *pgx.Conn
}

func New(
	connectionString string,
	migrations source.Driver,
) *Storage {
	return &Storage{
		connectionString: connectionString,
		migrations:       migrations,
	}
}

func (s *Storage) Open(ctx context.Context) error {
	var err error
	s.conn, err = pgx.Connect(ctx, s.connectionString)
	return err
}

func (s *Storage) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}
