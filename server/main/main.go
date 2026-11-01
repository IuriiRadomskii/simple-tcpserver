package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"signal"
	"syscall"
	"tcp-server/handlers"
	"tcp-server/logger"
	"tcp-server/processors"
	"tcp-server/server"
	"time"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	port := os.Getenv("SERVER_PORT")

	if port == "" {
		logger.LOG.Error("Server port is not specified")
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.LOG.Error("listen failed", "port", 55555, "error", err)
		os.Exit(1)
	}

	defer listener.Close()

	handler := handlers.NewHandler(processors.NewProcessor())

	server := &server.TcpServer{
		Handler: handler,
	}

	go server.Start(ctx, listener)

	<-ctx.Done()
	logger.LOG.Info("shutting down")
	server.Stop(10 * time.Second)

}
