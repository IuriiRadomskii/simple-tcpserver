package processors

import (
	"time"
)

// processor accepts raw bytes, does work and return result to handler
type Processor struct{}

func NewProcessor() Processor {
	return Processor{}
}

func (p Processor) Process(req []byte) []byte {
	input := string(req)
	switch input {
	case "FAST":
		time.Sleep(1_000_000_000 * 1)
	case "SLOW":
		time.Sleep(1_000_000_000 * 5)
	default:
		func() {}()
	}
	return []byte("\n--------------\nNOOP\n--------------\n")
}
