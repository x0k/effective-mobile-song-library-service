package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	microcks "microcks.io/testcontainers-go"
)

type tLogger struct {
	*testing.T
}

// Accept prints the log to stdout
func (lc *tLogger) Accept(l testcontainers.Log) {
	lc.Log(string(l.Content))
}

func setupApp(
	ctx context.Context,
	t *testing.T,
) string {

	nw, err := network.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupNetwork(t, nw)

	pgContainer, err := postgres.Run(ctx,
		"postgres:17.2-alpine3.20",
		network.WithNetwork([]string{"postgres"}, nw),
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

	microcksContainer, err := microcks.Run(
		ctx,
		"quay.io/microcks/microcks-uber:1.10.1",
		network.WithNetwork([]string{"music-info"}, nw),
		microcks.WithArtifact("../api/music-info.yml", true),
	)
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, microcksContainer)

	networkName := nw.Name
	appContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "..",
				Dockerfile: "Dockerfile",
			},
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForLog("press CTRL-C to exit"),
			Networks:     []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"app"},
			},
			Env: map[string]string{
				"LOGGER_LEVEL":               "debug",
				"PG_CONNECTION_URI":          "postgres://test:test@postgres:5432/songs?sslmode=disable",
				"MUSIC_INFO_SERVICE_ADDRESS": "http://music-info:8080/rest/Music+info/0.0.1",
			},
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{&tLogger{t}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, appContainer)

	mappedPort, err := appContainer.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatal(err)
	}
	host, err := appContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
}
