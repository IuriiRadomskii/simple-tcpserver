package processors

import (
	"fmt"
	"tcp-server/logger"
)

type Processor struct{}

func NewProcessor() Processor {
	return Processor{}
}

func (p Processor) Process(req []byte) ([]byte, error) {
	logger.LOG.Info(fmt.Sprintf("Input: %v", string(req)))
	return []byte("NOOP"), nil
}
