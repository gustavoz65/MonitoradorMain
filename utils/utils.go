package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// PrintColored imprime mensagens coloridas no terminal
func PrintColored(message string, color string) {
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

// ClearScreen limpa o terminal, independente do SO
func ClearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run() // ignora erro
}

// EsperarEnter pausa até o usuário pressionar Enter
func EsperarEnter() {
	fmt.Println("Pressione Enter para continuar...")
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao aguardar a entrada. Continuando o programa...")
	}
}

// SimularCarregamento gera uma animação simples de "..."
func SimularCarregamento(mensagem string, duracaoSegundos int) {
	ClearScreen()
	fmt.Println(mensagem)
	for i := 0; i < duracaoSegundos; i++ {
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
	fmt.Println("\nConcluído!")
}
