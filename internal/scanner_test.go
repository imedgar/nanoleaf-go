package internal

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

type mockDialer struct {
	shouldConnect bool
	connectDelay  time.Duration
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !m.shouldConnect {
		return nil, fmt.Errorf("failed to connect")
	}

	// Simulate network delay
	time.Sleep(m.connectDelay)

	return &mockConn{},
		nil
}

type mockConn struct{}

func (c *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (c *mockConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (c *mockConn) Close() error                       { return nil }
func (c *mockConn) LocalAddr() net.Addr                { return nil }
func (c *mockConn) RemoteAddr() net.Addr               { return nil }
func (c *mockConn) SetDeadline(t time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNetworkScanner(t *testing.T) {
	t.Run("Scan discovers devices", func(t *testing.T) {
		scanner := NewNetworkScanner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		scanner.dialer = &mockDialer{shouldConnect: true}

		devices, err := scanner.Scan(ctx)
		if err != nil {
			t.Fatalf("failed to scan: %v", err)
		}

		if len(devices) == 0 {
			t.Errorf("expected to discover at least one device, but got none")
		}
	})

	t.Run("Scan timeout", func(t *testing.T) {
		scanner := NewNetworkScanner()
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		scanner.dialer = &mockDialer{shouldConnect: true, connectDelay: 100 * time.Millisecond}

		_, err := scanner.Scan(ctx)
		if err == nil {
			t.Fatal("expected timeout error, but got nil")
		}

		if ctx.Err() == nil {
			t.Fatal("expected timeout error, but got nil")
		}
	})

	t.Run("isNanoleafDevice", func(t *testing.T) {
		scanner := NewNetworkScanner()
		ctx := context.Background()

		scanner.dialer = &mockDialer{shouldConnect: true}
		if !scanner.isNanoleafDevice(ctx, "192.168.1.100") {
			t.Errorf("expected device to be a Nanoleaf device, but it wasn't")
		}

		scanner.dialer = &mockDialer{shouldConnect: false}
		if scanner.isNanoleafDevice(ctx, "192.168.1.101") {
			t.Errorf("expected device not to be a Nanoleaf device, but it was")
		}
	})
}

func TestNetworkScanner_getLocalSubnet(t *testing.T) {
	t.Run("getLocalSubnet", func(t *testing.T) {
		scanner := NewNetworkScanner()
		subnet, err := scanner.getLocalSubnet()
		if err != nil {
			t.Fatalf("failed to get local subnet: %v", err)
		}

		if subnet == "" {
			t.Errorf("expected a subnet, but got an empty string")
		}
	})
}
