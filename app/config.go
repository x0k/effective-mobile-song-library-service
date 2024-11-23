package app

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	Level       string `env:"LOGGER_LEVEL" env-default:"info"`
	HandlerType string `env:"LOGGER_HANDLER_TYPE" env-default:"text"`
}

type MusicInfoServiceConfig struct {
	Address string `env:"MUSIC_INFO_SERVICE_ADDRESS" env-required:"true"`
}

type PostgresConfig struct {
	ConnectionString string `env:"POSTGRES_CONNECTION_STRING" env-required:"true"`
}

type Config struct {
	Logger           LoggerConfig
	MusicInfoService MusicInfoServiceConfig
	Postgres         PostgresConfig
}

func mustLoadConfig() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("can't load config: %s", err)
	}
	return cfg
}
