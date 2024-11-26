package testutils

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgx_adapter "github.com/x0k/effective-mobile-song-library-service/internal/adapters/pgx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
	microcks "microcks.io/testcontainers-go"
)

func SetupPgx(ctx context.Context, log *slog.Logger, t testing.TB) *pgx.Conn {
	pgContainer, err := postgres.Run(ctx,
		"postgres:17.2-alpine3.20",
		postgres.WithDatabase("songs"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, pgContainer)

	uri, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := pgx_adapter.Migrate(ctx, log, uri, "file://../../migrations"); err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(ctx); err != nil {
			log.LogAttrs(ctx, slog.LevelError, "cannot close connection", slog.String("error", err.Error()))
		}
	})
	return conn
}

func SetupMusicInfoClient(ctx context.Context, t testing.TB) *music_info.ClientWithResponses {
	microcksContainer, err := microcks.Run(
		ctx,
		"quay.io/microcks/microcks-uber:1.10.1",
		microcks.WithArtifact("../../api/music-info.yml", true),
	)
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, microcksContainer)

	endpoint, err := microcksContainer.RestMockEndpoint(ctx, "Music+info", "0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	client, err := music_info.NewClientWithResponses(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
