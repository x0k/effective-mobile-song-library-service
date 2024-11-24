//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

const CONNECTION_URI_ENV = "PG_CONNECTION_URI"
const MIGRATIONS_PATH_ENV = "MIGRATIONS_PATH"

func init() {
	connectionURI, ok := os.LookupEnv(CONNECTION_URI_ENV)
	if !ok || len(connectionURI) == 0 {
		log.Fatalf("migrate: environment variable not declared: %s", CONNECTION_URI_ENV)
	}
	migrationsPath, ok := os.LookupEnv(MIGRATIONS_PATH_ENV)
	if !ok || len(migrationsPath) == 0 {
		migrationsPath = "db/migrations"
	}

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New(
			fmt.Sprintf("file://%s", migrationsPath),
			connectionURI,
		)
		if err == nil {
			break
		}

		log.Printf("Migrate: can't connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	if err != nil {
		log.Fatalf("Migrate: up error: %s", err)
	}

	log.Printf("Migrate: up success")
}
