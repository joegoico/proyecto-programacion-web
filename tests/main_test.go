package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"proyecto-programacion-web/backend/db"
)

var testQueries *db.Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	dbUser := os.Getenv("TEST_DB_USER")
	dbPass := os.Getenv("TEST_DB_PASSWORD")
	dbName := os.Getenv("TEST_DB_NAME")
	dbPort := os.Getenv("TEST_DB_PORT")

	if dbUser == "" { dbUser = "testuser" }
	if dbPass == "" { dbPass = "testpassword" }
	if dbName == "" { dbName = "notas_db_test" }
	if dbPort == "" { dbPort = "5433" }

	connStr := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", dbUser, dbPass, dbPort, dbName)
	
	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("No se pudo conectar a la DB de test: %v", err)
	}
	defer testDB.Close()

	testQueries = db.New(testDB)

	// Ejecuta todos los tests
	code := m.Run()
	os.Exit(code)
}

// clearDB borra todos los datos de las tablas.
// Debe llamarse al inicio de cada test individual.
func clearDB() {
	_, err := testDB.Exec("DELETE FROM nota_categoria")
	if err != nil {
		log.Fatalf("Error limpiando nota_categoria: %v", err)
	}
	_, err = testDB.Exec("DELETE FROM categoria")
	if err != nil {
		log.Fatalf("Error limpiando categoria: %v", err)
	}
	_, err = testDB.Exec("DELETE FROM nota")
	if err != nil {
		log.Fatalf("Error limpiando nota: %v", err)
	}
}
