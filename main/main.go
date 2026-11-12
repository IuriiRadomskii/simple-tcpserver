package main

import (
	"context"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"tcp-server/config"
	"tcp-server/handlers"
	"tcp-server/logger"
	"tcp-server/processors"
	"tcp-server/server"
	"time"
)

func main() {
	run()
}

func run() {
	notifyContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := config.Config{
		AddrPort:              netip.MustParseAddrPort("127.0.0.1:8080"),
		MaxBytes:              10,
		MaxConnections:        8,
		AcceptRetry:           time.Millisecond * 250,
		ServerShutdownTimeout: time.Second * 10,
		ConnTimeout:           time.Second * 60,
	}

	handler := handlers.NewHandler(processors.NewProcessor(), config)

	server := server.NewTCPServer(config, handler)

	cancelCtx, cancel := context.WithCancel(context.Background())

	go server.Start(cancelCtx)

	<-notifyContext.Done() //main blocked until os signall is arrived
	logger.LOG.Info("shutting down...")

	server.Stop(cancel, config.ServerShutdownTimeout)
}
