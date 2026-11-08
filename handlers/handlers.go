package handlers

import (
	"io"
	"tcp-server/config"
	"tcp-server/logger"
	"tcp-server/processors"
)

// handler doesn't close connections
// handler should read bytes slice, pass it to processor and write result
type Handler interface {
	Read(reader io.Reader) ([]byte, error)
	Process(req []byte) []byte
	Write(writer io.Writer, resp []byte) (int, error)
}

type HandlerImpl struct {
	config    config.Config
	processor processors.Processor
}

var _ Handler = HandlerImpl{}

func NewHandler(p processors.Processor, c config.Config) HandlerImpl {
	return HandlerImpl{processor: p, config: c}
}

func (h HandlerImpl) Read(reader io.Reader) ([]byte, error) {
	req, err := io.ReadAll(io.LimitReader(reader, int64(h.config.MaxBytes)))
	if err != nil {
		logger.LOG.Error("error reading", "error", err)
		return []byte{}, err
	}
	logger.LOG.Debug("request: " + string(req))
	return req, nil
}

func (h HandlerImpl) Process(req []byte) []byte {
	return h.processor.Process(req)
}

func (h HandlerImpl) Write(writer io.Writer, resp []byte) (int, error) {
	logger.LOG.Debug("response: " + string(resp))
	return writer.Write(resp)
}
