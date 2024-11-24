package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type stdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *stdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
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
	networkName := nw.Name

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:17.2-alpine3.20",
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "songs",
			},
			Networks: []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"postgres"},
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, pgContainer)

	appContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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
				"PG_CONNECTION_URI":          "postgres://test:test@postgres:5432/songs?sslmode=disable",
				"MUSIC_INFO_SERVICE_ADDRESS": "http://localhost:8081",
				"LOGGER_LEVEL":               "debug",
			},
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{
					&stdoutLogConsumer{},
				},
			},
		},
		Started: true,
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
