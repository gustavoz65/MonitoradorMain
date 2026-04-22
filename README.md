# MoniMaster CLI

## Instalacao

Linux/macOS:

```bash
curl -sSf https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.ps1 | iex
```

Windows CMD:

```cmd
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.ps1 | iex"
```

Ou baixe o binario na pagina de [Releases](https://github.com/gustavoz65/MonitoradorMain/releases).

### Como o instalador funciona

Em todas as plataformas, o instalador:

- detecta sistema operacional e arquitetura
- consulta a release mais recente no GitHub
- baixa o binario correspondente
- baixa `checksums.txt`
- valida o arquivo por SHA256 antes de instalar
- executa `monimaster version` no final para confirmar a instalacao

### Onde o binario fica

Linux/macOS:

- instalacao em `/usr/local/bin/monimaster`
- se necessario, o script usa `sudo` para mover o binario

Windows:

- instalacao em `%LOCALAPPDATA%\MoniMaster\bin\monimaster.exe`
- o script adiciona essa pasta ao `PATH` do usuario
- pode ser necessario fechar e abrir o terminal depois da instalacao

### Verificando a instalacao

Depois de instalar, rode:

```bash
monimaster version
```

Saida esperada:

```text
MoniMaster CLI 3.0.0
```

### Instalacao manual

Se preferir nao usar script:

1. Abra a pagina de [Releases](https://github.com/gustavoz65/MonitoradorMain/releases).
2. Baixe o arquivo da sua plataforma.
3. Extraia o binario.
4. Coloque o executavel em uma pasta do `PATH`.
5. Rode `monimaster version`.

## Quickstart

```bash
monimaster
# Selecione "Continuar anonimo" para comecar sem banco

monimaster> sites add https://example.com --check-cert
monimaster> monitor once
monimaster> monitor start
monimaster> monitor dashboard
monimaster> logs show
```

O `MoniMaster` agora e uma ferramenta de terminal focada em monitoramento de sites, port scan, notificacoes e diagnostico operacional. Ele foi refatorado para funcionar como CLI de verdade: entra por uma interface interativa e, depois de ativa, passa a operar por comandos.

## O que a ferramenta faz

- monitora sites uma vez ou continuamente
- escaneia portas com concorrencia
- registra logs e historico local
- envia alertas por e-mail
- gera relatorios simples de uptime e port scan
- funciona sem banco em modo anonimo
- suporta persistencia relacional opcional para cadastro/login e historico autenticado

## Diferenciais desta versao

- CLI modular e limpa
- monitoramento concorrente
- port scan concorrente
- suporte opcional a aceleracao via `cgo`
- modo anonimo funcional
- storage relacional opcional
- comando `doctor` para validar ambiente
- assistente inicial de configuracao

## Bancos suportados

- PostgreSQL
- MySQL
- SQLite
- Oracle

Se nenhum banco for configurado, a ferramenta continua funcionando em modo local/anônimo.

## Como executar

```bash
go run .
```

Ao iniciar, a CLI mostra uma entrada interativa:

- `1` login
- `2` cadastro
- `3` continuar anonimo
- `4` configurar banco
- `5` assistente inicial
- `0` sair

Depois disso, voce entra no prompt:

```text
monimaster>
```

### Primeiros passos recomendados

Para um primeiro uso sem banco:

```text
Continuar anonimo
sites add https://example.com --check-cert
monitor once
monitor start
monitor dashboard
logs show
```

Para um ambiente mais completo:

1. configurar banco com `config db set ...` se quiser login e historico autenticado
2. configurar SMTP com `config smtp set ...` se quiser alertas por e-mail
3. definir e-mail de alerta com `notify email set ...`
4. validar ambiente com `doctor run`
5. adicionar sites e iniciar o monitoramento

## Estrutura local da ferramenta

Por padrao, o MoniMaster cria a pasta `.monimaster` no diretorio atual. Voce pode mudar isso com a variavel:

```bash
MONIMASTER_HOME=/caminho/customizado
```

Arquivos gerados:

- `.monimaster/config.json`
- `.monimaster/sites.json`
- `.monimaster/logs.jsonl`
- `.monimaster/history.jsonl`

## Comandos principais

### Ajuda e sessao

```text
help
profile
version
exit
```

### Autenticacao

```text
auth login
auth register
auth logout
```

Observacao:

- login e cadastro exigem banco configurado
- sem banco, a CLI trabalha em modo anonimo

### Configuracao

Mostrar a configuracao atual:

```text
config show
```

Rodar o assistente:

```text
config wizard
```

Configurar banco:

```text
config db set --driver postgres --dsn "postgres://user:pass@localhost:5432/monimaster?sslmode=disable"
config db set --driver mysql --dsn "user:pass@tcp(localhost:3306)/monimaster"
config db set --driver sqlite --dsn "monimaster.db"
config db set --driver oracle --dsn "oracle://user:pass@localhost:1521/XEPDB1"
```

Desabilitar banco:

```text
config db disable
```

Configurar SMTP:

```text
config smtp set --host smtp.example.com --port 587 --user no-reply@example.com --password secret --from monitor@example.com
```

Configurar provider de notificacao:

```text
config notify provider set smtp
config notify provider set resend --api-key re_xxx --from noreply@seudominio.com
```

### Diagnostico

Verificar ambiente:

```text
doctor run
```

Isso checa workspace, sites, banco e SMTP.

### Sites

Listar:

```text
sites list
```

Adicionar:

```text
sites add https://example.com --check-cert
sites add https://example.com --method GET --expected-status 200-299 --body-match "pong" --check-cert --cert-warn-days 14
```

Remover:

```text
sites remove https://example.com
```

Atualizar:

```text
sites update https://example.com --expected-status 200 --check-cert
sites update https://example.com --no-check-cert
```

Importar de arquivo texto:

```text
sites import --file sites.txt
```

### Monitoramento

Rodar uma vez:

```text
monitor once
```

Rodar continuamente:

```text
monitor start
```

Rodar por horas:

```text
monitor start --hours 2
```

Ver status:

```text
monitor status
```

Parar:

```text
monitor stop
```

Configurar thresholds:

```text
monitor alert set --latency-warn 500ms --latency-crit 2s --cert-warn-days 30
```

Dashboard TUI:

```text
monitor dashboard
```

### Logs

Mostrar:

```text
logs show
```

Limpar:

```text
logs clear
```

Exportar:

```text
logs export --format json --output .monimaster/export/logs
logs export --format csv --output .monimaster/export/logs
logs export --format txt --output .monimaster/export/logs
```

### Notificacoes

Definir e-mail de alerta:

```text
notify email set user@example.com
```

Enviar e-mail de teste:

```text
notify email test
```

Definir limpeza automatica dos logs:

```text
cleanup interval set 7d
```

### Port scan

Escanear portas padrao:

```text
portscan run --host example.com
```

Escanear conjunto customizado:

```text
portscan run --host 127.0.0.1 --ports 22,80,443,8000-8010
portscan run --host 127.0.0.1 --ports 22,80,443 --timeout 800ms
```

### Historico e relatorios

Historico:

```text
history show --limit 20
```

Relatorio de uptime:

```text
report uptime
```

Relatorio de port scan:

```text
report ports
```

## Fluxo recomendado para terceiros

1. Rodar `go run .`
2. Usar `5` para abrir o assistente inicial
3. Configurar banco se quiser login e historico autenticado
4. Configurar SMTP se quiser alertas
5. Adicionar sites com `sites add`
6. Validar o ambiente com `doctor run`
7. Iniciar com `monitor once`
8. Colocar em execucao com `monitor start --hours 2` ou `monitor start`

## Observacoes sobre Oracle

O projeto aceita Oracle como backend configuravel via `database/sql`. A string DSN exata pode variar conforme o ambiente do usuario.

## Desenvolvimento

Rodar testes:

```bash
go test ./...
```

Build local:

```bash
go build ./...
```

Build do binario principal:

```bash
go build -o monimaster .
./monimaster version
```

Sem `cgo`:

```bash
CGO_ENABLED=0 go build -o monimaster-nocgo .
./monimaster-nocgo version
```

## Solucao de problemas

### Windows: `bash` nao encontrado

Se voce rodar `install.sh` no `cmd` e aparecer erro relacionado a `/bin/bash`, use o instalador PowerShell:

```powershell
irm https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.ps1 | iex
```

Ou no `cmd`:

```cmd
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.ps1 | iex"
```

### Windows: `monimaster` ainda nao e reconhecido

- feche e abra o terminal
- confirme que `%LOCALAPPDATA%\MoniMaster\bin` entrou no `PATH`
- rode diretamente:

```powershell
$env:LOCALAPPDATA\MoniMaster\bin\monimaster.exe version
```

### Linux/macOS: permissao ao instalar

Se o script precisar escrever em `/usr/local/bin`, ele pode solicitar `sudo`.

### Instalacao por release errada

Use sempre `monimaster version` depois da instalacao para confirmar a versao ativa.

## Estrutura do projeto

```text
internal/app       fluxo principal, shell e sessao
internal/auth      cadastro, login e senha
internal/cli       parser e ajuda
internal/config    configuracao local
internal/doctor    diagnostico de ambiente
internal/monitor   monitoramento concorrente
internal/native    aceleracao opcional com cgo
internal/notify    notificacoes
internal/portscan  port scan concorrente
internal/report    logs, exportacao e historico
internal/storage   storage opcional para bancos relacionais
internal/shared    tipos e utilitarios pequenos
```
