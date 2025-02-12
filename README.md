# MoniMaster

O **MoniMaster** é um sistema de monitoramento de disponibilidade de sites e varredura de portas, com registro de logs, envio de notificações por e-mail em caso de falhas e uma interface de linha de comando (CLI) simples para gerenciar as opções.

## Visão Geral

- **Linguagem:** Go (Golang)
- **Arquitetura:** Separado em múltiplos pacotes, seguindo uma estrutura parecida com MVC.

### Funcionalidades Principais:

#### **Monitoramento de Sites**
- **Modo infinito:** verifica a disponibilidade dos sites repetidamente até ser interrompido.
- **Modo por X horas:** verifica a disponibilidade dos sites pelo tempo definido, então encerra.
- **Registro de Logs:** Grava em `log.txt` cada resultado ("SUCESSO" ou "ERRO").
- **Envio de E-mails:** Quando um site falha, dispara e-mail de notificação (configurável).
- **Limpeza de Logs:** Faz limpeza automática do arquivo de log após um intervalo configurado.

#### **Varredura de Portas**
- Permite identificar quais portas comuns estão abertas em um domínio/IP.

#### **Interface CLI**
- **Menu Interativo:** Exibe opções e aguarda a escolha do usuário.
- **Chat Commands:** Goroutine que permite encerrar o monitoramento ao digitar `/sair`.

## Estrutura de Pastas

```
MoniMaster/
  ├── go.mod
  ├── go.sum
  ├── main.go
  ├── controllers/
  │   ├── log_controller.go
  │   ├── mail_controller.go
  │   ├── menu_controller.go
  │   ├── monitor_controller.go
  │   ├── portscan_controller.go
  │   ├── site_controller.go
  │   ├── ticker_controller.go
  │   └── user_controller.go
  ├── models/
  │   └── app_vars.go
  ├── utils/
  │   └── utils.go
  └── sites.txt
```

## Fluxo de Execução

### **Início (`main.go`)**
1. Solicita login do usuário (**admin/admin123** por padrão).
2. Exibe apresentação (**ExibirIntroducao**).
3. Inicia a goroutine de gerenciamento do **Ticker** (limpeza automática de logs).
4. Entra no loop principal do menu.

### **Menu Principal**

1 - **Iniciar Monitoramento**
   - Escolha entre **Monitoramento Infinito** ou **Monitoramento por X horas**.
   - Cria uma **goroutine** que escuta `/sair` (encerrando o monitoramento).
   - Lê sites de `sites.txt`, checa cada um, grava logs e envia e-mail em caso de falha.

2 - **Exibir Logs**
   - Lê o arquivo `log.txt` e imprime no terminal.

3 - **Configurar Limpeza de Logs**
   - Define o intervalo (em dias) para a limpeza automática do `log.txt`.

4 - **Configurar E-mail de Notificação**
   - Define o e-mail para onde serão enviados alertas de indisponibilidade.

5 - **Realizar Varredura de Portas**
   - Solicita um IP/domínio e verifica as portas mais comuns (ex.: 80, 443, 22 etc.).

0 - **Sair**
   - Encerra o programa imediatamente.

## Monitoramento

- As rotinas de monitoramento leem a lista de sites (`sites.txt`).
- Para cada site, faz um `GET` (`http.Get(site)`).
- Se houver erro de conexão:
  - Registra **ERRO** no log.
  - Envia e-mail (se configurado).
- Se der status **200**, registra **SUCESSO**.
- Se der outro status code, registra **ERRO** e envia e-mail com o status code.

### **Log (`log.txt`)**

```
[ 23/01/2025 19:48:02 ] SUCESSO: Site http://exemplo.com está online
[ 23/01/2025 19:49:10 ] ERRO: Site http://exemplo.com está offline ou inacessível
```

## **Ticker de Limpeza de Logs**

- O `ticker_controller.go` cria um `time.Ticker` que dispara a cada `IntervaloLimpeza` (padrão: 24h).
- Quando esse intervalo ocorre, o arquivo de log (`log.txt`) é limpo automaticamente.

## **E-mail**

- Configurável via `ConfigurarEmail`.
- Se o SMTP não estiver configurado corretamente em variáveis de ambiente (ex.: `SMTP_HOST`, `SMTP_PORT`, etc.), o envio falha.
- Usa a biblioteca `gopkg.in/gomail.v2`.

## **Varredura de Portas**

- O usuário informa um domínio/IP.
- O programa testa portas comuns (ex.: **80, 443, 22** etc.) usando `net.DialTimeout`.
- Exibe "porta aberta" ou "porta fechada" no terminal.

## **Pré-Requisitos**

- **Go 1.18** ou superior.
- Configurar variáveis de ambiente para envio de e-mail (**SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD**).
- Conexão com a Internet (para testar sites e enviar e-mail).

## **Como Executar**

1. Clonar este repositório (ou baixar o código-fonte).
2. No diretório do projeto, inicializar/atualizar módulos:
   ```bash
   go mod tidy
   ```
3. Executar:
   ```bash
   go run main.go
   ```
4. O programa perguntará usuário e senha (**admin / admin123** por padrão).
5. Após login, será exibido o menu principal no terminal.

## **Contribuindo**

- **Issues:** Abra uma issue para relatar bugs ou sugerir novas funcionalidades.
- **Pull Requests:** Crie um PR com suas melhorias e aguarde a revisão.

## **Licença**

Indefinida.

