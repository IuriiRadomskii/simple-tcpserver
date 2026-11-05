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
)

func main() {
	run()
}

func run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, _ := config.LoadTestConfig()

	handler := handlers.NewHandler(processors.NewProcessor(), config)

	server := &server.Server{
		Config:  config,
		Handler: handler,
	}

	go server.Start(ctx)

	<-ctx.Done() //main blocked until os signall is arrived
	logger.LOG.Info("shutting down...")

	server.Stop(ctx, config.ServerShutdownTimeout)
}
