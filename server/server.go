package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"tcp-server/config"
	"tcp-server/handlers"
	"tcp-server/logger"
	"time"
)

type Server struct {
	isRunning   atomic.Bool
	mu          sync.Mutex
	listener    net.Listener
	connCounter atomic.Int32
	Config      config.Config
	wg          sync.WaitGroup
	Handler     handlers.Handler
	acceptDone  chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewServer(cfg config.Config, h handlers.Handler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		isRunning:  atomic.Bool{},
		Config:     cfg,
		Handler:    h,
		acceptDone: make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Server) Start(notifyContext context.Context) bool {
	if s == nil {
		logger.LOG.Error("nil server pointer")
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning.CompareAndSwap(false, true) {
		logger.LOG.Warn("server already started")
		return false
	}
	logger.LOG.Info("starting server...")

	listener, err := s.listen()
	if err != nil {
		//do smth else if Start() was called twice in a row
		logger.LOG.Error("listen failed", "port", s.Config.ServerPort, "error", err)
		return false
	}
	go s.startInternal(notifyContext, listener)
	return true
}

func (s *Server) startInternal(notifyContext context.Context, listener net.Listener) error {

	defer close(s.acceptDone)
	stop := context.AfterFunc(notifyContext, func() {
		logger.LOG.Info("stop listen")
		s.listener.Close()
	})
	defer stop()

	logger.LOG.Info("listen", "addr", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.LOG.Debug("connection closed. exit main loop.", "error", err)
				break
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
					s.handle(s.ctx, conn)
				})
			}
		}
	}
	return nil
}

func (s *Server) Stop(ctx context.Context, timeout time.Duration) bool {
	if s == nil {
		logger.LOG.Error("nil server pointer")
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning.CompareAndSwap(true, false) {
		logger.LOG.Warn("server already stoped")
		return false
	}
	<-s.acceptDone
	gracefull := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(gracefull)
	}()
	select {
	case <-gracefull:
		logger.LOG.Info("all handlers finished")
		s.cancel()
	case <-time.After(timeout):
		logger.LOG.Warn("finished by timeout")
		s.cancel()
	}
	logger.LOG.Info("closing resources")
	<-time.After(time.Duration(1_000_000_000)) //wait a little while resources are being closed and events are being logged
	logger.LOG.Info("server stoped")
	return true
}

func (s *Server) listen() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+s.Config.ServerPort)
	s.listener = listener
	return listener, err
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	logger.LOG.Debug("handle connection")
	//overall socket call timeout
	connCtx, cancel := context.WithTimeout(ctx, s.Config.ConnTimeout)
	defer cancel()
	stop := context.AfterFunc(connCtx, func() {
		//TODO write error before deadline
		conn.SetDeadline(time.Now())
	})
	defer stop()

	req, errorRead := s.Handler.Read(conn)
	if errorRead != nil {
		logger.LOG.Error("error read", "error", errorRead)
		return
	}

	resp := s.Handler.Process(req)

	_, errorWrite := s.Handler.Write(conn, resp)
	if errorWrite != nil {
		logger.LOG.Error("error write", "error", errorWrite)
	}

}
