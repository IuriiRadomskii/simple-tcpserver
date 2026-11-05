package handlers

import (
	"io"
	"tcp-server/config"
	"tcp-server/logger"
	"tcp-server/processors"
)

// handler doesn't close connections
// handler should read bytes slice, pass it to processor and write result
type Handler struct {
	config    config.Config
	processor processors.Processor
}

func NewHandler(p processors.Processor, c config.Config) Handler {
	return Handler{processor: p, config: c}
}

func (h Handler) Read(reader io.Reader) ([]byte, error) {
	req, err := io.ReadAll(io.LimitReader(reader, int64(h.config.MaxBytes)))
	if err != nil {
		logger.LOG.Error("error reading", "error", err)
		return []byte{}, err
	}
	logger.LOG.Debug("request: " + string(req))
	return req, nil
}

func (h Handler) Process(req []byte) []byte {
	return h.processor.Process(req)
}

func (h Handler) Write(writer io.Writer, resp []byte) (int, error) {
	logger.LOG.Debug("response: " + string(resp))
	return writer.Write(resp)
}
