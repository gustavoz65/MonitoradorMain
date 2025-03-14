package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TestaSite verifica se o site está online, registra log e envia e-mail em caso de erro.
func TestaSite(site string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, site, nil)
	if err != nil {
		handleSiteError(site, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		handleSiteError(site, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		printColored(fmt.Sprintf("✅ [SUCESSO] Site %s está online.", site), "green")
		RegistraLog(site, true)
	} else {
		printColored(fmt.Sprintf("❌ [ERRO] Site %s retornou o status code: %d", site, resp.StatusCode), "red")
		RegistraLog(site, false)
		EnviarEmail(site, fmt.Sprintf("Status Code: %d", resp.StatusCode))
	}
}

func handleSiteError(site string, err error) {
	printColored(fmt.Sprintf("❌ [ERRO] Site %s está offline ou inacessível: %v", site, err), "red")
	RegistraLog(site, false)
	EnviarEmail(site, err.Error())
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
