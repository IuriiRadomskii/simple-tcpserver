package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerPort            string        `env:"SERVER_PORT" envDefault:"55555"`
	MaxBytes              uint          `env:"MAX_BYTES" envDefault:"1000000"`
	MaxConnections        uint          `env:"MAX_CONNS" envDefault:"16"`
	AcceptRetry           time.Duration `env:"ACCEPT_RETRY" envDefault:"250ms"`
	ServerShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ReadTimeout           time.Duration `env:"READ_TIMEOUT"  envDefault:"5s"`
	WriteTimeout          time.Duration `env:"WRITE_TIMEOUT" envDefault:"5s"`
	ConnTimeout           time.Duration `env:"CONN_TIMEOUT"  envDefault:"60s"`
}

func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

func LoadTestConfig() (Config, error) {
	return Config{
		ServerPort:            "8080",
		MaxBytes:              10,
		MaxConnections:        8,
		AcceptRetry:           time.Millisecond * 250,
		ServerShutdownTimeout: time.Second * 10,
		ReadTimeout:           time.Second * 5,
		WriteTimeout:          time.Second * 5,
		ConnTimeout:           time.Second * 60,
	}, nil
}
