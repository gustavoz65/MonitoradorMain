package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/native"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

// checkSite roda HTTP check e TLS check em goroutines paralelas e junta os resultados.
func checkSite(ctx context.Context, cfg shared.SiteConfig, client *http.Client) shared.SiteResult {
	cfg = cfg.Effective()
	var (
		mu     sync.Mutex
		result shared.SiteResult
		wg     sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		r := doHTTPCheck(ctx, cfg, client)
		mu.Lock()
		result = r
		mu.Unlock()
	}()

	if cfg.CheckCert {
		wg.Add(1)
		go func() {
			defer wg.Done()
			expiry, err := checkCert(cfg.URL)
			mu.Lock()
			if err == nil {
				result.CertExpiry = expiry
				if time.Until(expiry).Hours()/24 < float64(cfg.CertWarnDays) {
					result.CertWarn = true
				}
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if cfg.LatencyWarn > 0 && result.Latency > cfg.LatencyWarn {
		result.LatencyWarn = true
	}
	return result
}

func doHTTPCheck(ctx context.Context, cfg shared.SiteConfig, defaultClient *http.Client) shared.SiteResult {
	result := shared.SiteResult{Site: cfg.URL, CheckedAt: time.Now()}
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, nil)
	if err != nil {
		result.Error = err.Error()
		result.Latency = time.Since(start)
		return result
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	client := defaultClient
	if cfg.Timeout > 0 {
		client = &http.Client{Timeout: cfg.Timeout}
	}

	resp, err := client.Do(req)
	result.Latency = time.Since(start)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Online = matchStatus(resp.StatusCode, cfg.ExpectedStatus)

	if cfg.BodyMatch != "" {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr == nil {
			if !native.ContainsBytes(body, []byte(cfg.BodyMatch)) {
				result.Online = false
				result.Error = fmt.Sprintf("body nao contem: %s", cfg.BodyMatch)
			}
			result.ContentHash = native.HashBytes(body)
		}
	}
	return result
}

func matchStatus(code int, expected string) bool {
	if expected == "" {
		return code >= 200 && code < 400
	}
	expected = strings.TrimSpace(expected)
	if strings.Contains(expected, "-") {
		parts := strings.SplitN(expected, "-", 2)
		low, e1 := strconv.Atoi(parts[0])
		high, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil {
			return code >= 200 && code < 400
		}
		return code >= low && code <= high
	}
	if strings.HasSuffix(expected, "xx") {
		if n, err := strconv.Atoi(strings.TrimSuffix(expected, "xx")); err == nil {
			return code/100 == n
		}
	}
	if n, err := strconv.Atoi(expected); err == nil {
		return code == n
	}
	return code >= 200 && code < 400
}

func checkCert(rawURL string) (time.Time, error) {
	host := strings.TrimPrefix(rawURL, "https://")
	host = strings.TrimPrefix(host, "http://")
	if idx := strings.IndexByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}
	if !strings.Contains(host, ":") {
		host += ":443"
	}
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second}, "tcp", host,
		&tls.Config{InsecureSkipVerify: false},
	)
	if err != nil {
		return time.Time{}, err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, fmt.Errorf("sem certificados TLS")
	}
	return certs[0].NotAfter, nil
}
