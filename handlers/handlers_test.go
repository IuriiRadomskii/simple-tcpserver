package handlers

import (
	"bytes"
	"errors"
	"tcp-server/config"
	"tcp-server/processors"
	"testing"
	"testing/iotest"
)

func newHandler(maxBytes uint) HandlerImpl {
	return NewHandler(processors.NewProcessor(), config.Config{
		MaxBytes: maxBytes,
	})
}

func TestRead(t *testing.T) {
	h := newHandler(10)
	requestBytes := []byte("test")
	reader := bytes.NewReader(requestBytes)
	res, _ := h.Read(reader)
	if !bytes.Equal(res, requestBytes) {
		t.Fatal("reader fail", "actual:", string(res), "expected:", string(requestBytes))
	}
}

func TestReadMaxBytesExceeded(t *testing.T) {
	maxBytes := 2
	h := newHandler(uint(maxBytes))
	requestBytes := []byte("test")
	reader := bytes.NewReader(requestBytes)
	res, _ := h.Read(reader)
	if !bytes.Equal(res, requestBytes[:maxBytes]) {
		t.Fatal("reader fail", "| actual:", string(res), "| expected:", string(requestBytes))
	}
}

func TestReadError(t *testing.T) {
	errRead := errors.New("test read error")
	h := newHandler(uint(100))
	reader := iotest.ErrReader(errRead)
	_, err := h.Read(reader)
	if err == nil {
		t.Fatal("handler read should return err on reader error", "| expected error:", errRead.Error())
	}

}

func TestProcess(t *testing.T) {
	h := newHandler(10)
	res := h.Process([]byte("test"))
	if !bytes.Equal(res, []byte(processors.NOOP)) {
		t.Fatal("process fail", "| actual:", string(res), "| expected:", processors.NOOP)
	}
}

func TestWrite(t *testing.T) {
	h := newHandler(10)
	responseBytes := []byte("response")
	var writer bytes.Buffer
	n, err := h.Write(&writer, responseBytes)
	if err != nil {
		t.Fatal("write fail", "| error:", err)
	}
	if n != len(responseBytes) {
		t.Fatal("write fail", "| actual bytes:", n, "| expected bytes:", len(responseBytes))
	}
	if !bytes.Equal(writer.Bytes(), responseBytes) {
		t.Fatal("write fail", "| actual:", writer.String(), "| expected:", string(responseBytes))
	}
}

// errWriter returns err on every Write call
type errWriter struct {
	err error
}

func (w errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestWriteError(t *testing.T) {
	errWrite := errors.New("test write error")
	h := newHandler(10)
	_, err := h.Write(errWriter{err: errWrite}, []byte("response"))
	if !errors.Is(err, errWrite) {
		t.Fatal("handler write should return err on writer error", "| actual:", err, "| expected:", errWrite)
	}
}
