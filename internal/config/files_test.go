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
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 2 {
		t.Fatalf("got %d, want 2", len(configs))
	}
	if configs[0].URL != "https://a.com" {
		t.Fatalf("got %s", configs[0].URL)
	}
}

func TestSaveAndLoadSiteConfigs(t *testing.T) {
	dir := t.TempDir()
	mgr := &Manager{homeDir: dir}

	in := []shared.SiteConfig{{URL: "https://example.com", Method: "GET", CheckCert: true}}
	if err := mgr.SaveSiteConfigs(in); err != nil {
		t.Fatal(err)
	}

	out, err := mgr.LoadSiteConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].URL != "https://example.com" {
		t.Fatal("round-trip falhou")
	}
	if !out[0].CheckCert {
		t.Fatal("check_cert nao persistido")
	}
}
