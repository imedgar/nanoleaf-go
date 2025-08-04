package internal

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

func scanForDevices(ctx context.Context) ([]string, error) {
	// Get local IP to determine subnet
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var subnet string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				ip := ipNet.IP.To4()
				if ip[0] == 192 && ip[1] == 168 {
					subnet = fmt.Sprintf("192.168.%d", ip[2])
					break
				}
			}
		}
		if subnet != "" {
			break
		}
	}

	if subnet == "" {
		return nil, fmt.Errorf("no suitable network interface found")
	}

	// Scan the subnet for Nanoleaf devices (port 16021)
	var devices []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 1; i < 255; i++ {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", ip+":16021", 100*time.Millisecond)
			if err == nil {
				conn.Close()
				mu.Lock()
				devices = append(devices, ip)
				mu.Unlock()
			}
		}(fmt.Sprintf("%s.%d", subnet, i))
	}

	// Wait for all scans to complete or context cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return devices, nil
	}
}
