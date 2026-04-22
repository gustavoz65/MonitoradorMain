package monitor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type Options struct {
	Workers int
	Timeout time.Duration
	Delay   time.Duration
	Hours   int
}

type Status struct {
	Running    bool
	SiteCount  int
	StartedAt  time.Time
	LastRunAt  time.Time
	CycleCount int
}

type Service struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	status Status
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CheckSitesOnce(ctx context.Context, sites []shared.SiteConfig, opts Options) []shared.SiteResult {
	if opts.Workers <= 0 {
		opts.Workers = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}

	jobs := make(chan shared.SiteConfig)
	results := make(chan shared.SiteResult, len(sites))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: opts.Timeout}

	for i := 0; i < opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cfg := range jobs {
				results <- checkSite(ctx, cfg, client)
			}
		}()
	}
	for _, site := range sites {
		jobs <- site
	}
	close(jobs)
	wg.Wait()
	close(results)

	var collected []shared.SiteResult
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

func (s *Service) Start(parent context.Context, sites []shared.SiteConfig, opts Options, onCycle func([]shared.SiteResult)) error {
	s.mu.Lock()
	if s.status.Running {
		s.mu.Unlock()
		return context.Canceled
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.status = Status{Running: true, StartedAt: time.Now(), SiteCount: len(sites)}
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.status.Running = false
			s.cancel = nil
			s.mu.Unlock()
		}()
		var deadline time.Time
		if opts.Hours > 0 {
			deadline = time.Now().Add(time.Duration(opts.Hours) * time.Hour)
		}
		for {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return
			}
			results := s.CheckSitesOnce(ctx, sites, opts)
			s.mu.Lock()
			s.status.LastRunAt = time.Now()
			s.status.CycleCount++
			s.mu.Unlock()
			onCycle(results)
			select {
			case <-ctx.Done():
				return
			case <-time.After(opts.Delay):
			}
		}
	}()
	return nil
}

func (s *Service) Stop() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel == nil {
		return false
	}
	s.cancel()
	return true
}

func (s *Service) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}
