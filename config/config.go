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
	ConnTimeout           time.Duration `env:"CONN_TIMEOUT"  envDefault:"60s"`
}

func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
