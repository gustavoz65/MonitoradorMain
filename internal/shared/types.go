package shared

import (
	"fmt"
	"strings"
	"time"
)

const (
	SessionModeAnonymous     = "anonymous"
	SessionModeAuthenticated = "authenticated"
)

type Identity struct {
	ID       string
	Username string
	Email    string
}

type HistoryRecord struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Mode      string    `json:"mode"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Success   bool      `json:"success"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	Target    string    `json:"target"`
	Message   string    `json:"message"`
}

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
	if s.Method == "" {
		s.Method = "GET"
	}
	if s.CertWarnDays == 0 {
		s.CertWarnDays = 30
	}
	return s
}

func SiteConfigFromURL(url string) SiteConfig {
	return SiteConfig{URL: url}
}

type PortResult struct {
	Host      string        `json:"host"`
	Port      int           `json:"port"`
	Open      bool          `json:"open"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

func NewID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func ParseFlexibleDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return 0, fmt.Errorf("duration vazia")
	}
	if strings.HasSuffix(value, "d") {
		daysPart := strings.TrimSuffix(value, "d")
		var days int
		_, err := fmt.Sscanf(daysPart, "%d", &days)
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("duracao em dias invalida: %s", value)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func MaskSecret(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
