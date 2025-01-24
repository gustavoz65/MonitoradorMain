package controllers

import (
	"fmt"
	"time"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

// SubMonitoramento exibe opções de monitoramento e inicia conforme escolhido.
func SubMonitoramento() {
	utils.ClearScreen()
	utils.SimularCarregamento("Preparando monitoramento", 3)
	fmt.Println("===== Tipo de Monitoramento =====")
	fmt.Println("1- Monitoramento infinito")
	fmt.Println("2- Monitoramento por X horas")
	fmt.Println("3- Voltar ao menu principal")
	fmt.Println("==================================")
	fmt.Print("Escolha uma opção: ")

	var opcao int
	fmt.Scanln(&opcao)

	switch opcao {
	case 1:
		utils.ClearScreen()
		go IniciarChatCommands() // inicia goroutine que escuta /sair
		monitoramentoInfinito()
	case 2:
		utils.ClearScreen()
		fmt.Println("Digite a quantidade de horas para o monitoramento:")
		var horas int
		fmt.Scanln(&horas)
		if horas > 0 {
			utils.ClearScreen()
			go IniciarChatCommands()
			iniciarMonitoramentoPorHoras(horas)
		} else {
			fmt.Println("Valor de horas inválido.")
			utils.EsperarEnter()
		}
	case 3:
		return
	default:
		fmt.Println("Opção inválida.")
		utils.EsperarEnter()
	}
}

func monitoramentoInfinito() {
	utils.ClearScreen()
	fmt.Println("Iniciando Monitoramento Infinito...")
	sites := LerSitesDoArquivo()
	for {
		select {
		case <-models.EncerrarMonitoramento:
			utils.ClearScreen()
			fmt.Println("Monitoramento encerrado pelo chat commands.")
			utils.EsperarEnter()
			return
		default:
			for i := 0; i < models.Monitoramentos; i++ {
				for idx, site := range sites {
					if site == "" {
						continue
					}
					fmt.Printf("Testando site %d: %s\n", idx+1, site)
					TestaSite(site)
				}
				time.Sleep(models.Delay * time.Second)
			}
		}
	}
}

func iniciarMonitoramentoPorHoras(horas int) {
	utils.ClearScreen()
	fmt.Printf("Iniciando Monitoramento por %d horas...\n", horas)
	sites := LerSitesDoArquivo()
	duracao := time.Duration(horas) * time.Hour
	inicio := time.Now()

	for {
		select {
		case <-models.EncerrarMonitoramento:
			utils.ClearScreen()
			fmt.Println("Monitoramento encerrado pelo chat commands.")
			utils.EsperarEnter()
			return
		default:
			if time.Since(inicio) >= duracao {
				utils.ClearScreen()
				fmt.Println("Tempo de monitoramento finalizado!")
				utils.EsperarEnter()
				return
			}
			for i := 0; i < models.Monitoramentos; i++ {
				for idx, site := range sites {
					if site == "" {
						continue
					}
					fmt.Printf("Testando site %d: %s\n", idx+1, site)
					TestaSite(site)
				}
				time.Sleep(models.Delay * time.Second)
			}
		}
	}
}
