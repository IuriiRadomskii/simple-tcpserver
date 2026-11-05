package server

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"tcp-server/config"
	"tcp-server/handlers"
	"tcp-server/logger"
	"time"
)

type Server struct {
	listener       net.Listener
	connCounter    atomic.Int32
	Config         config.Config
	wg             sync.WaitGroup
	Handler        handlers.Handler
	acceptDone     chan struct{}
	connParentCtx  context.Context
	connCancelFunc context.CancelFunc
}

func (s *Server) Start(ctx context.Context) {
	logger.LOG.Info("starting server...")
	listener, err := s.listen()
	if err != nil {
		//do smth else if Start() was called twice in a row
		logger.LOG.Error("listen failed", "port", s.Config.ServerPort, "error", err)
		os.Exit(1)
	}

	stop := context.AfterFunc(ctx, func() {
		logger.LOG.Info("stop listen")
		s.listener.Close()
	})
	defer stop()

	logger.LOG.Info("listen", "addr", listener.Addr().String())

	s.acceptDone = make(chan struct{})
	defer close(s.acceptDone)
	s.connParentCtx, s.connCancelFunc = context.WithCancel(context.Background())

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.LOG.Debug("connection closed. exit main loop.", "error", err)
				return
			}
			time.Sleep(s.Config.AcceptRetry)
			continue
		} else {
			curr := s.connCounter.Add(1)
			if curr > int32(s.Config.MaxConnections) {
				//TODO pass conn to simple error handler which returns "server is busy"
				logger.LOG.Warn("max connection exceeded")
				conn.Close()
				s.connCounter.Add(-1)
			} else {
				s.wg.Go(func() {
					logger.LOG.Debug("open connection")
					defer func() {
						conn.Close()
						s.connCounter.Add(-1)
						logger.LOG.Debug("close connection")
					}()
					s.handle(s.connParentCtx, conn)
				})
			}
		}
	}
}

func (s *Server) Stop(ctx context.Context, timeout time.Duration) {
	gracefull := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(gracefull)
	}()
	select {
	case <-gracefull:
		logger.LOG.Info("all handlers finished")
		s.connCancelFunc()
	case <-time.After(timeout):
		logger.LOG.Warn("finished by timeout")
		s.connCancelFunc()
		<-gracefull
	}
	logger.LOG.Info("server stopeed")
}

func (s *Server) listen() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+s.Config.ServerPort)
	s.listener = listener
	return listener, err
}

func (s *Server) handle(connParentCtx context.Context, conn net.Conn) {
	logger.LOG.Debug("handle connection")
	connTimeoutCtx, cancel := context.WithTimeout(connParentCtx, s.Config.ConnTimeout)
	defer cancel()

	stop := context.AfterFunc(connTimeoutCtx, func() { conn.SetDeadline(time.Now()) })
	defer stop()

	conn.SetReadDeadline(currDeadline(connTimeoutCtx, s.Config.ReadTimeout))
	if connTimeoutCtx.Err() != nil {
		logger.LOG.Error("connCtx.Err")
		return
	}
	req, errorRead := s.Handler.Read(conn)
	if errorRead != nil {
		logger.LOG.Error("error read")
		return
	}

	resp := s.Handler.Process(req)

	conn.SetWriteDeadline(currDeadline(connTimeoutCtx, s.Config.WriteTimeout))
	if connTimeoutCtx.Err() != nil {
		return
	}
	_, errorWrite := s.Handler.Write(conn, resp)
	if errorWrite != nil {
		logger.LOG.Error("error write")
	}
}

func currDeadline(ctx context.Context, d time.Duration) time.Time {
	t := time.Now().Add(d)
	if dl, ok := ctx.Deadline(); ok && dl.Before(t) {
		return dl
	}
	return t
}
