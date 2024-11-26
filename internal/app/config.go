package app

import (
	"log"
	"os"

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
	MigrationsURI string `env:"PG_MIGRATIONS_URI" env-default:"file://migrations"`
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

func mustLoadConfig(configPath string) *Config {
	cfg := &Config{}
	var cfgErr error
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfgErr = cleanenv.ReadEnv(cfg)
	} else if err == nil {
		cfgErr = cleanenv.ReadConfig(configPath, cfg)
	} else {
		cfgErr = err
	}
	if cfgErr != nil {
		log.Fatalf("cannot read config: %s", cfgErr)
	}
	return cfg
}
