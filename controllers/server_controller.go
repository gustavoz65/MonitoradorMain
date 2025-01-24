package controllers

import (
	"fmt"
	"net/http"
)

// Se quiser rodar em paralelo via go IniciarServidorHTTP() no main.
func IniciarServidorHTTP() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Pong! Servidor está ativo.")
	})

	port := ":8000"
	fmt.Printf("Servidor HTTP iniciado em http://localhost%s/ping\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Erro ao iniciar servidor HTTP: %v\n", err)
	}
}
