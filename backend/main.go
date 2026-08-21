package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	// 2. Registra un manejador (handler) para la ruta raíz "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		content, err := os.ReadFile("frontend/index.html")
		if err != nil {
			http.Error(w, "Error al cargar el archivo HTML", http.StatusInternalServerError)
			return
		}

		// 2. Establecemos explícitamente el header Content-Type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 3. Enviamos el contenido
		w.Write(content)
	})
	port := "8080"
	fmt.Printf("Servidor escuchando en http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}