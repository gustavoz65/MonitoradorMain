package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gustavoz65/MoniMaster/config"
	"github.com/gustavoz65/MoniMaster/routes"
	"github.com/joho/godotenv"
)

func main() {
	// Carrega variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Não foi possível carregar .env; utilizando variáveis de ambiente existentes")
	}

	// Inicializa a conexão com o PostgreSQL
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Realiza as migrações (criação das tabelas)
	config.Migrate(db)

	// Cria o servidor Gin
	router := gin.Default()

	// Configura as rotas e injeta a conexão com o DB via middleware
	routes.SetupRoutes(router, db)

	// Inicia o servidor na porta definida em SERVER_PORT (ou 8080 por padrão)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor rodando na porta %s", port)
	router.Run(":" + port)
}
