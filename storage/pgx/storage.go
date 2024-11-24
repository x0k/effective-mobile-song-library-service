package pgx_storage

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	connectionString string
	sourceName       string
	migrations       source.Driver
	conn             *pgx.Conn
}

func New(
	connectionString string,
	sourceName string,
	migrations source.Driver,
) *Storage {
	return &Storage{
		connectionString: connectionString,
		sourceName:       sourceName,
		migrations:       migrations,
	}
}

func (s *Storage) Open(ctx context.Context) error {
	migrator, err := migrate.NewWithSourceInstance(s.sourceName, s.migrations, s.connectionString)
	if err != nil {
		return err
	}
	if err := migrator.Up(); err != nil {
		return err
	}
	s.conn, err = pgx.Connect(ctx, s.connectionString)
	return err
}

func (s *Storage) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}
