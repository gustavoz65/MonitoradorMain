package controllers

import (
	"fmt"
	"os"
	"time"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

func GerenciarTicker(stopChan <-chan struct{}) {
	if models.Ticker != nil {
		models.Ticker.Stop()
	}
	models.Ticker = time.NewTicker(models.IntervaloLimpeza)
	defer models.Ticker.Stop()

	for {
		select {
		case <-models.Ticker.C:
			limparLogsAutomaticamente()
		case <-stopChan:
			fmt.Println("Encerrando Ticker de limpeza de logs.")
			return
		}
	}
}

func ReiniciarTicker() {
	if models.Ticker != nil {
		models.Ticker.Stop()
	}
	models.Ticker = time.NewTicker(models.IntervaloLimpeza)
	fmt.Println("Ticker reiniciado com o novo intervalo.")
}

func LimparLogsAutomaticamente() {
	models.Mutex.Lock()
	defer models.Mutex.Unlock()

	if err := os.WriteFile(models.LogFileName, []byte{}, 0666); err != nil {
		fmt.Println("Erro ao limpar os logs:", err)
		return
	}
	fmt.Println("Logs limpos automaticamente.")
}

func ConfigurarIntervaloLimpeza() {
	models.Mutex.Lock()
	defer models.Mutex.Unlock()

	utils.ClearScreen()
	fmt.Println("===== Configurar Intervalo de Limpeza de Logs =====")
	fmt.Println("Digite o intervalo em dias:")
	var dias int
	fmt.Scanln(&dias)
	if dias < 1 {
		fmt.Println("Intervalo inválido.")
		utils.EsperarEnter()
		return
	}
	models.IntervaloLimpeza = time.Duration(dias) * 24 * time.Hour
	ReiniciarTicker()
	fmt.Printf("Intervalo configurado para %d dias.\n", dias)
	utils.EsperarEnter()
}

// Alias para facilitar a chamada
func limparLogsAutomaticamente() {
	LimparLogsAutomaticamente()
}
