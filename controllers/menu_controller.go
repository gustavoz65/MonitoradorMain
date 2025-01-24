package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gustavoz65/MoniMaster/utils"
)

func ExibirMenu() {
	roxoEscuro := "\033[35m"
	reset := "\033[0m"
	fmt.Printf("%s===== MoniMaster =====%s\n", roxoEscuro, reset)
	fmt.Println("1- Iniciar Monitoramento")
	fmt.Println("2- Exibir Logs")
	fmt.Println("3- Configurar Limpeza de Logs")
	fmt.Println("4- Configurar E-mail de Notificação")
	fmt.Println("5- Realizar Varredura de Portas")
	fmt.Println("0- Sair do Programa")
	fmt.Println("======================")
	fmt.Print("Escolha uma opção: ")
}

// LerComando lê a opção digitada no menu.
func LerComando() int {
	var input string
	fmt.Scanln(&input)

	// Se digitar "sair", encerra
	if strings.ToLower(input) == "sair" {
		fmt.Println("Saindo do Programa...")
		os.Exit(0)
	}

	var comando int
	_, err := fmt.Sscanf(input, "%d", &comando)
	if err != nil {
		fmt.Println("Entrada inválida. Por favor, digite um número.")
		utils.EsperarEnter()
		return -1
	}
	return comando
}

// ExibirIntroducao apenas imprime a introdução.
func ExibirIntroducao() {
	utils.ClearScreen()
	roxoEscuro := "\033[35m"
	reset := "\033[0m"
	fmt.Printf("%sBem-vindo ao MoniMaster%s\n", roxoEscuro, reset)
	fmt.Println()
}

func RealizarLogin() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite seu usuario: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Digite sua senha: ")
	senha, _ := reader.ReadString('\n')
	senha = strings.TrimSpace(senha)

	if ValidarCredenciais(username, senha) {
		utils.SimularCarregamento("Validando credenciais", 3)
		fmt.Println("Login realizado com sucesso!")
		return true
	}

	fmt.Println("Credenciais inválidas. Tente novamente.")
	return false
}

func ValidarCredenciais(username, senha string) bool {
	// Simulação de validação de credenciais
	return username == "admin" && senha == "admin123"
}
