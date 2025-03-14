package controllers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

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
		ctx, cancel := context.WithCancel(context.Background())
		go IniciarChatCommandsComCancel(cancel)
		monitoramentoInfinito(ctx)
	case 2:
		utils.ClearScreen()
		fmt.Println("Digite a quantidade de horas para o monitoramento:")
		var horas int
		fmt.Scanln(&horas)
		if horas > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(horas)*time.Hour)
			defer cancel()
			go IniciarChatCommandsComCancel(cancel)
			iniciarMonitoramentoPorHoras(ctx)
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

func monitoramentoInfinito(ctx context.Context) {
	utils.ClearScreen()
	fmt.Println("Iniciando Monitoramento Infinito...")
	sites := LerSitesDoArquivo()
	if len(sites) == 0 {
		fmt.Println("Nenhum site para monitorar. Certifique-se de adicionar URLs ao arquivo 'sites.txt'.")
		utils.EsperarEnter()
		return
	}
	for {
		select {
		case <-ctx.Done():
			utils.ClearScreen()
			fmt.Println("Monitoramento encerrado.")
			utils.EsperarEnter()
			return
		default:
			for _, site := range sites {
				if site != "" {
					fmt.Printf("Testando site: %s\n", site)
					TestaSite(site)
				}
			}
			time.Sleep(time.Duration(models.Delay) * time.Second)
		}
	}
}

func iniciarMonitoramentoPorHoras(ctx context.Context) {
	utils.ClearScreen()
	fmt.Println("Iniciando Monitoramento por tempo determinado...")
	sites := LerSitesDoArquivo()
	if len(sites) == 0 {
		fmt.Println("Nenhum site para monitorar. Verifique o arquivo 'sites.txt'.")
		utils.EsperarEnter()
		return
	}
	for {
		select {
		case <-ctx.Done():
			utils.ClearScreen()
			fmt.Println("Tempo de monitoramento finalizado ou cancelado!")
			utils.EsperarEnter()
			return
		default:
			for idx, site := range sites {
				if site == "" {
					continue
				}
				fmt.Printf("Testando site %d: %s\n", idx+1, site)
				TestaSite(site)
			}
			time.Sleep(time.Duration(models.Delay) * time.Second)
		}
	}
}

func IniciarChatCommandsComCancel(cancelFunc context.CancelFunc) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("===== Chat Commands =====")
		fmt.Println("Digite um comando (/sair para encerrar o monitoramento):")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler comando:", err)
			continue
		}
		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "/sair" {
			fmt.Println("Encerrando monitoramento pelo Chat Commands...")
			cancelFunc()
			return
		} else {
			fmt.Println("Comando não reconhecido. Tente novamente.")
		}
	}
}
