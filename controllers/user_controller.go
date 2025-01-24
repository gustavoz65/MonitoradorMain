package controllers

import (
	"fmt"
	"strings"

	"github.com/gustavoz65/MoniMaster/models"
)

// IniciarChatCommands fica “escutando” comandos do usuário pra encerrar monitoramento.
// IniciarChatCommands fica “escutando” comandos do usuário para encerrar monitoramento.
func IniciarChatCommands() {
	for {
		fmt.Println("===== Chat Commands =====")
		fmt.Println("Digite um comando (/sair para encerrar o monitoramento):")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "/sair" {
			fmt.Println("Encerrando monitoramento pelo Chat Commands...")
			models.EncerrarMonitoramento <- true
			return // Sai do loop sem chamar EsperarEnter
		} else {
			fmt.Println("Comando não reconhecido. Tente novamente.")
		}
	}
}
