# MoniMaster CLI

## Instalacao

```bash
curl -sSf https://raw.githubusercontent.com/gustavoz65/MonitoradorMain/main/install.sh | bash
```

Ou baixe o binario na pagina de [Releases](https://github.com/gustavoz65/MonitoradorMain/releases).

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
sites add https://example.com
```

Remover:

```text
sites remove https://example.com
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
monitor start --mode infinite
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
8. Colocar em execucao com `monitor start --hours 2` ou `monitor start --mode infinite`

## Observacoes sobre Oracle

O projeto aceita Oracle como backend configuravel via `database/sql`. A string DSN exata pode variar conforme o ambiente do usuario.

## Desenvolvimento

Rodar testes:

```bash
go test ./...
```

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
