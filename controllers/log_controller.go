package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

// LerSitesDoArquivo lê o arquivo sites.txt e retorna uma slice com os sites.
func LerSitesDoArquivo() []string {
	var sites []string
	file, err := os.Open("sites.txt")
	if err != nil {
		// Se o arquivo não existir, cria um arquivo vazio.
		if os.IsNotExist(err) {
			fmt.Println("Arquivo 'sites.txt' não encontrado. Criando um arquivo vazio...")
			if newFile, createErr := os.Create("sites.txt"); createErr != nil {
				fmt.Println("Erro ao criar o arquivo 'sites.txt':", createErr)
				return sites
			} else {
				newFile.Close()
				fmt.Println("Arquivo 'sites.txt' criado com sucesso. Adicione os sites que deseja monitorar e execute novamente.")
			}
			return sites
		}
		fmt.Println("Erro ao abrir o arquivo de sites:", err)
		return sites
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha != "" {
			sites = append(sites, linha)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
	}
	return sites
}

// RegistraLog salva no log.txt se o site está online ou não.
func RegistraLog(site string, status bool) {
	models.Mutex.Lock()
	defer models.Mutex.Unlock()

	file, err := os.OpenFile(models.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo de log:", err)
		return
	}
	defer file.Close()

	dataEHora := time.Now().Format("02/01/2006 15:04:05")
	var linhaLog string
	if status {
		linhaLog = fmt.Sprintf("[ %s ] SUCESSO: Site %s está online\n", dataEHora, site)
	} else {
		linhaLog = fmt.Sprintf("[ %s ] ERRO: Site %s está offline ou inacessível\n", dataEHora, site)
	}

	if _, err := file.WriteString(linhaLog); err != nil {
		fmt.Println("Erro ao escrever no log:", err)
	}
}

// ImprimirLogs mostra o conteúdo do log.txt na tela.
func ImprimirLogs() {
	utils.ClearScreen()
	fmt.Println("===== Logs de Monitoramento =====")
	conteudo, err := os.ReadFile(models.LogFileName)
	if err != nil {
		fmt.Println("Erro ao ler log.txt:", err)
		utils.EsperarEnter()
		return
	}
	if len(conteudo) == 0 {
		fmt.Println("Nenhum log encontrado. O arquivo está vazio.")
		utils.EsperarEnter()
		return
	}
	fmt.Println(string(conteudo))
	fmt.Println("==================================")
	utils.EsperarEnter()
}
