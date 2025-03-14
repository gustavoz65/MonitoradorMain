package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gustavoz65/MoniMaster/controllers"
	"github.com/jinzhu/gorm"
)

// SetupRoutes configura as rotas do servidor e injeta o DB no contexto.
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Middleware para injetar o DB no contexto de cada request
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Tela inicial (pode ser expandida para servir HTML)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Bem-vindo ao MoniMaster. Escolha entre login ou cadastro."})
	})

	// Rotas de autenticação
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)

	// Rota de exemplo para varredura de portas
	router.GET("/portscan", controllers.RealizarVarreduraDePortasHTTP)
}
