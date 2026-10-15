package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/backend/handlers"

	_ "github.com/lib/pq"
)

func main() {
	// 1. Conexión con la Base de Datos
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/notas_db?sslmode=disable"
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error al conectarse a la base de datos: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Error al hacer ping a la base de datos: %v", err)
	}

	// Instanciar repositorio y handlers
	queries := db.New(dbConn)
	categoriaHandler := &handlers.CategoriaHandler{Queries: queries}
	notaHandler := &handlers.NotaHandler{Queries: queries}

	mux := http.NewServeMux()

	// Handler para el frontend
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		content, err := os.ReadFile("frontend/index.html")
		if err != nil {
			http.Error(w, "Error al cargar el archivo HTML", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	// Rutas API para Categorías
	mux.HandleFunc("GET /api/categorias", categoriaHandler.List)
	mux.HandleFunc("GET /api/categorias/{id}", categoriaHandler.Get)
	mux.HandleFunc("POST /api/categorias", categoriaHandler.Create)
	mux.HandleFunc("PUT /api/categorias/{id}", categoriaHandler.Update)
	mux.HandleFunc("DELETE /api/categorias/{id}", categoriaHandler.Delete)

	// Rutas API para Notas
	mux.HandleFunc("GET /api/notas", notaHandler.List)
	mux.HandleFunc("GET /api/notas/{id}", notaHandler.Get)
	mux.HandleFunc("POST /api/notas", notaHandler.Create)
	mux.HandleFunc("PUT /api/notas/{id}", notaHandler.Update)
	mux.HandleFunc("DELETE /api/notas/{id}", notaHandler.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor escuchando en http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
