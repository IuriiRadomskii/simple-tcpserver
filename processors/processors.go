package processors

import (
	"time"
)

const (
	NOOP = "\n--------------\nNOOP\n--------------\n"
)

// processor accepts raw bytes, does work and return result to handler
type Processor interface {
	Process(req []byte) []byte
}

type ProcessorImpl struct{}

var _ Processor = ProcessorImpl{}

func NewProcessor() ProcessorImpl {
	return ProcessorImpl{}
}

func (p ProcessorImpl) Process(req []byte) []byte {
	input := string(req)
	switch input {
	case "FAST":
		time.Sleep(1_000_000_000 * 1)
	case "SLOW":
		time.Sleep(1_000_000_000 * 5)
	default:
		func() {}()
	}
	return []byte(NOOP)
}
