# MoniMaster v3.0.0 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Evoluir o MoniMaster para v3.0.0 com checks HTTP avançados (cgo + goroutines), TUI reativa Charm, e distribuição como binário único via GoReleaser.

**Architecture:** Três fases sequenciais: B (monitoramento avançado), A (Charm TUI), C (distribuição). Toda lógica de domínio permanece em `internal/`; a TUI é camada de apresentação pura que implementa `AppFacade`.

**Tech Stack:** Go 1.23, cgo (internal/native), bubbletea + lipgloss + bubbles, GoReleaser, GitHub Actions.

---

## PHASE B — Monitoramento Avançado

### Task 1: Expandir internal/native com ContainsBytes e HashBytes

**Files:**
- Modify: `internal/native/native.go`
- Modify: `internal/native/native_cgo.go`
- Modify: `internal/native/native_stub.go`
- Create: `internal/native/native_test.go`

**Step 1: Escrever teste que falha**

```go
// internal/native/native_test.go
package native

import "testing"

func TestContainsBytes(t *testing.T) {
	if !ContainsBytes([]byte("hello world"), []byte("world")) {
		t.Fatal("expected true")
	}
	if ContainsBytes([]byte("hello"), []byte("world")) {
		t.Fatal("expected false")
	}
	if !ContainsBytes([]byte("abc"), []byte("")) {
		t.Fatal("empty pattern should match")
	}
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("monimaster"))
	h2 := HashBytes([]byte("monimaster"))
	h3 := HashBytes([]byte("other"))
	if h1 != h2 {
		t.Fatal("same input must produce same hash")
	}
	if h1 == h3 {
		t.Fatal("different inputs should produce different hashes")
	}
}
```

**Step 2: Rodar para confirmar falha**

```
go test ./internal/native/...
```
Esperado: `undefined: ContainsBytes`, `undefined: HashBytes`

**Step 3: Atualizar native.go**

```go
package native

func NormalizeASCII(value string) string  { return normalizeASCII(value) }
func ContainsBytes(body, pattern []byte) bool { return containsBytes(body, pattern) }
func HashBytes(data []byte) uint32            { return hashBytes(data) }
```

**Step 4: Atualizar native_cgo.go**

```go
//go:build cgo

package native

/*
#include <ctype.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

static void monimaster_ascii_lower(char *s) {
	while (*s) { *s = (char)tolower((unsigned char)*s); s++; }
}

static int monimaster_contains(const char *hay, size_t hlen, const char *needle, size_t nlen) {
	if (nlen == 0) return 1;
	if (hlen < nlen) return 0;
	for (size_t i = 0; i <= hlen - nlen; i++) {
		if (memcmp(hay + i, needle, nlen) == 0) return 1;
	}
	return 0;
}

static uint32_t monimaster_crc32(const uint8_t *data, size_t len) {
	uint32_t crc = 0xFFFFFFFF;
	for (size_t i = 0; i < len; i++) {
		crc ^= data[i];
		for (int j = 0; j < 8; j++) {
			crc = (crc >> 1) ^ (0xEDB88320u & (uint32_t)(-(int32_t)(crc & 1)));
		}
	}
	return ~crc;
}
*/
import "C"
import "unsafe"

func normalizeASCII(value string) string {
	cstr := C.CString(value)
	defer C.free(unsafe.Pointer(cstr))
	C.monimaster_ascii_lower(cstr)
	return C.GoString(cstr)
}

func containsBytes(body, pattern []byte) bool {
	if len(pattern) == 0 { return true }
	if len(body) == 0 { return false }
	return C.monimaster_contains(
		(*C.char)(unsafe.Pointer(&body[0])), C.size_t(len(body)),
		(*C.char)(unsafe.Pointer(&pattern[0])), C.size_t(len(pattern)),
	) == 1
}

func hashBytes(data []byte) uint32 {
	if len(data) == 0 { return 0 }
	return uint32(C.monimaster_crc32((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data))))
}
```

**Step 5: Atualizar native_stub.go**

```go
//go:build !cgo

package native

import (
	"bytes"
	"hash/crc32"
	"strings"
)

func normalizeASCII(value string) string       { return strings.ToLower(value) }
func containsBytes(body, pattern []byte) bool  { return bytes.Contains(body, pattern) }
func hashBytes(data []byte) uint32             { return crc32.ChecksumIEEE(data) }
```

**Step 6: Rodar testes**

```
go test ./internal/native/...
```
Esperado: PASS

**Step 7: Commit**

```bash
git add internal/native/
git commit -m "feat(native): ContainsBytes e HashBytes via cgo com fallback Go puro"
```

---

### Task 2: Adicionar SiteConfig e expandir SiteResult em shared/types.go

**Files:**
- Modify: `internal/shared/types.go`
- Modify: `internal/shared/types_test.go`

**Step 1: Escrever testes**

```go
// adicionar em internal/shared/types_test.go
func TestSiteConfigEffective(t *testing.T) {
	cfg := SiteConfig{URL: "https://example.com"}.Effective()
	if cfg.Method != "GET" { t.Fatal("default method should be GET") }
	if cfg.CertWarnDays != 30 { t.Fatal("default cert warn days should be 30") }
}

func TestSiteConfigFromURL(t *testing.T) {
	cfg := SiteConfigFromURL("https://example.com")
	if cfg.URL != "https://example.com" { t.Fatalf("url mismatch: %s", cfg.URL) }
}
```

**Step 2: Rodar para confirmar falha**

```
go test ./internal/shared/...
```

**Step 3: Adicionar em types.go (após SiteResult)**

```go
type SiteConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	ExpectedStatus string            `json:"expected_status,omitempty"`
	BodyMatch      string            `json:"body_match,omitempty"`
	CheckCert      bool              `json:"check_cert,omitempty"`
	CertWarnDays   int               `json:"cert_warn_days,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	LatencyWarn    time.Duration     `json:"latency_warn,omitempty"`
}

func (s SiteConfig) Effective() SiteConfig {
	if s.Method == ""     { s.Method = "GET" }
	if s.CertWarnDays == 0 { s.CertWarnDays = 30 }
	return s
}

func SiteConfigFromURL(url string) SiteConfig { return SiteConfig{URL: url} }
```

Atualizar `SiteResult` — adicionar campos:

```go
type SiteResult struct {
	Site        string        `json:"site"`
	Online      bool          `json:"online"`
	StatusCode  int           `json:"status_code"`
	Latency     time.Duration `json:"latency"`
	Error       string        `json:"error,omitempty"`
	CheckedAt   time.Time     `json:"checked_at"`
	ContentHash uint32        `json:"content_hash,omitempty"`
	CertExpiry  time.Time     `json:"cert_expiry,omitempty"`
	CertWarn    bool          `json:"cert_warn,omitempty"`
	LatencyWarn bool          `json:"latency_warn,omitempty"`
}
```

**Step 4: Rodar testes**

```
go test ./internal/shared/...
```

**Step 5: Commit**

```bash
git add internal/shared/
git commit -m "feat(shared): SiteConfig com checks customizaveis e SiteResult expandido"
```

---

### Task 3: Criar internal/monitor/checker.go

**Files:**
- Create: `internal/monitor/checker.go`
- Create: `internal/monitor/checker_test.go`

**Step 1: Escrever teste de matchStatus**

```go
// internal/monitor/checker_test.go
package monitor

import "testing"

func TestMatchStatus(t *testing.T) {
	cases := []struct {
		code     int
		expected string
		want     bool
	}{
		{200, "", true},
		{404, "", false},
		{201, "201", true},
		{200, "201", false},
		{201, "200-299", true},
		{301, "200-299", false},
		{201, "2xx", true},
		{301, "2xx", false},
		{301, "3xx", true},
	}
	for _, c := range cases {
		if got := matchStatus(c.code, c.expected); got != c.want {
			t.Errorf("matchStatus(%d, %q) = %v, want %v", c.code, c.expected, got, c.want)
		}
	}
}
```

**Step 2: Rodar para confirmar falha**

```
go test ./internal/monitor/...
```

**Step 3: Criar checker.go**

```go
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
```

**Step 4: Rodar testes**

```
go test ./internal/monitor/...
```

**Step 5: Commit**

```bash
git add internal/monitor/checker.go internal/monitor/checker_test.go
git commit -m "feat(monitor): checker com TLS paralelo, body match via cgo, matchStatus"
```

---

### Task 4: Atualizar monitor/service.go para usar []SiteConfig

**Files:**
- Modify: `internal/monitor/service.go`

**Step 1: Substituir []string por []shared.SiteConfig nas assinaturas**

Em `CheckSitesOnce`, trocar `sites []string` por `sites []shared.SiteConfig` e usar `checkSite` no worker:

```go
func (s *Service) CheckSitesOnce(ctx context.Context, sites []shared.SiteConfig, opts Options) []shared.SiteResult {
	if opts.Workers <= 0 { opts.Workers = 4 }
	if opts.Timeout <= 0 { opts.Timeout = 5 * time.Second }

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
	for r := range results {
		collected = append(collected, r)
	}
	return collected
}
```

Fazer o mesmo em `Start` — trocar `[]string` por `[]shared.SiteConfig`.

Remover import de `"net/http"` do service.go (agora está no checker.go). Manter apenas `sync`, `context`, `time`.

**Step 2: Rodar build (vai falhar em app.go — corrigido na Task 8)**

```
go build ./internal/monitor/...
```

**Step 3: Commit**

```bash
git add internal/monitor/service.go
git commit -m "feat(monitor): service aceita []SiteConfig com worker pool concorrente"
```

---

### Task 5: Sistema de notificação plugável

**Files:**
- Create: `internal/notify/provider.go`
- Create: `internal/notify/smtp.go`
- Create: `internal/notify/resend.go`
- Modify: `internal/notify/service.go`

**Step 1: Criar provider.go**

```go
package notify

import "github.com/gustavoz65/MoniMaster/internal/config"

type Provider interface {
	Name() string
	Send(cfg config.AppConfig, to, subject, body string) error
}
```

**Step 2: Criar smtp.go** (extraído de service.go)

```go
package notify

import (
	"fmt"
	"strings"

	"github.com/gustavoz65/MoniMaster/internal/config"
	"gopkg.in/gomail.v2"
)

type SMTPProvider struct{}

func (p *SMTPProvider) Name() string { return "smtp" }

func (p *SMTPProvider) Send(cfg config.AppConfig, to, subject, body string) error {
	c := cfg.SMTP
	if strings.TrimSpace(c.Host) == "" || c.Port == 0 || strings.TrimSpace(c.User) == "" || strings.TrimSpace(c.Password) == "" {
		return fmt.Errorf("smtp incompleto; use config smtp set")
	}
	from := c.From
	if strings.TrimSpace(from) == "" {
		from = c.User
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)
	return gomail.NewDialer(c.Host, c.Port, c.User, c.Password).DialAndSend(msg)
}
```

**Step 3: Criar resend.go**

```go
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gustavoz65/MoniMaster/internal/config"
)

type ResendProvider struct{}

func (p *ResendProvider) Name() string { return "resend" }

func (p *ResendProvider) Send(cfg config.AppConfig, to, subject, body string) error {
	apiKey := strings.TrimSpace(cfg.Notify.APIKey)
	from := strings.TrimSpace(cfg.Notify.From)
	if apiKey == "" {
		return fmt.Errorf("resend: api key nao configurada; use config notify provider set resend --api-key xxx")
	}
	if from == "" {
		return fmt.Errorf("resend: remetente nao configurado; use config notify provider set resend --from noreply@dominio.com")
	}
	payload, _ := json.Marshal(map[string]any{
		"from": from, "to": []string{to}, "subject": subject, "text": body,
	})
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend retornou status %d", resp.StatusCode)
	}
	return nil
}
```

**Step 4: Reescrever service.go com dispatch async via goroutine**

```go
package notify

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/config"
	"github.com/gustavoz65/MoniMaster/internal/shared"
	"github.com/gustavoz65/MoniMaster/internal/storage"
)

type alertJob struct {
	cfg     config.AppConfig
	to      string
	subject string
	body    string
}

type Service struct {
	store    storage.Store
	provider Provider
	alerts   chan alertJob
}

func NewService(store storage.Store) *Service {
	s := &Service{
		store:    store,
		provider: &SMTPProvider{},
		alerts:   make(chan alertJob, 64),
	}
	go s.dispatchLoop()
	return s
}

func (s *Service) SetProvider(p Provider) { s.provider = p }

func (s *Service) dispatchLoop() {
	for job := range s.alerts {
		_ = s.provider.Send(job.cfg, job.to, job.subject, job.body)
	}
}

// Send enfileira alerta de forma nao-bloqueante — nunca trava o ciclo do monitor.
func (s *Service) Send(cfg config.AppConfig, to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("destinatario nao configurado")
	}
	select {
	case s.alerts <- alertJob{cfg, to, subject, body}:
		return nil
	default:
		return fmt.Errorf("fila de alertas cheia")
	}
}

// SendSync envia de forma sincrona (usado em notify email test).
func (s *Service) SendSync(cfg config.AppConfig, to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("destinatario nao configurado")
	}
	return s.provider.Send(cfg, to, subject, body)
}

func (s *Service) ResolveTarget(cfg config.AppConfig, identity *shared.Identity) string {
	if identity != nil && s.store.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		email, err := s.store.GetNotificationEmail(ctx, identity.ID)
		if err == nil && strings.TrimSpace(email) != "" {
			return email
		}
	}
	return cfg.Notification.DefaultEmail
}

func (s *Service) SetTarget(cfg *config.AppConfig, identity *shared.Identity, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email vazio")
	}
	if identity != nil && s.store.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.store.SaveNotificationEmail(ctx, identity.ID, email); err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	cfg.Notification.DefaultEmail = email
	return nil
}
```

**Step 5: Rodar build**

```
go build ./internal/notify/...
```

**Step 6: Commit**

```bash
git add internal/notify/
git commit -m "feat(notify): sistema plugavel SMTP/Resend com dispatch async via goroutine"
```

---

### Task 6: Atualizar config/types.go com AlertConfig e NotifyConfig

**Files:**
- Modify: `internal/config/types.go`

**Step 1: Adicionar ao final do arquivo**

```go
type AlertConfig struct {
	LatencyWarn  string `json:"latency_warn"`   // ex: "500ms"
	LatencyCrit  string `json:"latency_crit"`   // ex: "2s"
	CertWarnDays int    `json:"cert_warn_days"` // default 30
}

type NotifyConfig struct {
	Provider string `json:"provider"` // "smtp" ou "resend"
	APIKey   string `json:"api_key"`
	From     string `json:"from"`
}
```

**Step 2: Adicionar campos em AppConfig**

```go
type AppConfig struct {
	Storage      StorageConfig      `json:"storage"`
	SMTP         SMTPConfig         `json:"smtp"`
	Monitor      MonitorConfig      `json:"monitor"`
	Notification NotificationConfig `json:"notification"`
	Alert        AlertConfig        `json:"alert"`
	Notify       NotifyConfig       `json:"notify"`
}
```

**Step 3: Atualizar Default()**

```go
func Default() AppConfig {
	return AppConfig{
		Storage: StorageConfig{Enabled: false},
		SMTP:    SMTPConfig{Port: 587},
		Monitor: MonitorConfig{
			DelaySeconds:    5,
			TimeoutSeconds:  5,
			WorkerCount:     6,
			CleanupInterval: "7d",
		},
		Notification: NotificationConfig{},
		Alert: AlertConfig{
			LatencyWarn:  "500ms",
			LatencyCrit:  "2s",
			CertWarnDays: 30,
		},
		Notify: NotifyConfig{Provider: "smtp"},
	}
}
```

**Step 4: Commit**

```bash
git add internal/config/types.go
git commit -m "feat(config): AlertConfig e NotifyConfig para thresholds e provider plugavel"
```

---

### Task 7: Atualizar config/files.go com LoadSiteConfigs e SaveSiteConfigs

**Files:**
- Modify: `internal/config/files.go`
- Create: `internal/config/files_test.go`

**Step 1: Escrever testes**

```go
// internal/config/files_test.go
package config

import (
	"os"
	"testing"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

func TestLoadSiteConfigsMigration(t *testing.T) {
	dir := t.TempDir()
	mgr := &Manager{homeDir: dir}

	if err := os.WriteFile(mgr.SitesPath(), []byte(`["https://a.com","https://b.com"]`), 0o644); err != nil {
		t.Fatal(err)
	}
	configs, err := mgr.LoadSiteConfigs()
	if err != nil { t.Fatal(err) }
	if len(configs) != 2 { t.Fatalf("got %d, want 2", len(configs)) }
	if configs[0].URL != "https://a.com" { t.Fatalf("got %s", configs[0].URL) }
}

func TestSaveAndLoadSiteConfigs(t *testing.T) {
	dir := t.TempDir()
	mgr := &Manager{homeDir: dir}

	in := []shared.SiteConfig{{URL: "https://example.com", Method: "GET", CheckCert: true}}
	if err := mgr.SaveSiteConfigs(in); err != nil { t.Fatal(err) }

	out, err := mgr.LoadSiteConfigs()
	if err != nil { t.Fatal(err) }
	if len(out) != 1 || out[0].URL != "https://example.com" { t.Fatal("round-trip falhou") }
	if !out[0].CheckCert { t.Fatal("check_cert nao persistido") }
}
```

**Step 2: Rodar para confirmar falha**

```
go test ./internal/config/...
```

**Step 3: Adicionar em files.go**

Adicionar imports `"fmt"` e `"github.com/gustavoz65/MoniMaster/internal/shared"`. Adicionar os métodos após `SaveSites`:

```go
func (m *Manager) LoadSiteConfigs() ([]shared.SiteConfig, error) {
	data, err := os.ReadFile(m.SitesPath())
	if err != nil {
		if os.IsNotExist(err) {
			_ = m.SaveSiteConfigs([]shared.SiteConfig{})
			return []shared.SiteConfig{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []shared.SiteConfig{}, nil
	}
	// Tentar formato novo primeiro
	var configs []shared.SiteConfig
	if err := json.Unmarshal(data, &configs); err == nil && len(configs) > 0 && configs[0].URL != "" {
		return configs, nil
	}
	// Migrar de []string (formato antigo)
	var urls []string
	if err := json.Unmarshal(data, &urls); err != nil {
		return nil, fmt.Errorf("formato de sites desconhecido: %w", err)
	}
	configs = make([]shared.SiteConfig, len(urls))
	for i, u := range urls {
		configs[i] = shared.SiteConfigFromURL(u)
	}
	_ = m.SaveSiteConfigs(configs)
	return configs, nil
}

func (m *Manager) SaveSiteConfigs(sites []shared.SiteConfig) error {
	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.SitesPath(), data, 0o644)
}
```

**Step 4: Rodar testes**

```
go test ./internal/config/...
```

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): LoadSiteConfigs/SaveSiteConfigs com migracao automatica de []string"
```

---

### Task 8: Atualizar app.go para novos comandos e SiteConfig

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Trocar LoadSites por LoadSiteConfigs em handleSites e handleMonitor**

Em `handleSites`: trocar `a.manager.LoadSites()` por `a.manager.LoadSiteConfigs()` e `a.manager.SaveSites(...)` por `a.manager.SaveSiteConfigs(...)`.

Em `handleMonitor`: trocar `a.manager.LoadSites()` por `a.manager.LoadSiteConfigs()`.

**Step 2: Atualizar case "list" em handleSites**

```go
case "list":
	if len(sites) == 0 {
		fmt.Println("Nenhum site configurado.")
		return nil
	}
	for i, site := range sites {
		extra := ""
		if site.CheckCert       { extra += " [cert]" }
		if site.BodyMatch != "" { extra += " [body]" }
		if site.Method != "" && site.Method != "GET" { extra += " [" + site.Method + "]" }
		fmt.Printf("%d. %s%s\n", i+1, site.URL, extra)
	}
```

**Step 3: Atualizar case "add" em handleSites**

```go
case "add":
	if len(cmd.Args) == 0 {
		return fmt.Errorf("use sites add <url> [--method GET] [--expected-status 200] [--body-match texto] [--check-cert]")
	}
	url := cmd.Args[0]
	for _, existing := range sites {
		if existing.URL == url { return fmt.Errorf("site ja existe") }
	}
	cfg := shared.SiteConfigFromURL(url)
	if v := cmd.Flags["method"]; v != ""           { cfg.Method = strings.ToUpper(v) }
	if v := cmd.Flags["expected-status"]; v != ""  { cfg.ExpectedStatus = v }
	if v := cmd.Flags["body-match"]; v != ""       { cfg.BodyMatch = v }
	if cmd.BoolFlags["check-cert"]                  { cfg.CheckCert = true }
	if v := cmd.Flags["cert-warn-days"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil { cfg.CertWarnDays = n }
	}
	sites = append(sites, cfg)
	sort.Slice(sites, func(i, j int) bool { return sites[i].URL < sites[j].URL })
	if err := a.manager.SaveSiteConfigs(sites); err != nil { return err }
	fmt.Println("Site adicionado.")
```

**Step 4: Adicionar case "update" em handleSites (após "remove")**

```go
case "update":
	if len(cmd.Args) == 0 { return fmt.Errorf("use sites update <url> [flags]") }
	target := cmd.Args[0]
	updated := false
	for i, s := range sites {
		if s.URL != target { continue }
		if v := cmd.Flags["method"]; v != ""           { sites[i].Method = strings.ToUpper(v) }
		if v := cmd.Flags["expected-status"]; v != ""  { sites[i].ExpectedStatus = v }
		if v := cmd.Flags["body-match"]; v != ""       { sites[i].BodyMatch = v }
		if cmd.BoolFlags["check-cert"]                  { sites[i].CheckCert = true }
		if cmd.BoolFlags["no-check-cert"]               { sites[i].CheckCert = false }
		updated = true
		break
	}
	if !updated { return fmt.Errorf("site nao encontrado: %s", target) }
	if err := a.manager.SaveSiteConfigs(sites); err != nil { return err }
	fmt.Println("Site atualizado.")
```

**Step 5: Adicionar monitor alert no handleMonitor**

```go
case "alert":
	return a.handleMonitorAlert(cmd)
```

Criar método:

```go
func (a *App) handleMonitorAlert(cmd cli.Command) error {
	if firstArg(cmd.Args) != "set" {
		return fmt.Errorf("use monitor alert set [--latency-warn 500ms] [--latency-crit 2s] [--cert-warn-days 30]")
	}
	if v := cmd.Flags["latency-warn"]; v != ""  { a.cfg.Alert.LatencyWarn = v }
	if v := cmd.Flags["latency-crit"]; v != ""  { a.cfg.Alert.LatencyCrit = v }
	if v := cmd.Flags["cert-warn-days"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil { a.cfg.Alert.CertWarnDays = n }
	}
	if err := a.manager.Save(a.cfg); err != nil { return err }
	fmt.Println("Thresholds de alerta atualizados.")
	return nil
}
```

**Step 6: Adicionar config notify no handleConfig**

```go
case "notify":
	return a.handleConfigNotify(cmd)
```

Criar método:

```go
func (a *App) handleConfigNotify(cmd cli.Command) error {
	if len(cmd.Path) < 3 || cmd.Path[2] != "provider" || firstArg(cmd.Args) != "set" {
		return fmt.Errorf("use config notify provider set smtp|resend [--api-key xxx] [--from email]")
	}
	if len(cmd.Args) < 2 { return fmt.Errorf("especifique o provider: smtp ou resend") }
	switch strings.ToLower(cmd.Args[1]) {
	case "smtp":
		a.cfg.Notify.Provider = "smtp"
		a.notify.SetProvider(&notify.SMTPProvider{})
	case "resend":
		a.cfg.Notify.Provider = "resend"
		if v := cmd.Flags["api-key"]; v != "" { a.cfg.Notify.APIKey = v }
		if v := cmd.Flags["from"]; v != ""    { a.cfg.Notify.From = v }
		a.notify.SetProvider(&notify.ResendProvider{})
	default:
		return fmt.Errorf("provider desconhecido; use smtp ou resend")
	}
	if err := a.manager.Save(a.cfg); err != nil { return err }
	fmt.Printf("Provider de notificacao: %s\n", a.cfg.Notify.Provider)
	return nil
}
```

**Step 7: Trocar notify.Send por notify.SendSync no case "test" de handleNotify**

```go
case "test":
	target := a.notify.ResolveTarget(a.cfg, a.session.Identity)
	if err := a.notify.SendSync(a.cfg, target, "Teste MoniMaster", "Seu canal de notificacao esta configurado."); err != nil {
		return err
	}
	fmt.Println("Email de teste enviado para", target)
```

**Step 8: Rodar build completo**

```
go build ./...
```

**Step 9: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): sites update, monitor alert set, config notify provider, SiteConfig"
```

---

### Task 9: Atualizar cli/help.go

**Files:**
- Modify: `internal/cli/help.go`

**Step 1: Substituir o conteúdo de HelpText**

```go
const HelpText = `
Comandos principais:
  help
  profile
  version
  exit

Autenticacao:
  auth login
  auth register
  auth logout

Configuracao:
  config show
  config wizard
  config db set --driver postgres --dsn "postgres://..."
  config db disable
  config smtp set --host smtp.example.com --port 587 --user me --password secret --from monitor@example.com
  config notify provider set smtp
  config notify provider set resend --api-key re_xxx --from noreply@seudominio.com

Diagnostico:
  doctor run

Sites:
  sites list
  sites add https://example.com [--method GET] [--expected-status 200-299] [--body-match "pong"] [--check-cert] [--cert-warn-days 14]
  sites update https://example.com [--expected-status 200] [--check-cert] [--no-check-cert]
  sites remove https://example.com
  sites import --file sites.txt

Monitoramento:
  monitor once
  monitor start [--hours 2]
  monitor status
  monitor stop
  monitor alert set [--latency-warn 500ms] [--latency-crit 2s] [--cert-warn-days 30]
  monitor dashboard

Logs e relatorios:
  logs show
  logs clear
  logs export [--format json] [--output caminho/base]
  history show [--limit 20]
  report uptime
  report ports

Notificacoes:
  notify email set user@example.com
  notify email test
  cleanup interval set 7d

Port scan:
  portscan run --host example.com [--ports 22,80,443,8000-8010] [--timeout 800ms]
`
```

**Step 2: Commit**

```bash
git add internal/cli/help.go
git commit -m "docs(cli): novos comandos na ajuda (sites update, monitor alert, config notify, dashboard)"
```

---

## PHASE A — Charm TUI

### Task 10: Instalar dependências Charm e criar internal/tui/styles.go

**Files:**
- Modify: `go.mod`
- Create: `internal/tui/styles.go`

**Step 1: Instalar dependências**

```
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
```

**Step 2: Criar styles.go**

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleOnline   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	StyleOffline  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	StyleWarn     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	StyleTitle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	StylePrompt   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	StyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	StyleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleSuccess  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleHeader   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Underline(true)
)

const (
	IconOnline  = "●"
	IconOffline = "✗"
	IconWarn    = "⚠"
)
```

**Step 3: Commit**

```bash
git add go.mod go.sum internal/tui/styles.go
git commit -m "feat(tui): dependencias Charm e paleta de estilos centralizada"
```

---

### Task 11: Criar internal/tui/table.go e progress.go

**Files:**
- Create: `internal/tui/table.go`
- Create: `internal/tui/progress.go`

**Step 1: Criar table.go**

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type TableRow []string

type Table struct {
	Headers []string
	Rows    []TableRow
}

func (t Table) Render() string {
	if len(t.Headers) == 0 { return "" }
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers { widths[i] = len(h) }
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] { widths[i] = len(cell) }
		}
	}
	headerCells := make([]string, len(t.Headers))
	for i, h := range t.Headers { headerCells[i] = StyleHeader.Render(pad(h, widths[i])) }

	sep := make([]string, len(t.Headers))
	for i, w := range widths { sep[i] = strings.Repeat("─", w) }

	var sb strings.Builder
	sb.WriteString(strings.Join(headerCells, "  ") + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Join(sep, "  ")) + "\n")
	for _, row := range t.Rows {
		cells := make([]string, len(t.Headers))
		for i := range t.Headers {
			cell := ""
			if i < len(row) { cell = row[i] }
			cells[i] = pad(cell, widths[i])
		}
		sb.WriteString(strings.Join(cells, "  ") + "\n")
	}
	return sb.String()
}

func pad(s string, width int) string {
	if len(s) >= width { return s }
	return s + strings.Repeat(" ", width-len(s))
}
```

**Step 2: Criar progress.go**

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressMsg struct{ Done, Total int }

type ProgressModel struct {
	bar   progress.Model
	done  int
	total int
	label string
}

func NewProgressModel(label string, total int) ProgressModel {
	return ProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: label, total: total}
}

func (m ProgressModel) Init() tea.Cmd { return nil }

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		m.done, m.total = msg.Done, msg.Total
		pct := 0.0
		if m.total > 0 { pct = float64(m.done) / float64(m.total) }
		return m, m.bar.SetPercent(pct)
	case progress.FrameMsg:
		bar, cmd := m.bar.Update(msg)
		m.bar = bar.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m ProgressModel) View() string {
	return fmt.Sprintf("%s\n%s  %d/%d\n", m.label, m.bar.View(), m.done, m.total)
}
```

**Step 3: Commit**

```bash
git add internal/tui/table.go internal/tui/progress.go
git commit -m "feat(tui): componentes de tabela e barra de progresso"
```

---

### Task 12: Criar internal/tui/facade.go

**Files:**
- Create: `internal/tui/facade.go`

**Step 1: Implementar**

```go
package tui

import (
	"github.com/gustavoz65/MoniMaster/internal/cli"
	"github.com/gustavoz65/MoniMaster/internal/monitor"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type EntryResult struct {
	Proceed  bool
	Identity *shared.Identity
	Mode     string
}

// AppFacade desacopla os modelos TUI da implementacao concreta de App.
type AppFacade interface {
	Login(username, password string) (*shared.Identity, error)
	Register(username, email, password string) (*shared.Identity, error)
	ConfigDB(driver, dsn string) error
	SetupWizard(useDB bool, driver, dsn, defaultEmail string) error
	Execute(cmd cli.Command) (exit bool, output string, err error)
	MonitorStatus() monitor.Status
	SubscribeResults() <-chan []shared.SiteResult
}
```

**Step 2: Commit**

```bash
git add internal/tui/facade.go
git commit -m "feat(tui): interface AppFacade para desacoplamento TUI/app"
```

---

### Task 13: Criar internal/tui/entry.go

**Files:**
- Create: `internal/tui/entry.go`

**Step 1: Implementar**

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type entryState int

const (
	stateMenu     entryState = iota
	stateLogin
	stateRegister
	stateConfigDB
)

var menuItems = []string{"Login", "Cadastro", "Continuar anônimo", "Configurar banco", "Assistente inicial", "Sair"}

type loginDoneMsg    struct{ identity *shared.Identity; err error }
type registerDoneMsg struct{ identity *shared.Identity; err error }
type configDBDoneMsg struct{ err error }

type EntryModel struct {
	app     AppFacade
	state   entryState
	cursor  int
	inputs  []textinput.Model
	inputIdx int
	err     error
	Result  EntryResult
	Done    bool
}

func NewEntryModel(app AppFacade) EntryModel { return EntryModel{app: app} }

func (m EntryModel) Init() tea.Cmd { return nil }

func (m EntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == stateMenu {
			return m.updateMenu(msg)
		}
		return m.updateForm(msg)
	case loginDoneMsg:
		if msg.err != nil { m.err = msg.err; m.state = stateMenu; return m, nil }
		m.Result = EntryResult{Proceed: true, Identity: msg.identity, Mode: shared.SessionModeAuthenticated}
		m.Done = true
		return m, tea.Quit
	case registerDoneMsg:
		if msg.err != nil { m.err = msg.err; m.state = stateMenu; return m, nil }
		m.Result = EntryResult{Proceed: true, Identity: msg.identity, Mode: shared.SessionModeAuthenticated}
		m.Done = true
		return m, tea.Quit
	case configDBDoneMsg:
		m.err = msg.err; m.state = stateMenu; return m, nil
	}
	return m, nil
}

func (m EntryModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 { m.cursor-- }
	case "down", "j":
		if m.cursor < len(menuItems)-1 { m.cursor++ }
	case "enter":
		m.err = nil
		switch m.cursor {
		case 0:
			m.state = stateLogin
			m.inputs = makeInputs([]inputSpec{{"Usuário", textinput.EchoNormal}, {"Senha", textinput.EchoPassword}})
			m.inputIdx = 0
		case 1:
			m.state = stateRegister
			m.inputs = makeInputs([]inputSpec{{"Usuário", textinput.EchoNormal}, {"Email", textinput.EchoNormal}, {"Senha", textinput.EchoPassword}})
			m.inputIdx = 0
		case 2:
			m.Result = EntryResult{Proceed: true, Mode: shared.SessionModeAnonymous}
			m.Done = true
			return m, tea.Quit
		case 3, 4:
			m.state = stateConfigDB
			m.inputs = makeInputs([]inputSpec{{"Driver (postgres/mysql/sqlite/oracle)", textinput.EchoNormal}, {"DSN", textinput.EchoNormal}})
			m.inputIdx = 0
		case 5:
			m.Result = EntryResult{Proceed: false}
			m.Done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m EntryModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc":
		m.state = stateMenu; return m, nil
	case "tab", "down":
		if m.inputIdx < len(m.inputs)-1 {
			m.inputs[m.inputIdx].Blur(); m.inputIdx++; m.inputs[m.inputIdx].Focus()
		}
	case "shift+tab", "up":
		if m.inputIdx > 0 {
			m.inputs[m.inputIdx].Blur(); m.inputIdx--; m.inputs[m.inputIdx].Focus()
		}
	case "enter":
		if m.inputIdx < len(m.inputs)-1 {
			m.inputs[m.inputIdx].Blur(); m.inputIdx++; m.inputs[m.inputIdx].Focus()
		} else {
			return m, m.submitForm()
		}
	default:
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
	}
	return m, cmd
}

func (m EntryModel) submitForm() tea.Cmd {
	app := m.app
	switch m.state {
	case stateLogin:
		u, p := m.inputs[0].Value(), m.inputs[1].Value()
		return func() tea.Msg { id, err := app.Login(u, p); return loginDoneMsg{id, err} }
	case stateRegister:
		u, e, p := m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value()
		return func() tea.Msg { id, err := app.Register(u, e, p); return registerDoneMsg{id, err} }
	case stateConfigDB:
		d, dsn := m.inputs[0].Value(), m.inputs[1].Value()
		return func() tea.Msg { return configDBDoneMsg{app.ConfigDB(d, dsn)} }
	}
	return nil
}

func (m EntryModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleTitle.Render("MoniMaster CLI") + "\n\n")
	if m.err != nil {
		sb.WriteString(StyleError.Render("Erro: "+m.err.Error()) + "\n\n")
	}
	if m.state == stateMenu {
		for i, item := range menuItems {
			if i == m.cursor {
				sb.WriteString("▶ " + StyleSelected.Render(item) + "\n")
			} else {
				sb.WriteString("  " + item + "\n")
			}
		}
		sb.WriteString("\n" + StyleInfo.Render("↑/↓ navegar  Enter confirmar") + "\n")
		return sb.String()
	}
	titles := map[entryState]string{stateLogin: "Login", stateRegister: "Cadastro", stateConfigDB: "Configurar banco"}
	sb.WriteString(StyleHeader.Render(titles[m.state]) + "\n\n")
	for i, inp := range m.inputs {
		prefix := "  "
		if i == m.inputIdx { prefix = StylePrompt.Render("▶ ") }
		sb.WriteString(prefix + inp.View() + "\n")
	}
	sb.WriteString("\n" + StyleInfo.Render("Tab navegar  Enter confirmar  Esc voltar") + "\n")
	return sb.String()
}

type inputSpec struct {
	placeholder string
	echo        textinput.EchoMode
}

func makeInputs(specs []inputSpec) []textinput.Model {
	inputs := make([]textinput.Model, len(specs))
	for i, spec := range specs {
		ti := textinput.New()
		ti.Placeholder = spec.placeholder
		ti.EchoMode = spec.echo
		if i == 0 { ti.Focus() }
		inputs[i] = ti
	}
	return inputs
}

func RunEntry(app AppFacade) (EntryResult, error) {
	p := tea.NewProgram(NewEntryModel(app), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil { return EntryResult{}, err }
	return final.(EntryModel).Result, nil
}

var _ tea.Model = EntryModel{}
```

**Step 2: Commit**

```bash
git add internal/tui/entry.go
git commit -m "feat(tui): tela de entrada com bubbletea, navegacao por setas e formularios"
```

---

### Task 14: Criar internal/tui/shell.go

**Files:**
- Create: `internal/tui/shell.go`

**Step 1: Implementar**

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/cli"
)

type execDoneMsg struct{ exit bool; output string; err error }

type ShellModel struct {
	app      AppFacade
	input    textinput.Model
	vp       viewport.Model
	history  []string
	histIdx  int
	ready    bool
	Exit     bool
}

func NewShellModel(app AppFacade) ShellModel {
	ti := textinput.New()
	ti.Placeholder = "digite um comando..."
	ti.Prompt = StylePrompt.Render("monimaster> ")
	ti.Focus()
	return ShellModel{app: app, input: ti, histIdx: -1}
}

func (m ShellModel) Init() tea.Cmd { return textinput.Blink }

func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd, vpCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.vp = viewport.New(msg.Width, msg.Height-3)
			m.vp.SetContent(StyleInfo.Render("Sessão pronta. Digite `help` para ver os comandos."))
			m.ready = true
		} else {
			m.vp.Width, m.vp.Height = msg.Width, msg.Height-3
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if len(m.history) > 0 {
				if m.histIdx < len(m.history)-1 { m.histIdx++ }
				m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
			}
			return m, nil
		case "down":
			if m.histIdx > 0 {
				m.histIdx--
				m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
			} else {
				m.histIdx = -1
				m.input.SetValue("")
			}
			return m, nil
		case "tab":
			m.input.SetValue(autoComplete(m.input.Value()))
			return m, nil
		case "enter":
			line := strings.TrimSpace(m.input.Value())
			if line == "" { return m, nil }
			m.history = append(m.history, line)
			m.histIdx = -1
			m.input.SetValue("")
			cmd, err := cli.Parse(line)
			if err != nil {
				if m.ready { m.vp.SetContent(StyleError.Render("Erro: " + err.Error())) }
				return m, nil
			}
			app := m.app
			return m, func() tea.Msg {
				exit, output, execErr := app.Execute(cmd)
				return execDoneMsg{exit, output, execErr}
			}
		}
	case execDoneMsg:
		if msg.exit { m.Exit = true; return m, tea.Quit }
		content := msg.output
		if msg.err != nil { content = StyleError.Render("Erro: " + msg.err.Error()) }
		if m.ready { m.vp.SetContent(content); m.vp.GotoBottom() }
	}
	m.input, tiCmd = m.input.Update(msg)
	if m.ready { m.vp, vpCmd = m.vp.Update(msg) }
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m ShellModel) View() string {
	if !m.ready { return "" }
	return m.vp.View() + "\n" + m.input.View() + "\n"
}

var completions = []string{
	"auth login", "auth register", "auth logout",
	"cleanup interval",
	"config db", "config notify", "config show", "config smtp", "config wizard",
	"doctor run",
	"exit", "help",
	"history show",
	"logs clear", "logs export", "logs show",
	"monitor alert", "monitor dashboard", "monitor once", "monitor start", "monitor status", "monitor stop",
	"notify email",
	"portscan run",
	"profile",
	"report ports", "report uptime",
	"sites add", "sites import", "sites list", "sites remove", "sites update",
	"version",
}

func autoComplete(input string) string {
	if input == "" { return input }
	for _, c := range completions {
		if strings.HasPrefix(c, input) { return c }
	}
	return input
}

func RunShell(app AppFacade) error {
	_, err := tea.NewProgram(NewShellModel(app), tea.WithAltScreen()).Run()
	return err
}

var _ tea.Model = ShellModel{}
```

**Step 2: Commit**

```bash
git add internal/tui/shell.go
git commit -m "feat(tui): shell com viewport, historico ↑↓ e tab auto-complete"
```

---

### Task 15: Criar internal/tui/dashboard.go

**Files:**
- Create: `internal/tui/dashboard.go`

**Step 1: Implementar**

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/monitor"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type dashTickMsg  time.Time
type dashResultMsg []shared.SiteResult

type DashboardModel struct {
	app       AppFacade
	results   []shared.SiteResult
	status    monitor.Status
	resultsCh <-chan []shared.SiteResult
	interval  time.Duration
}

func NewDashboardModel(app AppFacade, interval time.Duration) DashboardModel {
	return DashboardModel{app: app, resultsCh: app.SubscribeResults(), interval: interval}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(m.interval, func(t time.Time) tea.Msg { return dashTickMsg(t) }),
		m.waitResults(),
	)
}

func (m DashboardModel) waitResults() tea.Cmd {
	ch := m.resultsCh
	return func() tea.Msg { return dashResultMsg(<-ch) }
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" { return m, tea.Quit }
	case dashTickMsg:
		m.status = m.app.MonitorStatus()
		return m, tea.Tick(m.interval, func(t time.Time) tea.Msg { return dashTickMsg(t) })
	case dashResultMsg:
		m.results = []shared.SiteResult(msg)
		return m, m.waitResults()
	}
	return m, nil
}

func (m DashboardModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleTitle.Render("MoniMaster Dashboard") + "\n")

	online, offline, warn := 0, 0, 0
	for _, r := range m.results {
		switch {
		case !r.Online:                    offline++
		case r.LatencyWarn || r.CertWarn: warn++
		default:                           online++
		}
	}
	sb.WriteString(fmt.Sprintf("  %s %d online  %s %d offline  %s %d warn  ciclos: %d\n\n",
		StyleOnline.Render(IconOnline), online,
		StyleOffline.Render(IconOffline), offline,
		StyleWarn.Render(IconWarn), warn,
		m.status.CycleCount,
	))

	if len(m.results) == 0 {
		sb.WriteString(StyleInfo.Render("Aguardando resultados... (use monitor start primeiro)") + "\n")
	} else {
		t := Table{Headers: []string{"Site", "Status", "Cód", "Latência", "Cert", "Checado"}}
		for _, r := range m.results {
			status := StyleOnline.Render(IconOnline + " online")
			if !r.Online        { status = StyleOffline.Render(IconOffline + " offline") }
			if r.LatencyWarn    { status = StyleWarn.Render(IconWarn + " lento") }
			certInfo := "-"
			if !r.CertExpiry.IsZero() {
				days := int(time.Until(r.CertExpiry).Hours() / 24)
				if r.CertWarn {
					certInfo = StyleWarn.Render(fmt.Sprintf("%dd", days))
				} else {
					certInfo = StyleOnline.Render(fmt.Sprintf("%dd", days))
				}
			}
			t.Rows = append(t.Rows, TableRow{
				r.Site, status, fmt.Sprintf("%d", r.StatusCode),
				r.Latency.Round(time.Millisecond).String(),
				certInfo, r.CheckedAt.Format("15:04:05"),
			})
		}
		sb.WriteString(t.Render())
	}
	sb.WriteString("\n" + StyleInfo.Render("q para voltar ao shell") + "\n")
	return sb.String()
}

func RunDashboard(app AppFacade, interval time.Duration) error {
	_, err := tea.NewProgram(NewDashboardModel(app, interval), tea.WithAltScreen()).Run()
	return err
}

var _ tea.Model = DashboardModel{}
```

**Step 2: Commit**

```bash
git add internal/tui/dashboard.go
git commit -m "feat(tui): dashboard live com goroutine de resultados e tabela de status"
```

---

### Task 16: Atualizar app.go para implementar AppFacade e usar TUI

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Adicionar canal de resultados ao App e implement AppFacade**

No struct `App`, adicionar: `resultsCh chan []shared.SiteResult`

No `Run()`, após `app.monitor = monitor.NewService()`:
```go
app.resultsCh = make(chan []shared.SiteResult, 4)
```

Adicionar métodos de AppFacade:

```go
func (a *App) Login(username, password string) (*shared.Identity, error) {
	identity, err := a.auth.Login(username, password)
	if err != nil { return nil, err }
	a.session = Session{Mode: shared.SessionModeAuthenticated, Identity: &identity}
	return &identity, nil
}

func (a *App) Register(username, email, password string) (*shared.Identity, error) {
	identity, err := a.auth.Register(username, email, password)
	if err != nil { return nil, err }
	a.session = Session{Mode: shared.SessionModeAuthenticated, Identity: &identity}
	return &identity, nil
}

func (a *App) ConfigDB(driver, dsn string) error {
	a.cfg.Storage = config.StorageConfig{Enabled: true, Driver: driver, DSN: dsn}
	return a.reloadStoreAndSave()
}

func (a *App) SetupWizard(useDB bool, driver, dsn, defaultEmail string) error {
	if useDB {
		a.cfg.Storage = config.StorageConfig{Enabled: true, Driver: driver, DSN: dsn}
		if err := a.reloadStoreAndSave(); err != nil { return err }
	}
	if defaultEmail != "" { a.cfg.Notification.DefaultEmail = defaultEmail }
	return a.manager.Save(a.cfg)
}

func (a *App) MonitorStatus() monitor.Status { return a.monitor.Status() }

func (a *App) SubscribeResults() <-chan []shared.SiteResult { return a.resultsCh }
```

**Step 2: Refatorar handlers para retornar string em vez de imprimir**

Criar helper:
```go
func captureOutput(fn func(w io.Writer)) string {
	var sb strings.Builder
	fn(&sb)
	return sb.String()
}
```

Adicionar import `"io"` e `"strings"` se não estiverem.

Refatorar `handleSites`, `handleLogs`, `handleReport`, `handleHistory`, `handlePortscan`, `handleMonitor` para aceitar `io.Writer` e escrever nele em vez de chamar `fmt.Println` direto. Criar variante pública `Execute`:

```go
func (a *App) Execute(cmd cli.Command) (bool, string, error) {
	var sb strings.Builder
	exit, err := a.executeWriter(cmd, &sb)
	return exit, sb.String(), err
}
```

Renomear o método atual `execute` para `executeWriter` e passar `io.Writer` para todos os sub-handlers. Onde antes havia `fmt.Println(...)`, usar `fmt.Fprintln(w, ...)`.

**Step 3: Adicionar monitor dashboard ao executeWriter**

```go
case "monitor":
	if len(cmd.Path) >= 2 && cmd.Path[1] == "dashboard" {
		interval := time.Duration(a.cfg.Monitor.DelaySeconds) * time.Second
		if err := tui.RunDashboard(a, interval); err != nil {
			return false, err
		}
		return false, nil
	}
	return false, a.handleMonitor(cmd, w)
```

**Step 4: Substituir entry() e shell() no Run()**

```go
result, err := tui.RunEntry(app)
if err != nil { return err }
if !result.Proceed { return nil }
if result.Identity != nil {
	app.session = Session{Mode: result.Mode, Identity: result.Identity}
} else {
	app.session = Session{Mode: result.Mode}
}
return tui.RunShell(app)
```

**Step 5: Enviar resultados para canal no handleMonitorResults**

```go
select {
case a.resultsCh <- results:
default: // dashboard nao ativo, descartar
}
```

**Step 6: Rodar build e testes**

```
go build ./...
go test ./...
```

**Step 7: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): implementa AppFacade, delega entry/shell para tui Charm"
```

---

## PHASE C — Distribuição

### Task 17: Criar .goreleaser.yaml e GitHub Actions

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/release.yml`

**Step 1: Criar .goreleaser.yaml**

```yaml
version: 2
project_name: monimaster

before:
  hooks:
    - go mod tidy

builds:
  - id: native
    binary: monimaster
    env: [CGO_ENABLED=1]
    goos: [linux]
    goarch: [amd64]
    flags: ["-trimpath"]
    ldflags: ["-s -w -X main.version={{.Version}}"]

  - id: cross
    binary: monimaster
    env: [CGO_ENABLED=0]
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
    ignore:
      - goos: linux
        goarch: amd64
    flags: ["-trimpath"]
    ldflags: ["-s -w -X main.version={{.Version}}"]

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: gustavoz65
    name: MonitoradorMain
```

**Step 2: Criar .github/workflows/release.yml**

```yaml
name: Release
on:
  push:
    tags: ["v*"]
permissions:
  contents: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true
      - run: sudo apt-get install -y gcc
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 3: Commit**

```bash
git add .goreleaser.yaml .github/
git commit -m "ci: GoReleaser e GitHub Actions para release automatico"
```

---

### Task 18: Criar install.sh e atualizar README.md

**Files:**
- Create: `install.sh`
- Modify: `README.md`

**Step 1: Criar install.sh**

```bash
#!/usr/bin/env bash
set -e

REPO="gustavoz65/MonitoradorMain"
BIN="monimaster"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *) echo "Arquitetura $ARCH nao suportada" && exit 1 ;;
esac

LATEST=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
FILE="${BIN}_${LATEST#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILE"

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

echo "Baixando MoniMaster $LATEST ($OS/$ARCH)..."
curl -sSfL "$URL" -o "$TMP/$FILE"
curl -sSfL "https://github.com/$REPO/releases/download/$LATEST/checksums.txt" -o "$TMP/checksums.txt"

cd "$TMP"
grep "$FILE" checksums.txt | sha256sum -c -
tar -xzf "$FILE"

DEST="/usr/local/bin"
if [ ! -w "$DEST" ]; then
  sudo mv "$BIN" "$DEST/$BIN"
else
  mv "$BIN" "$DEST/$BIN"
fi

echo "MoniMaster instalado. Versao instalada:"
"$DEST/$BIN" version
```

**Step 2: Adicionar seção de instalação no início do README.md**

Adicionar antes do conteúdo existente:

```markdown
## Instalação

```bash
curl -sSf https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.sh | bash
```

Ou baixe o binário na página de [Releases](https://github.com/gustavoz65/MonitoradorMain/releases).

## Quickstart

```bash
monimaster
# Selecione "Continuar anônimo" para começar sem banco

monimaster> sites add https://example.com --check-cert
monimaster> monitor once
monimaster> monitor start
monimaster> monitor dashboard
monimaster> logs show
```
```

**Step 3: Commit**

```bash
chmod +x install.sh
git add install.sh README.md
git commit -m "ci: script de instalacao e README com quickstart"
```

---

### Task 19: Verificação final

**Step 1: Rodar todos os testes**

```
go test ./...
```
Esperado: PASS em todos os pacotes

**Step 2: Build com cgo**

```
go build -o monimaster .
./monimaster version
```
Esperado: `MoniMaster CLI 3.0.0`

**Step 3: Build sem cgo (simula cross-compile)**

```
CGO_ENABLED=0 go build -o monimaster-nocgo .
./monimaster-nocgo version
```

**Step 4: Tag e push**

```bash
git tag v3.0.0
git push origin main --tags
```
O GitHub Actions dispara o release automaticamente.
