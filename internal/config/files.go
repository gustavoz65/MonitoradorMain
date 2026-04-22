package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

const (
	configFileName  = "config.json"
	sitesFileName   = "sites.json"
	logsFileName    = "logs.jsonl"
	historyFileName = "history.jsonl"
)

type Manager struct {
	homeDir string
}

func NewManager() (*Manager, error) {
	homeDir := os.Getenv("MONIMASTER_HOME")
	if homeDir == "" {
		homeDir = ".monimaster"
	}
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{homeDir: homeDir}, nil
}

func (m *Manager) HomeDir() string {
	return m.homeDir
}

func (m *Manager) ConfigPath() string {
	return filepath.Join(m.homeDir, configFileName)
}

func (m *Manager) SitesPath() string {
	return filepath.Join(m.homeDir, sitesFileName)
}

func (m *Manager) LogsPath() string {
	return filepath.Join(m.homeDir, logsFileName)
}

func (m *Manager) HistoryPath() string {
	return filepath.Join(m.homeDir, historyFileName)
}

func (m *Manager) Load() (AppConfig, error) {
	cfg := Default()
	data, err := os.ReadFile(m.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			if err := m.Save(cfg); err != nil {
				return cfg, err
			}
			return cfg, nil
		}
		return cfg, err
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (m *Manager) Save(cfg AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.ConfigPath(), data, 0o644)
}

func (m *Manager) LoadSites() ([]string, error) {
	data, err := os.ReadFile(m.SitesPath())
	if err != nil {
		if os.IsNotExist(err) {
			if err := m.SaveSites([]string{}); err != nil {
				return nil, err
			}
			return []string{}, nil
		}
		return nil, err
	}
	var sites []string
	if len(data) == 0 {
		return []string{}, nil
	}
	if err := json.Unmarshal(data, &sites); err != nil {
		return nil, err
	}
	return sites, nil
}

func (m *Manager) SaveSites(sites []string) error {
	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.SitesPath(), data, 0o644)
}

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

	var configs []shared.SiteConfig
	if err := json.Unmarshal(data, &configs); err == nil && len(configs) > 0 && configs[0].URL != "" {
		return configs, nil
	}

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
