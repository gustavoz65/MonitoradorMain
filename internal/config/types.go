package config

type StorageConfig struct {
	Enabled bool   `json:"enabled"`
	Driver  string `json:"driver"`
	DSN     string `json:"dsn"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type MonitorConfig struct {
	DelaySeconds    int    `json:"delay_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	WorkerCount     int    `json:"worker_count"`
	CleanupInterval string `json:"cleanup_interval"`
}

type NotificationConfig struct {
	DefaultEmail string `json:"default_email"`
}

type AlertConfig struct {
	LatencyWarn  string `json:"latency_warn"`
	LatencyCrit  string `json:"latency_crit"`
	CertWarnDays int    `json:"cert_warn_days"`
}

type NotifyConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	From     string `json:"from"`
}

type AppConfig struct {
	Storage      StorageConfig      `json:"storage"`
	SMTP         SMTPConfig         `json:"smtp"`
	Monitor      MonitorConfig      `json:"monitor"`
	Notification NotificationConfig `json:"notification"`
	Alert        AlertConfig        `json:"alert"`
	Notify       NotifyConfig       `json:"notify"`
}

func Default() AppConfig {
	return AppConfig{
		Storage: StorageConfig{
			Enabled: false,
			Driver:  "",
			DSN:     "",
		},
		SMTP: SMTPConfig{
			Port: 587,
		},
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
		Notify: NotifyConfig{
			Provider: "smtp",
		},
	}
}
