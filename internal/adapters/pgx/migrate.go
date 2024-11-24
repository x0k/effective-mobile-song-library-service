package pgx_adapter

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	mpgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
)

const (
	_defaultAttempts = 3
	_defaultTimeout  = time.Second
)

func NewMigrator(
	name string,
	log *slog.Logger,
	connectionURI string,
	migrationsURI string,
) module.Hook {
	return module.NewHook(name, func(ctx context.Context) error {
		var (
			attempts = _defaultAttempts
			err      error
			db       *sql.DB
			driver   database.Driver
			m        *migrate.Migrate
		)

		l := log.With(slog.String("operation", "migrate"))
		for attempts > 0 {
			if db, err = sql.Open("pgx", connectionURI); err == nil {
				if driver, err = mpgx.WithInstance(db, &mpgx.Config{}); err == nil {
					if m, err = migrate.NewWithDatabaseInstance(
						migrationsURI,
						"pgx5",
						driver,
					); err == nil {
						break
					}
				}
			}
			l.LogAttrs(
				ctx,
				slog.LevelWarn,
				"can't connect",
				slog.String("error", err.Error()),
				slog.Int("attempts_left", attempts),
			)
			time.Sleep(_defaultTimeout)
			attempts--
		}
		if err != nil {
			return err
		}
		err = m.Up()
		defer m.Close()

		if errors.Is(err, migrate.ErrNoChange) {
			l.LogAttrs(ctx, slog.LevelInfo, "no change")
			return nil
		}
		if err != nil {
			l.LogAttrs(ctx, slog.LevelError, "up error", slog.String("error", err.Error()))
			return err
		}
		l.LogAttrs(ctx, slog.LevelInfo, "migrated successfully")
		return nil
	})
}
