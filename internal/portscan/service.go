package portscan

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

var DefaultPorts = []int{21, 22, 25, 53, 80, 110, 143, 443, 3306, 5432, 6379, 8080}

type Options struct {
	Ports   []int
	Workers int
	Timeout time.Duration
}

func ParsePorts(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return append([]int{}, DefaultPorts...), nil
	}
	set := map[int]struct{}{}
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("faixa invalida: %s", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, err
			}
			for port := start; port <= end; port++ {
				set[port] = struct{}{}
			}
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		set[port] = struct{}{}
	}
	var ports []int
	for port := range set {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	return ports, nil
}

func Scan(ctx context.Context, host string, opts Options) []shared.PortResult {
	if len(opts.Ports) == 0 {
		opts.Ports = append([]int{}, DefaultPorts...)
	}
	if opts.Workers <= 0 {
		opts.Workers = 20
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 800 * time.Millisecond
	}

	jobs := make(chan int)
	results := make(chan shared.PortResult, len(opts.Ports))
	var wg sync.WaitGroup

	for i := 0; i < opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				start := time.Now()
				address := net.JoinHostPort(host, strconv.Itoa(port))
				conn, err := (&net.Dialer{Timeout: opts.Timeout}).DialContext(ctx, "tcp", address)
				if err != nil {
					results <- shared.PortResult{Host: host, Port: port, Open: false, Error: err.Error(), CheckedAt: time.Now(), Latency: time.Since(start)}
					continue
				}
				_ = conn.Close()
				results <- shared.PortResult{Host: host, Port: port, Open: true, CheckedAt: time.Now(), Latency: time.Since(start)}
			}
		}()
	}

	for _, port := range opts.Ports {
		jobs <- port
	}
	close(jobs)
	wg.Wait()
	close(results)

	collected := make([]shared.PortResult, 0, len(opts.Ports))
	for result := range results {
		collected = append(collected, result)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].Port < collected[j].Port })
	return collected
}
