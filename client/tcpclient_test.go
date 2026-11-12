package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"tcp-server/config"
	"testing"
	"time"
)

func startEchoServer(t *testing.T, reply func([]byte) []byte) netip.AddrPort {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		req, _ := io.ReadAll(conn)
		conn.Write(reply(req))
	}()
	return netip.MustParseAddrPort(listener.Addr().String())
}

func newClient(addr netip.AddrPort, maxResp uint, timeout time.Duration) *TCPClient {
	return NewTCPClient(config.ClientConfig{
		ServerAddrPort: addr,
		MaxRespBytes:   maxResp,
		DialTimeout:    time.Second,
		RequestTimeout: timeout,
	})
}

func TestSendReturnsResponse(t *testing.T) {
	addr := startEchoServer(t, func(req []byte) []byte {
		return append([]byte("echo:"), req...)
	})
	c := newClient(addr, 100, time.Second)
	resp, err := c.Send(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(resp, []byte("echo:hello")) {
		t.Fatalf("unexpected response: %q", resp)
	}
}

func TestSendResponseTooLarge(t *testing.T) {
	addr := startEchoServer(t, func(req []byte) []byte { return []byte("0123456789") })
	c := newClient(addr, 5, time.Second)
	_, err := c.Send(context.Background(), []byte("x"))
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestSendTimeout(t *testing.T) {
	addr := startEchoServer(t, func(req []byte) []byte {
		time.Sleep(time.Second)
		return []byte("late")
	})
	c := newClient(addr, 100, 100*time.Millisecond)
	_, err := c.Send(context.Background(), []byte("x"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestSendDialError(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := netip.MustParseAddrPort(listener.Addr().String())
	listener.Close()
	c := newClient(addr, 100, time.Second)
	if _, err := c.Send(context.Background(), []byte("x")); err == nil {
		t.Fatal("expected dial error")
	}
}
