package controllers

import (
	"fmt"
	"net/http"
)

// TestaSite verifica se o site está online ou não e registra log + envia e-mail.
func TestaSite(site string) {
	resp, err := http.Get(site)
	if err != nil {
		mensagem := fmt.Sprintf("❌ [ERRO] Site %s está offline ou inacessível: %v", site, err)
		printColored(mensagem, "red")
		RegistraLog(site, false)
		EnviarEmail(site, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		mensagem := fmt.Sprintf("✅ [SUCESSO] Site %s está online.", site)
		printColored(mensagem, "green")
		RegistraLog(site, true)
	} else {
		mensagem := fmt.Sprintf("❌ [ERRO] Site %s retornou o status code: %d", site, resp.StatusCode)
		printColored(mensagem, "red")
		RegistraLog(site, false)
		EnviarEmail(site, fmt.Sprintf("Status Code: %d", resp.StatusCode))
	}
}

func printColored(message string, color string) {
	colors := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"purple": "\033[35m",
		"reset":  "\033[0m",
	}

	c, exists := colors[color]
	if !exists {
		c = colors["reset"]
	}
	fmt.Println(c + message + colors["reset"])
}
