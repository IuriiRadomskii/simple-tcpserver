package client

import (
	"context"
	"errors"
	"io"
	"net"
	"tcp-server/config"
	"tcp-server/logger"
	"time"
)

var ErrResponseTooLarge = errors.New("response exceeds max bytes")

type TCPClient struct {
	Config config.ClientConfig
	dialer net.Dialer
}

func NewTCPClient(cfg config.ClientConfig) *TCPClient {
	return &TCPClient{
		Config: cfg,
		dialer: net.Dialer{Timeout: cfg.DialTimeout},
	}
}

func (c *TCPClient) Send(ctx context.Context, req []byte) ([]byte, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, c.Config.RequestTimeout)
	defer cancel()

	conn, err := c.dialer.DialContext(ctx, "tcp", c.Config.ServerAddrPort.String())
	if err != nil {
		return nil, errors.Join(errors.New("dial error"), err)
	}
	defer conn.Close()
	logger.LOG.Debug("client connected", "addr", conn.RemoteAddr().String())

	stop := context.AfterFunc(ctx, func() {
		conn.SetDeadline(time.Now())
	})
	defer stop()

	if _, err := conn.Write(req); err != nil {
		return nil, c.wrap(ctx, "write error", err)
	}

	//send half-close
	if err := conn.(*net.TCPConn).CloseWrite(); err != nil {
		return nil, c.wrap(ctx, "close write error", err)
	}

	limit := int64(c.Config.MaxRespBytes)
	resp, err := io.ReadAll(io.LimitReader(conn, limit+1))
	if err != nil {
		return nil, c.wrap(ctx, "read error", err)
	}
	if int64(len(resp)) > limit {
		return nil, ErrResponseTooLarge
	}
	return resp, nil
}

func (c *TCPClient) wrap(ctx context.Context, msg string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return errors.Join(errors.New(msg), ctxErr, err)
	}
	return errors.Join(errors.New(msg), err)
}
