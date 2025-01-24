package controllers

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/utils"
)

func RealizarVarreduraDePortas() {
	fmt.Println("===== Varredura de Portas =====")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite o domínio ou IP para varredura de portas: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	if host == "" {
		fmt.Println("Nenhum domínio ou IP fornecido.")
		utils.EsperarEnter()
		return
	}

	fmt.Println("\n=== Varredura de Portas ===")
	varrerPortas(host)
	utils.EsperarEnter()
}

func varrerPortas(host string) {
	// Lista de portas comuns
	ports := []int{80, 443, 22, 21, 25, 3306, 8080, 8081, 8082, 8083, 8084, 8085, 8086, 8087, 8088, 8089, 8090,
		49152, 65535, 49151, 1024, 1023, 1022, 1021, 1020,
		1019, 1018, 1017, 1016, 1015, 1014, 1013, 1012, 1011,
		1010, 1009, 1008, 1007, 1006, 1005, 1004, 1003, 1002,
		1001, 1000, 0}

	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err != nil {
			fmt.Printf("❌ Porta %d está fechada\n", port)
			continue
		}
		conn.Close()
		fmt.Printf("✅ Porta %d está aberta\n", port)
	}
}
