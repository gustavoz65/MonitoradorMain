package main

import (
	"fmt"
	"os"

	"github.com/gustavoz65/MoniMaster/controllers"
	"github.com/gustavoz65/MoniMaster/utils"
)

func main() {
	// Se quiser rodar o servidor HTTP em paralelo, descomente:
	// go controllers.IniciarServidorHTTP()

	// Se precisar carregar .env:
	// err := godotenv.Load()
	// if err != nil {
	// 	fmt.Println("Erro ao carregar arquivo .env.")
	// }

	sucessoLogin := controllers.RealizarLogin()
	if !sucessoLogin {
		fmt.Println("Falha no login. Encerrando o programa.")
		os.Exit(1)
	}

	controllers.ExibirIntroducao()
	utils.EsperarEnter()

	// Inicia a goroutine que gerencia o ticker (limpeza automática de logs)
	go controllers.GerenciarTicker()

	// Loop principal
	for {
		utils.ClearScreen()
		controllers.ExibirMenu()
		comando := controllers.LerComando()

		switch comando {
		case 1:
			controllers.SubMonitoramento()
		case 2:
			controllers.ImprimirLogs()
		case 3:
			controllers.ConfigurarIntervaloLimpeza()
		case 4:
			controllers.ConfigurarEmail()
		case 5:
			controllers.RealizarVarreduraDePortas()
		case 0:
			utils.PrintColored("[❌] Saindo do programa...", "red")
			os.Exit(0)
		default:
			fmt.Println("Comando inválido. Por favor, escolha uma das opções do menu.")
			utils.EsperarEnter()
		}
	}
}
