package controllers

import (
	"fmt"
	"strings"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

// IniciarChatCommands fica “escutando” comandos do usuário pra encerrar monitoramento.
func IniciarChatCommands() {
	for {
		utils.ClearScreen()
		fmt.Println("===== Chat Commands =====")
		fmt.Println("Digite um comando (/sair para encerrar o monitoramento):")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "/sair" {
			fmt.Println("Encerrando monitoramento pelo Chat Commands...")
			models.EncerrarMonitoramento <- true
			utils.EsperarEnter()
			return
		}
		fmt.Println("Comando não reconhecido.")
		utils.EsperarEnter()
	}
}
