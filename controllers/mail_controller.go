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

var smtpConfig = struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}{}

func carregarSMTPConfig() error {
	smtpConfig.Host = os.Getenv("SMTP_HOST")
	if strings.TrimSpace(smtpConfig.Host) == "" {
		return errors.New("SMTP_HOST não está configurado")
	}
	portStr := os.Getenv("SMTP_PORT")
	if strings.TrimSpace(portStr) == "" {
		return errors.New("SMTP_PORT não está configurado")
	}
	var err error
	smtpConfig.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("SMTP_PORT inválido: %v", err)
	}
	smtpConfig.User = os.Getenv("SMTP_USER")
	if strings.TrimSpace(smtpConfig.User) == "" {
		return errors.New("SMTP_USER não está configurado")
	}
	smtpConfig.Password = os.Getenv("SMTP_PASSWORD")
	if strings.TrimSpace(smtpConfig.Password) == "" {
		return errors.New("SMTP_PASSWORD não está configurado")
	}
	// Permitir configuração do remetente ou usar um padrão
	smtpConfig.From = os.Getenv("SMTP_FROM")
	if strings.TrimSpace(smtpConfig.From) == "" {
		smtpConfig.From = "MonitoramentoSAOFC@Outlook.com"
	}
	return nil
}

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

	// Carregar configurações SMTP (se ainda não carregadas)
	if err := carregarSMTPConfig(); err != nil {
		fmt.Printf("Erro na configuração SMTP: %v\n", err)
		utils.EsperarEnter()
		return
	}

	if err := MandarEmail(email, assunto, mensagemErro); err != nil {
		fmt.Printf("Falha ao enviar e-mail: %v\n", err)
		utils.EsperarEnter()
	}
}

func MandarEmail(to string, subject string, body string) error {
	// Validação dos parâmetros
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("endereço de e-mail inválido: %v", err)
	}
	if strings.TrimSpace(subject) == "" {
		return errors.New("o assunto do e-mail não pode estar vazio")
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("o corpo do e-mail não pode estar vazio")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", smtpConfig.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.User, smtpConfig.Password)
	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("erro ao enviar e-mail: %v", err)
	}
	fmt.Println("E-mail enviado com sucesso")
	return nil
}
