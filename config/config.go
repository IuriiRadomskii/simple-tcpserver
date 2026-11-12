package config

import (
	"net/netip"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	DEF_SERVER_ADDR      = "127.0.0.1:55555"
	DEF_MAX_BYTES        = 1000000
	DEF_MAX_CONNS        = 16
	DEF_ACCEPT_RETRY     = "250ms"
	DEF_SHUTDOWN_TIMEOUT = "10s"
	DEF_CONN_TIMEOUT     = "60s"
)

type Config struct {
	AddrPort              netip.AddrPort `env:"SERVER_ADDR" envDefault:"127.0.0.1:55555"`
	MaxBytes              uint           `env:"MAX_BYTES" envDefault:"1000000"`
	MaxConnections        uint           `env:"MAX_CONNS" envDefault:"16"`
	AcceptRetry           time.Duration  `env:"ACCEPT_RETRY" envDefault:"250ms"`
	ServerShutdownTimeout time.Duration  `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ConnTimeout           time.Duration  `env:"CONN_TIMEOUT"  envDefault:"60s"`
}

func LoadConfig() (Config, error) {
	var err error
	var addrPort netip.AddrPort
	var acceptRetry, serverShutdownTimeout, connTimeout time.Duration

	addrPort, err = netip.ParseAddrPort(DEF_SERVER_ADDR)
	if err != nil {
		return Config{}, err
	}
	acceptRetry, err = time.ParseDuration(DEF_ACCEPT_RETRY)
	if err != nil {
		return Config{}, err
	}
	serverShutdownTimeout, err = time.ParseDuration(DEF_SHUTDOWN_TIMEOUT)
	if err != nil {
		return Config{}, err
	}
	connTimeout, err = time.ParseDuration(DEF_CONN_TIMEOUT)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		AddrPort:              addrPort,
		MaxBytes:              DEF_MAX_BYTES,
		MaxConnections:        DEF_MAX_CONNS,
		AcceptRetry:           acceptRetry,
		ServerShutdownTimeout: serverShutdownTimeout,
		ConnTimeout:           connTimeout,
	}
	err = env.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type ClientConfig struct {
	ServerAddrPort netip.AddrPort `env:"CLIENT_SERVER_ADDR" envDefault:"127.0.0.1:55555"`
	MaxRespBytes   uint           `env:"CLIENT_MAX_RESP_BYTES" envDefault:"1000000"`
	DialTimeout    time.Duration  `env:"CLIENT_DIAL_TIMEOUT" envDefault:"5s"`
	RequestTimeout time.Duration  `env:"CLIENT_REQUEST_TIMEOUT" envDefault:"60s"`
}

// LoadClientConfig builds client config from envDefault tags overridden by env variables
func LoadClientConfig() (ClientConfig, error) {
	var cfg ClientConfig
	if err := env.Parse(&cfg); err != nil {
		return ClientConfig{}, err
	}
	return cfg, nil
}
