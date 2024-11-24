package db_storage

import (
	"context"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/db"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
)

type Storage struct {
	log           *logger.Logger
	connectionURI string
	conn          *pgx.Conn
	queries       *db.Queries
}

func New(
	log *logger.Logger,
	connectionURI string,
) *Storage {
	return &Storage{
		log:           log,
		connectionURI: connectionURI,
	}
}

func (s *Storage) Open(ctx context.Context) error {
	var err error
	s.conn, err = pgx.Connect(ctx, s.connectionURI)
	if err != nil {
		return err
	}
	s.queries = db.New(s.conn)
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}
