package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/netip"
	"tcp-server/config"
	"tcp-server/handlers"
	"tcp-server/logger"
	"testing"
	"time"
)

func TestNilServerNotPanicsOnStart(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v\n", r)
		}
	}()
	var server *TCPServer
	ok, _ := server.Start(context.Background())
	if ok {
		t.Fatal("nil server should return false on Start")
	}
}

func TestServerStarts(t *testing.T) {
	config := config.Config{
		AddrPort: netip.MustParseAddrPort("127.0.0.1:8080"),
	}
	server := NewTCPServer(config, handlers.HandlerImpl{})
	ctx, cancel := context.WithCancel(context.Background())
	started, err := server.Start(ctx)
	if !started {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}
	t.Cleanup(func() {
		cancel()
	})
}

func TestServerConsecutiveStartsReturnsFalse(t *testing.T) {
	config := config.Config{
		AddrPort: netip.MustParseAddrPort("127.0.0.1:8080"),
	}
	server := NewTCPServer(config, handlers.HandlerImpl{})
	ctx, cancel := context.WithCancel(context.Background())
	started, err := server.Start(ctx)
	started, err = server.Start(ctx)
	if started {
		t.Fail()
	}
	if err == nil {
		t.Fail()
	}
	t.Cleanup(func() {
		cancel()
	})
}

func TestServerStartOnOccupiedPortFails(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	config := config.Config{
		AddrPort: netip.MustParseAddrPort("127.0.0.1:" + port),
	}
	server := NewTCPServer(config, handlers.HandlerImpl{})
	ctx, cancel := context.WithCancel(context.Background())
	started, _ := server.Start(ctx)
	if started {
		t.Fail()
	}
	t.Cleanup(func() {
		listener.Close()
		cancel()
	})
}

func TestServerStop(t *testing.T) {
	config := config.Config{
		AddrPort: netip.MustParseAddrPort("127.0.0.1:0"),
	}
	server := NewTCPServer(config, handlers.HandlerImpl{})
	ctx, cancel := context.WithCancel(context.Background())
	started, _ := server.Start(ctx)
	if !started {
		t.Fail()
	}
	stopped, _ := server.Stop(cancel, time.Millisecond*100)
	if !stopped {
		t.Fail()
	}
	t.Cleanup(func() {
		cancel()
	})
}

type MockHandler struct {
	readFunc    func(reader io.Reader) ([]byte, error)
	processFunc func(req []byte) []byte
	writeFunc   func(writer io.Writer, resp []byte) (int, error)
}

func (h MockHandler) Read(reader io.Reader) ([]byte, error) {
	return h.readFunc(reader)
}

func (h MockHandler) Process(req []byte) []byte {
	return h.processFunc(req)
}

func (h MockHandler) Write(writer io.Writer, resp []byte) (int, error) {
	return h.writeFunc(writer, resp)
}

func dial(serverAddr *net.TCPAddr, request []byte) ([]byte, error) {
	tcpConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return nil, err
	}
	defer tcpConn.Close()
	tcpConn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := tcpConn.Write(request); err != nil {
		return nil, err
	}
	if err := tcpConn.CloseWrite(); err != nil {
		return nil, err
	}
	return io.ReadAll(tcpConn)
}

func TestServerAcceptsRequests(t *testing.T) {
	request := []byte("Hello")
	response := []byte("World")

	config := config.Config{
		AddrPort:       netip.MustParseAddrPort("127.0.0.1:0"),
		MaxConnections: 1,
		ConnTimeout:    5 * time.Second,
	}

	received := make(chan []byte, 1)
	server := NewTCPServer(config, MockHandler{
		readFunc: func(reader io.Reader) ([]byte, error) {
			buf, err := io.ReadAll(reader)
			received <- buf
			return buf, err
		},
		processFunc: func(req []byte) []byte {
			return response
		},
		writeFunc: func(writer io.Writer, resp []byte) (int, error) {
			return writer.Write(resp)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	if ok, err := server.Start(ctx); !ok {
		t.Fatal("server start failed", "| error:", err)
	}
	t.Cleanup(func() {
		server.Stop(cancel, time.Second)
	})

	serverAddr := server.listener.Addr().(*net.TCPAddr)
	actualResponse, err := dial(serverAddr, request)
	if err != nil {
		t.Fatal("client dial failed", "| error:", err)
	}

	select {
	case actualRequest := <-received:
		logger.LOG.Info("actual request", "request", string(actualRequest))
		if !bytes.Equal(actualRequest, request) {
			t.Error("server read fail", "| actual:", string(actualRequest), "| expected:", string(request))
		}
	case <-time.After(time.Second):
		t.Error("server did not read request")
	}

	logger.LOG.Info("actual response", "response", string(actualResponse))
	if !bytes.Equal(actualResponse, response) {
		t.Error("client read fail", "| actual:", string(actualResponse), "| expected:", string(response))
	}

}
