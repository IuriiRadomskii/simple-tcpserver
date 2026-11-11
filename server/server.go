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

type state int

const (
	stateNew state = iota
	stateRunning
	stateStopped
)

type TCPServer struct {
	state       state
	mu          sync.Mutex
	listener    *net.TCPListener
	connCounter atomic.Int32
	Config      config.Config
	wg          sync.WaitGroup
	Handler     handlers.Handler
	acceptDone  chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewServer(cfg config.Config, h handlers.Handler) *TCPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPServer{
		Config:     cfg,
		Handler:    h,
		acceptDone: make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *TCPServer) Start(ctx context.Context) (bool, error) {
	if s == nil {
		return false, errors.New("nil server")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case stateRunning:
		return false, errors.New("server aready running")
	case stateStopped:
		return false, errors.New("server was stopped")
	}
	logger.LOG.Info("server is starting...")
	listener, err := s.listen()
	if err != nil {
		return false, errors.Join(errors.New("listener error"), err)
	}
	s.listener = listener
	s.state = stateRunning
	go s.startInternal(ctx, listener)
	return true, nil
}

func (s *TCPServer) startInternal(notifyContext context.Context, listener net.Listener) {
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
				logger.LOG.Info("connection closed. exit main loop.", "error", err)
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
	logger.LOG.Debug("exit main loop")
}

func (s *TCPServer) Stop(cancel context.CancelFunc, timeout time.Duration) (bool, error) {
	if s == nil {
		return false, errors.New("nil server")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case stateNew:
		return false, errors.New("server not running")
	case stateStopped:
		return false, errors.New("already stopped")
	}
	logger.LOG.Info("stopping server")
	cancel()
	<-s.acceptDone
	logger.LOG.Info("accept done")
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
	s.state = stateStopped
	logger.LOG.Info("server stoped")
	return true, nil
}

func (s *TCPServer) listen() (*net.TCPListener, error) {
	tcpAddr := net.TCPAddrFromAddrPort(s.Config.AddrPort)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	return listener, err
}

func (s *TCPServer) handle(ctx context.Context, conn net.Conn) {
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
