package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/auth"
	"github.com/gustavoz65/MoniMaster/internal/cli"
	"github.com/gustavoz65/MoniMaster/internal/config"
	"github.com/gustavoz65/MoniMaster/internal/doctor"
	"github.com/gustavoz65/MoniMaster/internal/monitor"
	"github.com/gustavoz65/MoniMaster/internal/notify"
	"github.com/gustavoz65/MoniMaster/internal/portscan"
	"github.com/gustavoz65/MoniMaster/internal/report"
	"github.com/gustavoz65/MoniMaster/internal/shared"
	"github.com/gustavoz65/MoniMaster/internal/storage"
	"github.com/gustavoz65/MoniMaster/internal/tui"
	"github.com/joho/godotenv"
)

var Version = "3.0.0"

type App struct {
	manager   *config.Manager
	cfg       config.AppConfig
	store     storage.Store
	auth      *auth.Service
	notify    *notify.Service
	monitor   *monitor.Service
	session   Session
	reader    *bufio.Reader
	resultsCh chan []shared.SiteResult
	rootCtx   context.Context
	rootDone  context.CancelFunc
}

func Run() error {
	_ = godotenv.Load()
	manager, err := config.NewManager()
	if err != nil {
		return err
	}
	cfg, err := manager.Load()
	if err != nil {
		return err
	}
	store, err := storage.New(cfg.Storage)
	if err != nil {
		fmt.Printf("Falha ao conectar no banco configurado: %v\n", err)
		fmt.Println("Continuando em modo local/anônimo.")
		store = storage.NewNullStore()
		cfg.Storage.Enabled = false
	}

	app := &App{
		manager:   manager,
		cfg:       cfg,
		store:     store,
		auth:      auth.NewService(store),
		notify:    notify.NewService(store),
		monitor:   monitor.NewService(),
		reader:    bufio.NewReader(os.Stdin),
		resultsCh: make(chan []shared.SiteResult, 4),
	}
	app.rootCtx, app.rootDone = context.WithCancel(context.Background())
	defer app.rootDone()
	defer app.store.Close()
	app.session = Session{Mode: shared.SessionModeAnonymous}
	app.applyNotifyProvider()

	if duration, err := shared.ParseFlexibleDuration(app.cfg.Monitor.CleanupInterval); err == nil {
		_ = report.PruneLogsOlderThan(app.manager.LogsPath(), duration)
	}

	result, err := tui.RunEntry(app)
	if err != nil {
		return err
	}
	if !result.Proceed {
		return nil
	}
	if result.Identity != nil {
		app.session = Session{Mode: result.Mode, Identity: result.Identity}
	} else {
		app.session = Session{Mode: result.Mode}
	}
	return tui.RunShell(app)
}

func (a *App) Execute(cmd cli.Command) (bool, string, error) {
	var sb strings.Builder
	exit, err := a.executeWriter(cmd, &sb)
	return exit, sb.String(), err
}

func (a *App) executeWriter(cmd cli.Command, w io.Writer) (bool, error) {
	if len(cmd.Path) == 0 {
		return false, nil
	}
	switch cmd.Path[0] {
	case "help":
		fmt.Fprint(w, cli.HelpText)
	case "exit", "quit":
		a.monitor.Stop()
		return true, nil
	case "version":
		fmt.Fprintf(w, "MoniMaster CLI %s\n", Version)
	case "profile":
		a.printProfile(w)
	case "auth":
		return false, a.handleAuth(cmd)
	case "config":
		return false, a.handleConfig(cmd, w)
	case "doctor":
		return false, a.handleDoctor(w)
	case "sites":
		return false, a.handleSites(cmd, w)
	case "monitor":
		if len(cmd.Path) >= 2 && cmd.Path[1] == "dashboard" {
			interval := time.Duration(a.cfg.Monitor.DelaySeconds) * time.Second
			if err := tui.RunDashboard(a, interval); err != nil {
				return false, err
			}
			return false, nil
		}
		return false, a.handleMonitor(cmd, w)
	case "logs":
		return false, a.handleLogs(cmd, w)
	case "notify":
		return false, a.handleNotify(cmd, w)
	case "cleanup":
		return false, a.handleCleanup(cmd, w)
	case "portscan":
		return false, a.handlePortscan(cmd, w)
	case "history":
		return false, a.handleHistory(cmd, w)
	case "report":
		return false, a.handleReport(cmd, w)
	default:
		fmt.Fprintln(w, "Comando desconhecido. Use `help`.")
	}
	return false, nil
}

func (a *App) Login(username, password string) (*shared.Identity, error) {
	identity, err := a.auth.Login(username, password)
	if err != nil {
		return nil, err
	}
	a.session = Session{Mode: shared.SessionModeAuthenticated, Identity: &identity}
	return &identity, nil
}

func (a *App) Register(username, email, password string) (*shared.Identity, error) {
	identity, err := a.auth.Register(username, email, password)
	if err != nil {
		return nil, err
	}
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
		if err := a.reloadStoreAndSave(); err != nil {
			return err
		}
	}
	if defaultEmail != "" {
		a.cfg.Notification.DefaultEmail = defaultEmail
	}
	return a.manager.Save(a.cfg)
}

func (a *App) MonitorStatus() monitor.Status { return a.monitor.Status() }

func (a *App) SubscribeResults() <-chan []shared.SiteResult { return a.resultsCh }

func (a *App) handleAuth(cmd cli.Command) error {
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use auth login|register|logout")
	}
	switch cmd.Path[1] {
	case "login":
		return a.interactiveLogin()
	case "register":
		return a.interactiveRegister()
	case "logout":
		a.session = Session{Mode: shared.SessionModeAnonymous}
		return nil
	default:
		return fmt.Errorf("subcomando auth inválido")
	}
}

func (a *App) handleConfig(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use config show|wizard|db|smtp|notify")
	}
	switch cmd.Path[1] {
	case "show":
		a.printConfig(w)
	case "wizard":
		return a.runSetupWizard()
	case "db":
		action := firstArg(cmd.Args)
		if len(cmd.Path) >= 3 && cmd.Path[2] == "disable" {
			a.cfg.Storage = config.StorageConfig{}
			return a.reloadStoreAndSave()
		}
		if action == "disable" {
			a.cfg.Storage = config.StorageConfig{}
			return a.reloadStoreAndSave()
		}
		if action != "" && action != "set" {
			return fmt.Errorf("use config db set --driver <driver> --dsn <dsn>")
		}
		driver := cmd.Flags["driver"]
		dsn := cmd.Flags["dsn"]
		if driver == "" || dsn == "" {
			return fmt.Errorf("use config db set --driver <driver> --dsn <dsn>")
		}
		a.cfg.Storage = config.StorageConfig{Enabled: true, Driver: driver, DSN: dsn}
		return a.reloadStoreAndSave()
	case "smtp":
		if firstArg(cmd.Args) != "set" {
			return fmt.Errorf("use config smtp set --host ... --port ... --user ... --password ...")
		}
		port, _ := strconv.Atoi(defaultString(cmd.Flags["port"], "587"))
		a.cfg.SMTP = config.SMTPConfig{
			Host:     cmd.Flags["host"],
			Port:     port,
			User:     cmd.Flags["user"],
			Password: cmd.Flags["password"],
			From:     cmd.Flags["from"],
		}
		if err := a.manager.Save(a.cfg); err != nil {
			return err
		}
		fmt.Fprintln(w, "SMTP atualizado.")
	case "notify":
		return a.handleConfigNotify(cmd, w)
	default:
		return fmt.Errorf("config inválido")
	}
	return nil
}

func (a *App) handleDoctor(w io.Writer) error {
	checks := doctor.Run(a.manager, a.cfg, a.store)
	for _, check := range checks {
		status := "OK"
		if !check.Healthy {
			status = "FAIL"
		}
		fmt.Fprintf(w, "[%s] %s - %s\n", status, check.Name, check.Details)
	}
	return nil
}

func (a *App) handleSites(cmd cli.Command, w io.Writer) error {
	sites, err := a.manager.LoadSiteConfigs()
	if err != nil {
		return err
	}
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use sites list|add|remove|update|import")
	}
	switch cmd.Path[1] {
	case "list":
		if len(sites) == 0 {
			fmt.Fprintln(w, "Nenhum site configurado.")
			return nil
		}
		for index, site := range sites {
			extra := ""
			if site.CheckCert {
				extra += " [cert]"
			}
			if site.BodyMatch != "" {
				extra += " [body]"
			}
			if site.Method != "" && site.Method != "GET" {
				extra += " [" + site.Method + "]"
			}
			fmt.Fprintf(w, "%d. %s%s\n", index+1, site.URL, extra)
		}
	case "add":
		if len(cmd.Args) == 0 {
			return fmt.Errorf("use sites add <url> [--method GET] [--expected-status 200] [--body-match texto] [--check-cert]")
		}
		url := cmd.Args[0]
		for _, existing := range sites {
			if existing.URL == url {
				return fmt.Errorf("site já existe")
			}
		}
		cfg := shared.SiteConfigFromURL(url)
		if v := cmd.Flags["method"]; v != "" {
			cfg.Method = strings.ToUpper(v)
		}
		if v := cmd.Flags["expected-status"]; v != "" {
			cfg.ExpectedStatus = v
		}
		if v := cmd.Flags["body-match"]; v != "" {
			cfg.BodyMatch = v
		}
		if cmd.BoolFlags["check-cert"] {
			cfg.CheckCert = true
		}
		if v := cmd.Flags["cert-warn-days"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.CertWarnDays = n
			}
		}
		sites = append(sites, cfg)
		sort.Slice(sites, func(i, j int) bool { return sites[i].URL < sites[j].URL })
		if err := a.manager.SaveSiteConfigs(sites); err != nil {
			return err
		}
		fmt.Fprintln(w, "Site adicionado.")
	case "remove":
		if len(cmd.Args) == 0 {
			return fmt.Errorf("use sites remove <url>")
		}
		target := cmd.Args[0]
		filtered := sites[:0]
		for _, site := range sites {
			if site.URL != target {
				filtered = append(filtered, site)
			}
		}
		if err := a.manager.SaveSiteConfigs(filtered); err != nil {
			return err
		}
		fmt.Fprintln(w, "Site removido.")
	case "update":
		if len(cmd.Args) == 0 {
			return fmt.Errorf("use sites update <url> [flags]")
		}
		target := cmd.Args[0]
		updated := false
		for i, site := range sites {
			if site.URL != target {
				continue
			}
			if v := cmd.Flags["method"]; v != "" {
				sites[i].Method = strings.ToUpper(v)
			}
			if v := cmd.Flags["expected-status"]; v != "" {
				sites[i].ExpectedStatus = v
			}
			if v := cmd.Flags["body-match"]; v != "" {
				sites[i].BodyMatch = v
			}
			if cmd.BoolFlags["check-cert"] {
				sites[i].CheckCert = true
			}
			if cmd.BoolFlags["no-check-cert"] {
				sites[i].CheckCert = false
			}
			updated = true
			break
		}
		if !updated {
			return fmt.Errorf("site não encontrado: %s", target)
		}
		if err := a.manager.SaveSiteConfigs(sites); err != nil {
			return err
		}
		fmt.Fprintln(w, "Site atualizado.")
	case "import":
		filePath := cmd.Flags["file"]
		if filePath == "" {
			return fmt.Errorf("use sites import --file arquivo.txt")
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		lines := strings.Split(string(data), "\n")
		set := map[string]struct{}{}
		for _, site := range sites {
			set[site.URL] = struct{}{}
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				set[line] = struct{}{}
			}
		}
		merged := make([]shared.SiteConfig, 0, len(set))
		for site := range set {
			merged = append(merged, shared.SiteConfigFromURL(site))
		}
		sort.Slice(merged, func(i, j int) bool { return merged[i].URL < merged[j].URL })
		if err := a.manager.SaveSiteConfigs(merged); err != nil {
			return err
		}
		fmt.Fprintf(w, "%d sites salvos.\n", len(merged))
	default:
		return fmt.Errorf("subcomando sites inválido")
	}
	return nil
}

func (a *App) handleMonitor(cmd cli.Command, w io.Writer) error {
	sites, err := a.manager.LoadSiteConfigs()
	if err != nil {
		return err
	}
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use monitor once|start|status|stop|alert")
	}
	if len(sites) == 0 && cmd.Path[1] != "status" && cmd.Path[1] != "alert" {
		return fmt.Errorf("nenhum site cadastrado; use sites add")
	}
	opts := monitor.Options{
		Workers: a.cfg.Monitor.WorkerCount,
		Timeout: time.Duration(a.cfg.Monitor.TimeoutSeconds) * time.Second,
		Delay:   time.Duration(a.cfg.Monitor.DelaySeconds) * time.Second,
	}
	switch cmd.Path[1] {
	case "once":
		results := a.monitor.CheckSitesOnce(a.rootCtx, sites, opts)
		a.handleMonitorResults(results)
		for _, result := range results {
			fmt.Fprintf(w, "[%s] %s -> %s\n", strings.ToUpper(resultLevel(result)), result.Site, resultMessage(result))
		}
	case "start":
		if cmd.Flags["hours"] != "" {
			hours, err := strconv.Atoi(cmd.Flags["hours"])
			if err != nil {
				return err
			}
			opts.Hours = hours
		}
		if err := a.monitor.Start(a.rootCtx, sites, opts, a.handleMonitorResults); err != nil {
			return fmt.Errorf("monitoramento já está ativo")
		}
		fmt.Fprintln(w, "Monitoramento iniciado em background.")
	case "status":
		status := a.monitor.Status()
		fmt.Fprintf(w, "running=%t sites=%d cycles=%d started=%s last=%s\n",
			status.Running,
			status.SiteCount,
			status.CycleCount,
			emptyTime(status.StartedAt),
			emptyTime(status.LastRunAt),
		)
	case "stop":
		if a.monitor.Stop() {
			fmt.Fprintln(w, "Monitoramento encerrado.")
		} else {
			fmt.Fprintln(w, "Nenhum monitoramento ativo.")
		}
	case "alert":
		return a.handleMonitorAlert(cmd, w)
	default:
		return fmt.Errorf("subcomando monitor inválido")
	}
	return nil
}

func (a *App) handleLogs(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use logs show|clear|export")
	}
	switch cmd.Path[1] {
	case "show":
		logs, err := report.ReadLogs(a.manager.LogsPath())
		if err != nil {
			return err
		}
		if len(logs) == 0 {
			fmt.Fprintln(w, "Nenhum log disponível.")
			return nil
		}
		for _, entry := range logs {
			fmt.Fprintf(w, "[%s] %s %s - %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"), strings.ToUpper(entry.Level), entry.Target, entry.Message)
		}
	case "clear":
		if err := report.ClearFile(a.manager.LogsPath()); err != nil {
			return err
		}
		fmt.Fprintln(w, "Logs limpos.")
	case "export":
		format := defaultString(cmd.Flags["format"], "json")
		output := defaultString(cmd.Flags["output"], filepath.Join(a.manager.HomeDir(), "export", "logs"))
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return err
		}
		logs, err := report.ReadLogs(a.manager.LogsPath())
		if err != nil {
			return err
		}
		path, err := report.ExportLogs(output, logs, format)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, "Logs exportados para", path)
	default:
		return fmt.Errorf("subcomando logs inválido")
	}
	return nil
}

func (a *App) handleNotify(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 || cmd.Path[1] != "email" {
		return fmt.Errorf("use notify email set|test")
	}
	action := firstArg(cmd.Args)
	if action == "" {
		return fmt.Errorf("use notify email set|test")
	}
	switch action {
	case "set":
		if len(cmd.Args) < 2 {
			return fmt.Errorf("use notify email set user@example.com")
		}
		if err := a.notify.SetTarget(&a.cfg, a.session.Identity, cmd.Args[1]); err != nil {
			return err
		}
		if err := a.manager.Save(a.cfg); err != nil {
			return err
		}
		fmt.Fprintln(w, "Email de notificação configurado.")
	case "test":
		target := a.notify.ResolveTarget(a.cfg, a.session.Identity)
		if err := a.notify.SendSync(a.cfg, target, "Teste MoniMaster", "Seu canal de notificacao esta configurado."); err != nil {
			return err
		}
		fmt.Fprintln(w, "Email de teste enviado para", target)
	default:
		return fmt.Errorf("subcomando notify inválido")
	}
	return nil
}

func (a *App) handleCleanup(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 || cmd.Path[1] != "interval" || firstArg(cmd.Args) != "set" {
		return fmt.Errorf("use cleanup interval set 7d")
	}
	if len(cmd.Args) < 2 {
		return fmt.Errorf("use cleanup interval set 7d")
	}
	if _, err := shared.ParseFlexibleDuration(cmd.Args[1]); err != nil {
		return err
	}
	a.cfg.Monitor.CleanupInterval = cmd.Args[1]
	if err := a.manager.Save(a.cfg); err != nil {
		return err
	}
	fmt.Fprintln(w, "Intervalo de limpeza atualizado.")
	return nil
}

func (a *App) handlePortscan(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 || cmd.Path[1] != "run" {
		return fmt.Errorf("use portscan run --host example.com")
	}
	host := cmd.Flags["host"]
	if host == "" && len(cmd.Args) > 0 {
		host = cmd.Args[0]
	}
	if host == "" {
		return fmt.Errorf("host é obrigatório")
	}
	ports, err := portscan.ParsePorts(cmd.Flags["ports"])
	if err != nil {
		return err
	}
	timeout := 800 * time.Millisecond
	if value := cmd.Flags["timeout"]; value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	}
	results := portscan.Scan(a.rootCtx, host, portscan.Options{Ports: ports, Workers: 40, Timeout: timeout})
	for _, result := range results {
		status := "closed"
		if result.Open {
			status = "open"
		}
		fmt.Fprintf(w, "%s:%d %s (%s)\n", result.Host, result.Port, status, result.Latency)
	}
	openCount := 0
	for _, result := range results {
		if result.Open {
			openCount++
		}
	}
	_ = a.recordHistory("portscan", host, true, fmt.Sprintf("%d/%d portas abertas", openCount, len(results)))
	return nil
}

func (a *App) handleHistory(cmd cli.Command, w io.Writer) error {
	limit := 20
	if cmd.Flags["limit"] != "" {
		value, err := strconv.Atoi(cmd.Flags["limit"])
		if err != nil {
			return err
		}
		limit = value
	}
	history, err := report.ReadHistory(a.manager.HistoryPath())
	if err != nil {
		return err
	}
	if len(history) == 0 {
		fmt.Fprintln(w, "Nenhum histórico local.")
		return nil
	}
	sort.Slice(history, func(i, j int) bool { return history[i].CreatedAt.After(history[j].CreatedAt) })
	if limit > len(history) {
		limit = len(history)
	}
	for _, entry := range history[:limit] {
		fmt.Fprintf(w, "[%s] %s %s -> %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"), entry.Actor, entry.Action, entry.Details)
	}
	return nil
}

func (a *App) handleReport(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 2 {
		return fmt.Errorf("use report uptime|ports")
	}
	switch cmd.Path[1] {
	case "uptime":
		logs, err := report.ReadLogs(a.manager.LogsPath())
		if err != nil {
			return err
		}
		if len(logs) == 0 {
			fmt.Fprintln(w, "Sem dados para relatório.")
			return nil
		}
		type counter struct{ ok, fail int }
		stats := map[string]counter{}
		for _, entry := range logs {
			value := stats[entry.Target]
			if entry.Level == "info" {
				value.ok++
			} else {
				value.fail++
			}
			stats[entry.Target] = value
		}
		for site, value := range stats {
			total := value.ok + value.fail
			uptime := float64(value.ok) / float64(total) * 100
			fmt.Fprintf(w, "%s -> uptime %.2f%% (%d ok / %d fail)\n", site, uptime, value.ok, value.fail)
		}
	case "ports":
		history, err := report.ReadHistory(a.manager.HistoryPath())
		if err != nil {
			return err
		}
		for _, entry := range history {
			if entry.Action == "portscan" {
				fmt.Fprintf(w, "[%s] %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"), entry.Details)
			}
		}
	default:
		return fmt.Errorf("relatório inválido")
	}
	return nil
}

func (a *App) handleMonitorResults(results []shared.SiteResult) {
	select {
	case a.resultsCh <- results:
	default:
	}
	for _, result := range results {
		level := resultLevel(result)
		message := resultMessage(result)
		if !result.Online {
			target := a.notify.ResolveTarget(a.cfg, a.session.Identity)
			if target != "" {
				_ = a.notify.Send(a.cfg, target, "Alerta MoniMaster: falha detectada", fmt.Sprintf("Falha em %s: %s", result.Site, message))
			}
		}
		entry := shared.LogEntry{
			Timestamp: result.CheckedAt,
			Level:     level,
			Category:  "monitor",
			Target:    result.Site,
			Message:   message,
		}
		_ = report.AppendJSONLine(a.manager.LogsPath(), entry)
		details := message
		if result.Online {
			details = fmt.Sprintf("online em %s", result.Latency)
		}
		_ = a.recordHistory("monitor", result.Site, result.Online, details)
	}
	if duration, err := shared.ParseFlexibleDuration(a.cfg.Monitor.CleanupInterval); err == nil {
		_ = report.PruneLogsOlderThan(a.manager.LogsPath(), duration)
	}
}

func (a *App) interactiveLogin() error {
	if !a.store.Enabled() {
		return fmt.Errorf("não há banco configurado; use configurar banco ou siga anônimo")
	}
	username := a.prompt("Usuário")
	password := a.prompt("Senha")
	identity, err := a.auth.Login(username, password)
	if err != nil {
		return err
	}
	a.session = Session{Mode: shared.SessionModeAuthenticated, Identity: &identity}
	fmt.Printf("Login realizado. Bem-vindo, %s.\n", identity.Username)
	return nil
}

func (a *App) interactiveRegister() error {
	if !a.store.Enabled() {
		return fmt.Errorf("não há banco configurado; use configurar banco antes do cadastro")
	}
	username := a.prompt("Novo usuário")
	email := a.prompt("Email")
	password := a.prompt("Senha")
	identity, err := a.auth.Register(username, email, password)
	if err != nil {
		return err
	}
	a.session = Session{Mode: shared.SessionModeAuthenticated, Identity: &identity}
	fmt.Printf("Conta criada. Bem-vindo, %s.\n", identity.Username)
	return nil
}

func (a *App) runSetupWizard() error {
	fmt.Println("Assistente inicial")
	if err := a.runConfigWizard(); err != nil {
		return err
	}
	if email := strings.TrimSpace(a.prompt("Email padrão para alertas (opcional)")); email != "" {
		a.cfg.Notification.DefaultEmail = email
	}
	return a.manager.Save(a.cfg)
}

func (a *App) runConfigWizard() error {
	useDB := strings.ToLower(strings.TrimSpace(a.prompt("Deseja configurar banco? (s/n)")))
	if useDB == "s" || useDB == "sim" {
		driver := strings.ToLower(strings.TrimSpace(a.prompt("Driver (postgres/mysql/sqlite/oracle)")))
		dsn := strings.TrimSpace(a.prompt("DSN"))
		a.cfg.Storage = config.StorageConfig{Enabled: true, Driver: driver, DSN: dsn}
	} else {
		a.cfg.Storage = config.StorageConfig{}
	}
	return a.reloadStoreAndSave()
}

func (a *App) reloadStoreAndSave() error {
	newStore, err := storage.New(a.cfg.Storage)
	if err != nil {
		return err
	}
	_ = a.store.Close()
	a.store = newStore
	a.auth = auth.NewService(newStore)
	a.notify = notify.NewService(newStore)
	a.applyNotifyProvider()
	if err := a.manager.Save(a.cfg); err != nil {
		return err
	}
	fmt.Println("Configuração salva.")
	return nil
}

func (a *App) printProfile(w io.Writer) {
	fmt.Fprintf(w, "mode=%s\n", a.session.Mode)
	if a.session.Identity != nil {
		fmt.Fprintf(w, "user=%s <%s>\n", a.session.Identity.Username, a.session.Identity.Email)
	}
	fmt.Fprintf(w, "storage=%s enabled=%t\n", a.store.Driver(), a.store.Enabled())
	fmt.Fprintf(w, "workspace=%s\n", a.manager.HomeDir())
}

func (a *App) printConfig(w io.Writer) {
	fmt.Fprintf(w, "storage.enabled=%t\n", a.cfg.Storage.Enabled)
	fmt.Fprintf(w, "storage.driver=%s\n", a.cfg.Storage.Driver)
	fmt.Fprintf(w, "storage.dsn=%s\n", shared.MaskSecret(a.cfg.Storage.DSN))
	fmt.Fprintf(w, "smtp.host=%s\n", a.cfg.SMTP.Host)
	fmt.Fprintf(w, "smtp.port=%d\n", a.cfg.SMTP.Port)
	fmt.Fprintf(w, "smtp.user=%s\n", a.cfg.SMTP.User)
	fmt.Fprintf(w, "smtp.password=%s\n", shared.MaskSecret(a.cfg.SMTP.Password))
	fmt.Fprintf(w, "smtp.from=%s\n", a.cfg.SMTP.From)
	fmt.Fprintf(w, "monitor.delay=%ds\n", a.cfg.Monitor.DelaySeconds)
	fmt.Fprintf(w, "monitor.timeout=%ds\n", a.cfg.Monitor.TimeoutSeconds)
	fmt.Fprintf(w, "monitor.workers=%d\n", a.cfg.Monitor.WorkerCount)
	fmt.Fprintf(w, "cleanup.interval=%s\n", a.cfg.Monitor.CleanupInterval)
	fmt.Fprintf(w, "alert.latency_warn=%s\n", a.cfg.Alert.LatencyWarn)
	fmt.Fprintf(w, "alert.latency_crit=%s\n", a.cfg.Alert.LatencyCrit)
	fmt.Fprintf(w, "alert.cert_warn_days=%d\n", a.cfg.Alert.CertWarnDays)
	fmt.Fprintf(w, "notify.provider=%s\n", a.cfg.Notify.Provider)
	fmt.Fprintf(w, "notify.email=%s\n", a.notify.ResolveTarget(a.cfg, a.session.Identity))
}

func (a *App) handleMonitorAlert(cmd cli.Command, w io.Writer) error {
	if firstArg(cmd.Args) != "set" {
		return fmt.Errorf("use monitor alert set [--latency-warn 500ms] [--latency-crit 2s] [--cert-warn-days 30]")
	}
	if v := cmd.Flags["latency-warn"]; v != "" {
		a.cfg.Alert.LatencyWarn = v
	}
	if v := cmd.Flags["latency-crit"]; v != "" {
		a.cfg.Alert.LatencyCrit = v
	}
	if v := cmd.Flags["cert-warn-days"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			a.cfg.Alert.CertWarnDays = n
		}
	}
	if err := a.manager.Save(a.cfg); err != nil {
		return err
	}
	fmt.Fprintln(w, "Thresholds de alerta atualizados.")
	return nil
}

func (a *App) handleConfigNotify(cmd cli.Command, w io.Writer) error {
	if len(cmd.Path) < 3 || cmd.Path[2] != "provider" || firstArg(cmd.Args) != "set" {
		return fmt.Errorf("use config notify provider set smtp|resend [--api-key xxx] [--from email]")
	}
	if len(cmd.Args) < 2 {
		return fmt.Errorf("especifique o provider: smtp ou resend")
	}
	switch strings.ToLower(cmd.Args[1]) {
	case "smtp":
		a.cfg.Notify.Provider = "smtp"
		a.notify.SetProvider(&notify.SMTPProvider{})
	case "resend":
		a.cfg.Notify.Provider = "resend"
		if v := cmd.Flags["api-key"]; v != "" {
			a.cfg.Notify.APIKey = v
		}
		if v := cmd.Flags["from"]; v != "" {
			a.cfg.Notify.From = v
		}
		a.notify.SetProvider(&notify.ResendProvider{})
	default:
		return fmt.Errorf("provider desconhecido; use smtp ou resend")
	}
	if err := a.manager.Save(a.cfg); err != nil {
		return err
	}
	fmt.Fprintf(w, "Provider de notificacao: %s\n", a.cfg.Notify.Provider)
	return nil
}

func (a *App) recordHistory(action, target string, success bool, details string) error {
	record := shared.HistoryRecord{
		ID:        shared.NewID("history"),
		Actor:     a.session.Actor(),
		Mode:      a.session.Mode,
		Action:    action,
		Target:    target,
		Success:   success,
		Details:   details,
		CreatedAt: time.Now(),
	}
	if err := report.AppendJSONLine(a.manager.HistoryPath(), record); err != nil {
		return err
	}
	if a.store.Enabled() {
		_ = a.store.AddHistory(context.Background(), record)
	}
	return nil
}

func (a *App) prompt(label string) string {
	fmt.Printf("%s: ", label)
	value, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func emptyTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04:05")
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(args[0]))
}

func (a *App) applyNotifyProvider() {
	switch strings.ToLower(strings.TrimSpace(a.cfg.Notify.Provider)) {
	case "", "smtp":
		a.notify.SetProvider(&notify.SMTPProvider{})
	case "resend":
		a.notify.SetProvider(&notify.ResendProvider{})
	}
}

func resultLevel(result shared.SiteResult) string {
	if !result.Online {
		return "error"
	}
	return "info"
}

func resultMessage(result shared.SiteResult) string {
	if !result.Online {
		if result.Error != "" {
			return result.Error
		}
		return fmt.Sprintf("site retornou status %d", result.StatusCode)
	}
	return fmt.Sprintf("site online (%d, %s)", result.StatusCode, result.Latency)
}
