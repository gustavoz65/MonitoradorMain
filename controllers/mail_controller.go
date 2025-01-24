package controllers

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
	"gopkg.in/gomail.v2"
)

func ConfigurarEmail() {
	models.Mutex.Lock()
	defer models.Mutex.Unlock()

	utils.ClearScreen()
	utils.SimularCarregamento("Carregando configuração de e-mail", 3)
	fmt.Println("===== Configurar E-mail de Notificação =====")
	fmt.Print("Digite o seu E-mail: ")
	var novoEmail string
	fmt.Scanln(&novoEmail)

	// Validação básica do e-mail
	if _, err := mail.ParseAddress(novoEmail); err != nil {
		fmt.Println("Endereço de e-mail inválido. Por favor, tente novamente.")
		utils.EsperarEnter()
		return
	}
	models.EmailNotificacao = novoEmail
	fmt.Printf("E-mail de notificação configurado para: %s\n", models.EmailNotificacao)
	utils.EsperarEnter()
}

// EnviarEmail dispara e-mail se houver erro no site.
func EnviarEmail(site string, mensagemErro string) {
	models.Mutex.Lock()
	email := models.EmailNotificacao
	models.Mutex.Unlock()

	if email == "" {
		fmt.Println("E-mail de notificação não está configurado. Use a opção 4 para configurar.")
		utils.EsperarEnter()
		return
	}

	assunto := "Erro de Monitoramento no Site: " + site
	err := MandarEmail(email, assunto, mensagemErro)
	if err != nil {
		fmt.Printf("Falha ao enviar e-mail: %v\n", err)
		utils.EsperarEnter()
	}
}

// MandarEmail faz a configuração com SMTP e envia o e-mail de fato.
func MandarEmail(to string, subject string, body string) error {

	// Valida e-mail destinatário
	_, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("endereço de e-mail inválido: %v", err)
	}

	// Assunto não vazio
	if strings.TrimSpace(subject) == "" {
		return errors.New("o assunto do e-mail não pode estar vazio")
	}

	// Corpo não vazio
	if strings.TrimSpace(body) == "" {
		return errors.New("o corpo do e-mail não pode estar vazio")
	}

	host := os.Getenv("SMTP_HOST")
	if strings.TrimSpace(host) == "" {
		return errors.New("SMTP_HOST não está configurado nas variáveis de ambiente")
	}

	portStr := os.Getenv("SMTP_PORT")
	if strings.TrimSpace(portStr) == "" {
		return errors.New("SMTP_PORT não está configurado nas variáveis de ambiente")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("SMTP_PORT inválido: %v", err)
	}

	user := os.Getenv("SMTP_USER")
	if strings.TrimSpace(user) == "" {
		return errors.New("SMTP_USER não está configurado nas variáveis de ambiente")
	}

	password := os.Getenv("SMTP_PASSWORD")
	if strings.TrimSpace(password) == "" {
		return errors.New("SMTP_PASSWORD não está configurado nas variáveis de ambiente")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", "MonitoramentoSAOFC@Outlook.com") // Remetente fixo
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(host, port, user, password)

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("erro ao enviar e-mail: %v", err)
	}

	fmt.Println("E-mail enviado com sucesso")
	return nil
}
