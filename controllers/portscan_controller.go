package controllers

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RealizarVarreduraDePortasHTTP realiza uma varredura simples (nas portas 80 e 443) no host informado.
func RealizarVarreduraDePortasHTTP(c *gin.Context) {
	host := c.Query("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O parâmetro 'host' é obrigatório"})
		return
	}

	var results []string
	ports := []int{80, 443}
	for _, port := range ports {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err != nil {
			results = append(results, fmt.Sprintf("Porta %d fechada", port))
			continue
		}
		conn.Close()
		results = append(results, fmt.Sprintf("Porta %d aberta", port))
	}
	c.JSON(http.StatusOK, gin.H{"host": host, "results": results})
}
