package main

import (
	"context"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := config.Config{
		ServerPort:            "8080",
		MaxBytes:              10,
		MaxConnections:        8,
		AcceptRetry:           time.Millisecond * 250,
		ServerShutdownTimeout: time.Second * 10,
		ConnTimeout:           time.Second * 60,
	}

	handler := handlers.NewHandler(processors.NewProcessor(), config)

	server := server.NewServer(config, handler)

	go server.Start(ctx)

	<-ctx.Done() //main blocked until os signall is arrived
	logger.LOG.Info("shutting down...")

	server.Stop(ctx, config.ServerShutdownTimeout)
}
