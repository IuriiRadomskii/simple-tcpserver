package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"tcp-server/handlers"
	"tcp-server/logger"
	"time"
)

const (
	WAIT = 250
)

type TcpServer struct {
	wg      sync.WaitGroup
	Handler handlers.Handler
}

func (s *TcpServer) Start(ctx context.Context, listener net.Listener) {
	stop := context.AfterFunc(ctx, func() { listener.Close() })
	defer stop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.LOG.Error("Connection closed. Exit main loop.", "error", err)
				return
			}
			time.Sleep(WAIT * time.Millisecond)
			continue
		} else {
			s.wg.Go(func() {
				defer conn.Close()
				s.Handler.Handle(ctx, conn)
			})
		}
	}
}

func (s *TcpServer) Stop(timeout time.Duration) {
	gracefull := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(gracefull)
	}()
	select {
	case <-gracefull:
		logger.LOG.Info("all handlers finished")
	case <-time.After(timeout):
		logger.LOG.Warn("finished by timeout")
	}
}
