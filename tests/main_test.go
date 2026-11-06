package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"proyecto-programacion-web/backend/db"
)

var testQueries *db.Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" { dbHost = "localhost" }
	dbUser := os.Getenv("TEST_DB_USER")
	dbPass := os.Getenv("TEST_DB_PASSWORD")
	dbName := os.Getenv("TEST_DB_NAME")
	dbPort := os.Getenv("TEST_DB_PORT")

	if dbUser == "" { dbUser = "testuser" }
	if dbPass == "" { dbPass = "testpassword" }
	if dbName == "" { dbName = "notas_db_test" }
	if dbPort == "" { dbPort = "5433" }

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	
	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("No se pudo conectar a la DB de test: %v", err)
	}
	defer testDB.Close()

	// Esperar activamente con ping a que Postgres esté listo para recibir queries
	for i := 0; i < 10; i++ {
		err = testDB.Ping()
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("Error al hacer ping a la base de datos de test: %v", err)
	}

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
