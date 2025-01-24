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

// SimularCarregamento gera uma animação personalizada de carregamento.
// SimularCarregamento gera uma barra de progresso verde no terminal.
func SimularCarregamento(mensagem string, duracaoSegundos int) {
	barraTamanho := 50 // Tamanho da barra de progresso
	incremento := float64(barraTamanho) / float64(duracaoSegundos*10)
	progresso := 0.0

	verde := "\033[32m" // Cor verde
	reset := "\033[0m"  // Reset da cor

	fmt.Printf("%s\n", mensagem)
	for i := 0; i < duracaoSegundos*10; i++ {
		progresso += incremento
		barra := "["
		for j := 0; j < barraTamanho; j++ {
			if float64(j) < progresso {
				barra += fmt.Sprintf("%s=%s", verde, reset) // Parte preenchida em verde
			} else {
				barra += " " // Parte vazia
			}
		}
		barra += "]"
		fmt.Printf("\r%s", barra) // Atualiza a barra no terminal
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("\nConcluído!") // Finaliza com mensagem de conclusão
}
