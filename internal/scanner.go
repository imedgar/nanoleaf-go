package internal

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	nanoleafPort = "16021"
	maxWorkers   = 50
	scanTimeout  = 300 * time.Millisecond
)

// Dialer is an interface for dialing a network connection
 type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type NetworkScanner struct{
	dialer Dialer
}

// NewNetworkScanner creates a new network scanner
func NewNetworkScanner() *NetworkScanner {
	return &NetworkScanner{
		dialer: &net.Dialer{
			Timeout: scanTimeout,
		},
	}
}

// Scan discovers Nanoleaf devices on the local network
func (s *NetworkScanner) Scan(ctx context.Context) ([]string, error) {
	subnet, err := s.getLocalSubnet()
	if err != nil {
		return nil, fmt.Errorf("failed to detect local subnet: %w", err)
	}

	// Create job queue
	jobs := make(chan string, 254)
	results := make(chan string, 254)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, jobs, results, &wg)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i := 1; i <= 254; i++ {
			ip := fmt.Sprintf("%s%d", subnet, i)
			select {
			case jobs <- ip:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	devices := make([]string, 0)
	for {
		select {
		case ip, ok := <-results:
			if !ok {
				return devices, nil
			}
			if ip != "" {
				devices = append(devices, ip)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// worker performs the actual device scanning
func (s *NetworkScanner) worker(ctx context.Context, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case ip, ok := <-jobs:
			if !ok {
				return
			}

			if s.isNanoleafDevice(ctx, ip) {
				results <- ip
			} else {
				results <- ""
			}
		case <-ctx.Done():
			return
		}
	}
}

// isNanoleafDevice checks if a device at the given IP is a Nanoleaf device
func (s *NetworkScanner) isNanoleafDevice(ctx context.Context, ip string) bool {
	conn, err := s.dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, nanoleafPort))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// getLocalSubnet determines the local subnet for scanning
func (s *NetworkScanner) getLocalSubnet() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() != nil {

			parts := strings.Split(ipnet.IP.String(), ".")
			if len(parts) >= 3 {
				return fmt.Sprintf("%s.%s.%s.", parts[0], parts[1], parts[2]), nil
			}
		}
	}

	return "", fmt.Errorf("no suitable network interface found")
}
