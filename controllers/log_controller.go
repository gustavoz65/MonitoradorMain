package controllers

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/models"
	"github.com/gustavoz65/MoniMaster/utils"
)

// LerSitesDoArquivo lê o arquivo sites.txt e retorna uma slice com os sites.
func LerSitesDoArquivo() []string {
	var sites []string
	arquivo, err := os.Open("sites.txt")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Arquivo 'sites.txt' não encontrado. Criando um arquivo vazio...")
			file, createErr := os.Create("sites.txt")
			if createErr != nil {
				fmt.Println("Erro ao criar o arquivo 'sites.txt':", createErr)
				return sites
			}
			defer file.Close()
			fmt.Println("Arquivo 'sites.txt' criado com sucesso. Adicione os sites que deseja monitorar e execute novamente.")
			return sites
		}
		fmt.Println("Erro ao abrir o arquivo de sites:", err)
		return sites
	}
	defer arquivo.Close()

	leitor := bufio.NewReader(arquivo)
	for {
		linha, err := leitor.ReadString('\n')
		linha = strings.TrimSpace(linha)
		if linha != "" {
			sites = append(sites, linha)
		}
		if err == io.EOF {
			break
		}
	}
	return sites
}

// RegistraLog salva no log.txt se o site está ok ou não.
func RegistraLog(site string, status bool) {
	models.Mutex.Lock()
	defer models.Mutex.Unlock()

	arquivo, err := os.OpenFile(models.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo de log:", err)
		return
	}
	defer arquivo.Close()

	dataEHora := time.Now().Format("02/01/2006 15:04:05")
	var linhaLog string
	if status {
		linhaLog = fmt.Sprintf("[ %s ] SUCESSO: Site %s está online\n", dataEHora, site)
	} else {
		linhaLog = fmt.Sprintf("[ %s ] ERRO: Site %s está offline ou inacessível\n", dataEHora, site)
	}

	if _, err := arquivo.WriteString(linhaLog); err != nil {
		fmt.Println("Erro ao escrever no log:", err)
	}
}

// ImprimirLogs mostra o conteúdo do log.txt na tela.
func ImprimirLogs() {
	utils.ClearScreen()
	fmt.Println("===== Logs de Monitoramento =====")
	conteudo, err := ioutil.ReadFile(models.LogFileName)
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
