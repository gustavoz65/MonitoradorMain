package cli

const HelpText = `
Comandos principais:
  help
  profile
  version
  exit

Autenticacao:
  auth login
  auth register
  auth logout

Configuracao:
  config show
  config wizard
  config db set --driver postgres --dsn "postgres://..."
  config db disable
  config smtp set --host smtp.example.com --port 587 --user me --password secret --from monitor@example.com
  config notify provider set smtp
  config notify provider set resend --api-key re_xxx --from noreply@seudominio.com

Diagnostico:
  doctor run

Sites:
  sites list
  sites add https://example.com [--method GET] [--expected-status 200-299] [--body-match "pong"] [--check-cert] [--cert-warn-days 14]
  sites update https://example.com [--expected-status 200] [--check-cert] [--no-check-cert]
  sites remove https://example.com
  sites import --file sites.txt

Monitoramento:
  monitor once
  monitor start [--hours 2]
  monitor status
  monitor stop
  monitor alert set [--latency-warn 500ms] [--latency-crit 2s] [--cert-warn-days 30]
  monitor dashboard

Logs e relatorios:
  logs show
  logs clear
  logs export [--format json] [--output caminho/base]
  history show [--limit 20]
  report uptime
  report ports

Notificacoes:
  notify email set user@example.com
  notify email test
  cleanup interval set 7d

Port scan:
  portscan run --host example.com [--ports 22,80,443,8000-8010] [--timeout 800ms]
`
