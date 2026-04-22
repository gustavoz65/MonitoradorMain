# MoniMaster Evolution Design

**Data:** 2026-04-22
**Versão alvo:** 3.0.0
**Sequência de fases:** B (Monitoramento Avançado) → A (Charm TUI) → C (Distribuição)

---

## Objetivo

Transformar o MoniMaster de uma CLI funcional em uma ferramenta profissional de monitoramento: checks HTTP ricos, TUI reativa com Charm, distribuição como binário único, e base arquitetural preparada para tracing de rede, auditoria em tempo real e IoT no futuro.

---

## Phase B — Monitoramento Avançado

### Verificação de certificado TLS

Cada site pode ter `--check-cert` ativado. O checker abre um dial TLS separado, extrai a data de expiração e emite alerta quando faltam menos de N dias (padrão 30). O resultado inclui campo `cert_expiry` nos logs e no `report uptime`.

Executado em goroutine separada da checagem HTTP — as duas rodam em paralelo por site, juntadas via `sync.WaitGroup`. A função de extração de dados do cert usa `internal/native` para operações de string sobre o campo CN/SAN.

### Checks HTTP customizados por site

Sites migram de lista de strings para `[]SiteConfig`. Cada `SiteConfig` carrega:

```go
type SiteConfig struct {
    URL            string
    Method         string            // GET (padrão), HEAD, POST
    Headers        map[string]string // headers customizados
    ExpectedStatus string            // "200", "200-299", "2xx"
    BodyMatch      string            // substring ou regex no corpo
    CheckCert      bool
    CertWarnDays   int
    Timeout        time.Duration
}
```

Novos comandos:
```
sites add https://api.example.com --method POST --expected-status 201 --header "Authorization=Bearer token" --check-cert --cert-warn-days 14
sites update https://api.example.com --expected-status 200-299
```

O arquivo `sites.json` passa a serializar `[]SiteConfig` em vez de `[]string`. Migração automática na leitura: string simples é promovida para `SiteConfig{URL: value, Method: "GET"}`.

### Expansão do `internal/native` para checks

Duas novas funções C com stubs Go em `native_stub.go`:

**`ContainsBytes(body, pattern []byte) bool`**
Usa `memmem` do C para varrer o corpo da resposta HTTP. Mais rápido que `bytes.Contains` em payloads grandes. Cada worker do monitor chama em goroutine já existente, sem overhead adicional.

**`HashBytes(data []byte) uint32`**
CRC32 em C para fingerprint de resposta. Detecta mudança de conteúdo entre ciclos de monitoramento sem comparar o corpo inteiro. Armazenado no `SiteResult` como `ContentHash uint32`.

Build tags existentes (`//go:build cgo` e `//go:build !cgo`) se aplicam às novas funções nos mesmos arquivos.

### Thresholds de alerta de latência

Configurável globalmente em `AppConfig.Alert` e por site em `SiteConfig.LatencyWarn`:

```
monitor alert set --latency-warn 500ms --latency-crit 2s
```

Site online mas acima do threshold dispara alerta com nível `warn` (não `error`). O log registra `level=warn` e o email de alerta inclui a latência medida.

### Sistema de notificação plugável

Interface `NotifyProvider` em `internal/notify/`:

```go
type Provider interface {
    Name() string
    Send(cfg AppConfig, to, subject, body string) error
}
```

Providers implementados:
- `SMTPProvider` — atual, migrado para satisfazer a interface
- `ResendProvider` — HTTP POST para api.resend.com com API key

Seleção via config:
```
config notify provider set resend --api-key re_xxxxx
config notify provider set smtp
```

`notify.Service` mantém o provider ativo e despacha alertas em goroutine com canal bufferizado — o ciclo do monitor nunca bloqueia esperando o envio de email.

### Arquivos afetados (Phase B)

| Ação | Arquivo |
|------|---------|
| Modify | `internal/shared/types.go` — `SiteConfig`, `SiteResult` com `ContentHash` e `CertExpiry` |
| Modify | `internal/monitor/service.go` — aceita `[]SiteConfig`, paraleliza TLS+HTTP por site |
| Create | `internal/monitor/checker.go` — lógica de check TLS, body match, threshold |
| Modify | `internal/native/native_cgo.go` — `ContainsBytes`, `HashBytes` |
| Modify | `internal/native/native_stub.go` — stubs Go para as mesmas funções |
| Modify | `internal/native/native.go` — exports públicos |
| Modify | `internal/notify/service.go` — usa interface `Provider`, dispatch async |
| Create | `internal/notify/smtp.go` — `SMTPProvider` |
| Create | `internal/notify/resend.go` — `ResendProvider` |
| Modify | `internal/config/types.go` — `AlertConfig`, `NotifyConfig` com provider/key |
| Modify | `internal/storage/storage.go` — `SaveSiteConfigs`, `LoadSiteConfigs` |
| Modify | `internal/app/app.go` — novos handlers `sites update`, `monitor alert`, `config notify` |
| Modify | `internal/cli/help.go` — novos comandos |

---

## Phase A — Charm TUI

### Dependências novas

```
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss
github.com/charmbracelet/bubbles
```

### Novo pacote `internal/tui/`

**`styles.go`** — paleta centralizada:
- verde: site online / check ok
- vermelho: offline / error
- amarelo: warn (threshold de latência, cert expirando)
- cinza: info / neutro
- azul: títulos e prompts

**`entry.go`** — modelo bubbletea da tela de entrada:
- navegação por ↑/↓ entre opções do menu
- Enter confirma, Esc sai
- Limpa e redesenha a cada tecla — sem scroll acumulado
- Submodelos inline para login, cadastro e wizard (cada um como estado do mesmo modelo)

**`shell.go`** — modelo bubbletea do shell `monimaster>`:
- `textinput` do bubbles para o campo de digitação
- histórico de comandos navegável com ↑/↓
- Tab auto-complete para primeiro e segundo nível de comando
- Output do comando anterior exibido em `viewport` acima do prompt
- A cada novo comando: viewport é limpo e substituído pelo novo output
- Resolve o problema de "tudo corrido" relatado no bash

**`table.go`** — componente de tabela reutilizável com lipgloss:
- bordas, cabeçalho destacado, linhas alternadas
- usado por: `logs show`, `history show`, `report uptime`, `sites list`, `portscan run`

**`progress.go`** — barra de progresso para operações longas:
- `monitor once`: barra avança conforme workers retornam resultados (canal de resultados alimenta `tea.Cmd`)
- `portscan run`: mesma mecânica, canal de resultados do scanner
- Goroutines dos workers enviam para canal; o modelo bubbletea recebe via `tea.Cmd` sem bloquear a UI

**`dashboard.go`** — novo comando `monitor dashboard`:
- Tela full-screen (alt-screen do bubbletea)
- Tabela de sites com: URL, status (●/✗), latência, último check, content hash diff
- Barra de stats no topo: N online / N offline / N warn
- Auto-refresh via `tea.Tick` pelo intervalo configurado em `cfg.Monitor.DelaySeconds`
- Goroutine do monitor envia resultados via canal; modelo recebe com `tea.Cmd`
- `q` volta ao shell

### Impacto na arquitetura

`internal/app/app.go`:
- `entry()` substituído por `tui.RunEntry(cfg) (SessionResult, error)`
- `shell()` substituído por `tui.RunShell(app) error`
- Toda lógica de negócio (`handleMonitor`, `handleSites`, etc.) permanece intacta
- `handleMonitorResults` passa a receber canal `<-chan []shared.SiteResult` em vez de callback síncrono

`internal/cli/parser.go` — sem alteração; shell TUI usa o mesmo `cli.Parse()`.

### Arquivos afetados (Phase A)

| Ação | Arquivo |
|------|---------|
| Create | `internal/tui/styles.go` |
| Create | `internal/tui/entry.go` |
| Create | `internal/tui/shell.go` |
| Create | `internal/tui/table.go` |
| Create | `internal/tui/progress.go` |
| Create | `internal/tui/dashboard.go` |
| Modify | `internal/app/app.go` — delega entry/shell para tui |
| Modify | `internal/monitor/service.go` — suporte a canal de resultados para dashboard |
| Modify | `go.mod` / `go.sum` — novas dependências Charm |
| Modify | `internal/cli/help.go` — adiciona `monitor dashboard` |

---

## Phase C — Distribuição

### GoReleaser

Arquivo `.goreleaser.yaml` na raiz com duas estratégias de build:

1. **Build nativo** (`CGO_ENABLED=1`): gerado para a plataforma corrente, cgo ativo, todos os hot paths em C disponíveis
2. **Builds de release** (`CGO_ENABLED=0`): compilação cruzada para todos os alvos, usa `native_stub.go` automaticamente via build tag

Alvos:
- Linux amd64, arm64
- macOS amd64, arm64
- Windows amd64

Artefatos: binários comprimidos (`.tar.gz` / `.zip`) + `checksums.txt` com SHA256.

### GitHub Actions

Workflow `.github/workflows/release.yml` disparado em `git tag v*`:
- checkout + setup-go + instala GoReleaser
- `goreleaser release --clean`
- Publica artefatos na GitHub Release automaticamente

### Install script

`install.sh` na raiz:
- detecta OS e arch
- baixa binário correto da última release via GitHub API
- verifica SHA256 antes de instalar
- coloca em `/usr/local/bin/monimaster` (Linux/macOS) ou instrui para Windows

```bash
curl -sSf https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.sh | bash
```

### README

Seções atualizadas:
- Instalação (script + download manual)
- Quickstart (5 comandos para começar a monitorar)
- Tabela completa de comandos

### Arquivos afetados (Phase C)

| Ação | Arquivo |
|------|---------|
| Create | `.goreleaser.yaml` |
| Create | `.github/workflows/release.yml` |
| Create | `install.sh` |
| Modify | `README.md` |

---

## Roadmap futuro (pós v3.0.0)

- **Tracing de rede**: ICMP ping nativo e traceroute via cgo, medição de TTL, fingerprint de dispositivos
- **Auditoria em tempo real**: streaming de audit trail completo por usuário autenticado
- **Geolocalização**: lookup de IP com lib C leve em `internal/native`
- **Monitoramento IoT**: suporte a MQTT, CoAP, ping de dispositivos embarcados
- **Notificações fase E**: Slack, Telegram, webhook via `NotifyProvider`

---

## Princípios de concorrência

Em toda implementação nova:
- Operações de I/O de rede sempre em goroutines com worker pool e limite configurável
- Canais com backpressure (bufferizado) para comunicação entre monitor e notificações
- Funções críticas de processamento de bytes delegadas ao `internal/native` (cgo quando disponível)
- Resultados agregados via channels, nunca por mutex em hot path
