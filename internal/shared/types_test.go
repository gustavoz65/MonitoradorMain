package shared

import (
	"testing"
	"time"
)

func TestParseFlexibleDuration(t *testing.T) {
	duration, err := ParseFlexibleDuration("7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if duration != 7*24*time.Hour {
		t.Fatalf("unexpected duration: %v", duration)
	}
}

func TestSiteConfigEffective(t *testing.T) {
	cfg := SiteConfig{URL: "https://example.com"}.Effective()
	if cfg.Method != "GET" {
		t.Fatal("default method should be GET")
	}
	if cfg.CertWarnDays != 30 {
		t.Fatal("default cert warn days should be 30")
	}
}

func TestSiteConfigFromURL(t *testing.T) {
	cfg := SiteConfigFromURL("https://example.com")
	if cfg.URL != "https://example.com" {
		t.Fatalf("url mismatch: %s", cfg.URL)
	}
}
