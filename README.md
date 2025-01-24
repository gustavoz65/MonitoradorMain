MoniMaster
O MoniMaster é um sistema de monitoramento de disponibilidade de sites e varredura de portas, com registro de logs, envio de notificações por e-mail em caso de falhas e uma interface de linha de comando (CLI) simples para gerenciar as opções.

Visão Geral
Linguagem: Go (Golang)
Arquitetura: Separação em múltiplos pacotes (seguindo uma estrutura parecida com MVC)
Funcionalidades Principais:
Monitoramento de Sites:
Modo infinito: verifica a disponibilidade dos sites repetidamente até ser interrompido.
Modo por X horas: verifica a disponibilidade dos sites pelo tempo definido, então encerra.
Registro de Logs: Grava em log.txt cada resultado (“SUCESSO” ou “ERRO”).
Envio de E-mails: Quando um site falha, dispara e-mail de notificação (configurável).
Limpeza de Logs: Faz limpeza automática do arquivo de log após um intervalo configurado.
Varredura de Portas: Permite identificar quais portas comuns estão abertas em um domínio/IP.
Menu Interativo: Interface em CLI que exibe opções e aguarda o usuário escolher.
Chat Commands: Uma goroutine que permite encerrar o monitoramento ao digitar /sair.
Estrutura de Pastas (Exemplo)
go
Copiar
MoniMaster/
  ├── go.mod
  ├── go.sum
  ├── main.go
  │
  ├── controllers/
  │   ├── log_controller.go
  │   ├── mail_controller.go
  │   ├── menu_controller.go
  │   ├── monitor_controller.go
  │   ├── portscan_controller.go
  │   ├── site_controller.go
  │   ├── ticker_controller.go
  │   └── user_controller.go
  │
  ├── models/
  │   └── app_vars.go
  │
  ├── utils/
  │   └── utils.go
  │
  └── sites.txt
Fluxo de Execução
Início (main.go)

Solicita login do usuário (admin/admin123 por padrão).
Exibe apresentação (ExibirIntroducao).
Inicia a goroutine de gerenciamento do Ticker (limpeza automática de logs).
Entra no loop principal do menu.
Menu Principal

1 - Iniciar Monitoramento:
Escolha entre Monitoramento Infinito ou Monitoramento por X horas.
Cria uma goroutine que fica escutando /sair (encerrando o monitoramento).
Lê sites de sites.txt e começa a checar cada um, gravando logs e enviando e-mail em caso de falha.
2 - Exibir Logs:
Lê o arquivo log.txt e imprime no terminal.
3 - Configurar Limpeza de Logs:
Permite definir o intervalo (em dias) para a limpeza automática de log.txt.
4 - Configurar E-mail de Notificação:
Define o e-mail para onde serão enviados alertas de indisponibilidade.
5 - Realizar Varredura de Portas:
Solicita um IP/domínio e verifica as portas mais comuns (ex.: 80, 443, 22 etc.).
0 - Sair:
Encerra o programa imediatamente.
Monitoramento

As rotinas de monitoramento leem a lista de sites (arquivo sites.txt).
Para cada site, faz um GET (http.Get(site)).
Se houver erro de conexão, registra ERRO no log e envia e-mail (se configurado).
Se der status 200, registra SUCESSO.
Se der outro status code, registra ERRO e envia e-mail com o status code.
Repetição de acordo com modo infinito ou até acabar o tempo (horas) configurado.
Log

Ao final de cada teste de site, salva em log.txt algo como:
less
Copiar
[ 23/01/2025 19:48:02 ] SUCESSO: Site http://exemplo.com está online
ou
less
Copiar
[ 23/01/2025 19:49:10 ] ERRO: Site http://exemplo.com está offline ou inacessível
Ticker de Limpeza de Logs

O ticker_controller.go cria um time.Ticker que dispara a cada IntervaloLimpeza (por padrão, 24h).
Quando esse intervalo ocorre, o arquivo de log (log.txt) é limpo automaticamente.
E-mail

Configurável via ConfigurarEmail.
Se o SMTP não estiver configurado corretamente em variáveis de ambiente (ex.: SMTP_HOST, SMTP_PORT, etc.), o envio falha.
Usa a biblioteca gopkg.in/gomail.v2.
Varredura de Portas

O usuário informa um domínio/IP.
O programa testa portas comuns (ex.: 80, 443, 22 etc.) usando net.DialTimeout.
Exibe “porta aberta” ou “porta fechada” no terminal.
Pré-Requisitos
Go 1.18 ou superior.
Configurar variáveis de ambiente para envio de e-mail, se desejar (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD).
Conexão com a Internet (para testar sites e enviar e-mail).
Como Executar
Clonar este repositório (ou baixar o código-fonte).
No diretório do projeto, inicializar/atualizar módulos:
bash
Copiar
go mod tidy
Executar:
bash
Copiar
go run main.go
O programa perguntará usuário e senha (por padrão, admin / admin123).
Após login, será exibido o menu principal no terminal.
Possíveis Ajustes
Adição/remoção de sites no arquivo sites.txt para monitorar diferentes endereços.
Tempo de monitoramento no modo infinito pode ser alterado na variável delay (intervalo entre requisições).
Destinatário de e-mail: configurado via opção “4 - Configurar E-mail de Notificação” do menu.
Intervalo de limpeza: configurado via opção “3 - Configurar Limpeza de Logs”.
Contribuindo
Issues: abra uma issue para relatar bugs ou sugerir novas funcionalidades.
Pull Requests: crie um PR com suas melhorias e aguarde a revisão.
Licença
(indefinida).
