# MoniMaster CLI Design

**Goal**

Transformar o projeto em uma CLI profissional, orientada a comandos, com entrada interativa para acesso inicial, suporte opcional a persistencia em bancos relacionais e foco total em monitoramento, notificacao e diagnostico.

**Product Direction**

O MoniMaster deixa de ser uma API HTTP e passa a ser uma ferramenta de terminal para terceiros. A experiencia de uso e dividida em duas fases:

1. Entrada interativa para escolha de contexto:
- login
- cadastro
- modo anonimo
- configurar banco
- sair

2. Shell de comandos apos sessao ativa, com prompt proprio `monimaster>`.

**Supported Modes**

- `anonymous`: sem banco, sem conta, sem persistencia relacional obrigatoria
- `authenticated`: com banco configurado e conta local MoniMaster

**Supported Databases**

- PostgreSQL
- MySQL
- SQLite
- Oracle

Todos os bancos devem ser acessados por uma camada unica de storage. Se nao houver banco configurado, a CLI continua funcional em modo anonimo.

**Primary Command Areas**

- `help`, `exit`, `version`, `profile`
- `auth login`, `auth register`, `auth logout`
- `config wizard`, `config show`, `config db set`, `config smtp set`
- `doctor run`
- `sites list`, `sites add`, `sites remove`, `sites import`
- `monitor once`, `monitor start`, `monitor stop`, `monitor status`
- `logs show`, `logs clear`, `logs export`
- `notify email set`, `notify email test`
- `cleanup interval set`
- `portscan run`
- `report uptime`, `report ports`
- `history show`

**Architecture**

O projeto deve ser reorganizado em modulos focados:

- `cmd/monimaster`: bootstrap
- `internal/app`: orquestracao de sessao, shell e fluxo principal
- `internal/cli`: parser, prompt, help e UX textual
- `internal/auth`: regras de autenticacao e identidade
- `internal/config`: leitura e gravacao de configuracao local
- `internal/storage`: interface comum e adaptadores por banco
- `internal/monitor`: monitoramento de sites e concorrencia
- `internal/notify`: notificacoes por email e canais futuros
- `internal/portscan`: varredura de portas com concorrencia
- `internal/report`: historico, exportacao e relatorios
- `internal/shared`: utilitarios pequenos e tipos comuns

**Concurrency**

Partes criticas devem usar concorrencia:

- monitoramento de multiplos sites em paralelo com limite de workers
- port scan com pool de workers e timeout por porta
- exportacoes e agregacoes desacopladas das rotinas principais quando apropriado

**C Interop**

O uso de C deve ser restrito a areas de valor real e baixo risco. O plano inicial e adicionar um pequeno modulo via cgo para operacoes criticas e deterministicas de processamento simples, sem prender a ferramenta a bibliotecas externas pesadas. Se a plataforma nao suportar cgo, a CLI precisa continuar funcional com fallback em Go puro.

**Persistence Strategy**

- configuracao local: arquivo JSON em diretorio da aplicacao
- sites locais: arquivo JSON
- logs locais: arquivo estruturado para leitura e exportacao
- persistencia relacional opcional: usuarios, historico, configuracoes e auditoria

**Error Handling**

- mensagens objetivas e acionaveis
- segredos mascarados ao exibir configuracao
- comandos invalidos mostram ajuda contextual
- falha em banco nao derruba a CLI; apenas reduz capacidades autenticadas

**Testing**

- testes unitarios para parser, config, auth, monitor e port scan
- testes de integracao para storage SQLite
- validacao de fluxo principal via comandos principais

**Out of Scope**

- API HTTP
- frontend web
- dependencia obrigatoria de banco para uso basico
