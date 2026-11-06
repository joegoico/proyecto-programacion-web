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
	webHandler := handlers.NewWebHandler(queries)

	mux := http.NewServeMux()

	// Hoja de estilos embebida
	mux.HandleFunc("GET /styles.css", handlers.Styles)

	// Rutas web SSR y formularios clásicos (Patrón PRG)
	mux.HandleFunc("GET /", webHandler.HandleIndex)
	mux.HandleFunc("POST /notas", webHandler.HandleCreateNota)
	mux.HandleFunc("POST /notas/{id}/editar", webHandler.HandleUpdateNota)
	mux.HandleFunc("POST /notas/{id}/eliminar", webHandler.HandleDeleteNota)
	mux.HandleFunc("POST /categorias", webHandler.HandleCreateCategoria)
	mux.HandleFunc("POST /categorias/{id}/editar", webHandler.HandleUpdateCategoria)
	mux.HandleFunc("POST /categorias/{id}/eliminar", webHandler.HandleDeleteCategoria)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor escuchando en http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
