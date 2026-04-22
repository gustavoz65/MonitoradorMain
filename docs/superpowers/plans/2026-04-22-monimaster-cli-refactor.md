# MoniMaster CLI Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reestruturar o MoniMaster como uma CLI completa, modular e usavel, removendo a API HTTP e consolidando as funcionalidades em comandos.

**Architecture:** A aplicacao sera dividida entre entrada interativa, shell de comandos, modulos de dominio e camadas de persistencia opcionais. A experiencia precisa funcionar sem banco e aproveitar banco relacional quando configurado.

**Tech Stack:** Go, cgo opcional, arquivos JSON locais, drivers SQL opcionais, testes Go.

---

### Task 1: Remodelar a estrutura base da aplicacao

**Files:**
- Create: `cmd/monimaster/main.go`
- Create: `internal/app/app.go`
- Create: `internal/app/session.go`
- Modify: `main.go`

- [ ] Criar bootstrap CLI dedicado e tornar `main.go` um ponto de entrada simples para a nova aplicacao
- [ ] Definir estrutura de sessao com modo anonimo/autenticado
- [ ] Remover acoplamento com Gin e rotas HTTP

### Task 2: Criar camada de configuracao local

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/files.go`
- Create: `internal/config/types.go`

- [ ] Implementar leitura e gravacao da configuracao local em JSON
- [ ] Guardar driver de banco, DSN, SMTP e opcoes gerais com mascaramento de segredos
- [ ] Garantir defaults consistentes e criacao automatica dos arquivos

### Task 3: Criar camada de storage

**Files:**
- Create: `internal/storage/storage.go`
- Create: `internal/storage/local.go`
- Create: `internal/storage/sqlite.go`
- Create: `internal/storage/sql.go`
- Create: `internal/storage/models.go`

- [ ] Definir interface unica de storage para usuarios, historico, auditoria e configs
- [ ] Implementar backend local/noop para modo anonimo
- [ ] Implementar adaptadores iniciais para SQLite e SQL generico com dialetos configuraveis

### Task 4: Migrar autenticacao e identidade

**Files:**
- Create: `internal/auth/service.go`
- Create: `internal/auth/password.go`
- Modify: `models/user.go`
- Remove usages from: `controllers/user_controller.go`

- [ ] Levar cadastro/login para fluxo CLI
- [ ] Suportar usuario anonimo quando banco nao estiver disponivel
- [ ] Melhorar regras de senha e validacao de identidade

### Task 5: Implementar shell de comandos

**Files:**
- Create: `internal/cli/shell.go`
- Create: `internal/cli/parser.go`
- Create: `internal/cli/help.go`
- Create: `internal/cli/commands.go`

- [ ] Implementar prompt `monimaster>`
- [ ] Parsear comandos e flags principais
- [ ] Exibir ajuda contextual por modulo

### Task 6: Migrar monitoramento e logs

**Files:**
- Create: `internal/monitor/service.go`
- Create: `internal/monitor/worker.go`
- Create: `internal/report/logs.go`
- Modify: `controllers/monitor_controller.go`
- Modify: `controllers/site_controller.go`
- Modify: `controllers/log_controller.go`

- [ ] Reaproveitar regras existentes de monitoramento
- [ ] Adicionar execucao concorrente com limite de workers
- [ ] Expor `monitor once`, `monitor start`, `monitor status`, `monitor stop`, `logs show`, `logs clear`, `logs export`

### Task 7: Migrar notificacoes e limpeza

**Files:**
- Create: `internal/notify/email.go`
- Create: `internal/notify/service.go`
- Modify: `controllers/mail_controller.go`
- Modify: `controllers/ticker_controller.go`

- [ ] Tornar configuracao de email parte da CLI
- [ ] Implementar `notify email set`, `notify email test` e `cleanup interval set`
- [ ] Preparar estrutura para futuros canais de notificacao

### Task 8: Evoluir port scan e diagnostico

**Files:**
- Create: `internal/portscan/service.go`
- Create: `internal/doctor/service.go`
- Modify: `controllers/portscan_controller.go`

- [ ] Melhorar port scan com concorrencia e parametros configuraveis
- [ ] Adicionar `doctor run` para verificar ambiente
- [ ] Adicionar `report ports` e `history show`

### Task 9: Adicionar cgo opcional em area critica

**Files:**
- Create: `internal/native/native.go`
- Create: `internal/native/native_cgo.go`
- Create: `internal/native/native_stub.go`

- [ ] Introduzir modulo nativo opcional com fallback puro em Go
- [ ] Usar o modulo apenas em operacoes pequenas e bem encapsuladas
- [ ] Garantir build funcional sem cgo

### Task 10: Limpar legado e atualizar documentacao

**Files:**
- Modify: `README.md`
- Modify: `go.mod`
- Remove or repurpose: `routes/routes.go`, `controllers/server_controller.go`, `docker-compose.yml`

- [ ] Remover narrativa de API HTTP
- [ ] Documentar instalacao, configuracao, fluxo de entrada e comandos reais
- [ ] Registrar exemplos de uso e modos com/sem banco

### Task 11: Verificacao final

**Files:**
- Create: `internal/.../*_test.go`

- [ ] Adicionar testes nas partes mais importantes
- [ ] Rodar `go test ./...`
- [ ] Validar fluxo basico da CLI e corrigir arestas
