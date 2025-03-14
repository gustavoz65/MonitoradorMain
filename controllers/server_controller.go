package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func IniciarServidorHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Pong! Servidor está ativo.")
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8000"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Canal para captura de interrupção (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		fmt.Println("\nEncerrando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("Erro ao desligar servidor: %v\n", err)
		}
	}()

	fmt.Printf("Servidor HTTP iniciado em http://localhost%s/ping\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Erro ao iniciar servidor HTTP: %v\n", err)
	}
}
