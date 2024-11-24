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

type PgConfig struct {
	ConnectionURI string `env:"PG_CONNECTION_URI" env-required:"true"`
}

type ServerConfig struct {
	Address string `env:"SERVER_ADDRESS" env-default:"0.0.0.0:8080"`
}

type Config struct {
	Logger           LoggerConfig
	MusicInfoService MusicInfoServiceConfig
	Postgres         PgConfig
	Server           ServerConfig
}

func mustLoadConfig() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("can't load config: %s", err)
	}
	return cfg
}
