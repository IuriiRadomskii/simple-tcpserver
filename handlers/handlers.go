package handlers

import (
	"context"
	"io"
	"net"
	"tcp-server/logger"
	"tcp-server/processors"
	"time"
)

const (
	MAX_BYTES    = 1024
	CALL_TIMEOUT = 1 * time.Second
)

// handler doesn't close connections
// handler should read bytes slice, pass it to processor and write result
type Handler struct {
	Processor processors.Processor
}

func NewHandler(p processors.Processor) Handler {
	return Handler{Processor: p}
}

func (h Handler) Handle(ctx context.Context, conn net.Conn) {
	conn.SetDeadline(time.Now().Add(CALL_TIMEOUT))

	req, err := io.ReadAll(io.LimitReader(conn, MAX_BYTES))
	if err != nil {
		logger.LOG.Error("error reading", "error", err)
		return
	}
	result, err := h.Processor.Process(req)
	if err != nil {
		logger.LOG.Error("Processing error", "error", err)
		conn.Write([]byte("processing error"))
	} else {
		conn.Write(result)
	}
}
