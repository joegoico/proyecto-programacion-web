package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/backend/handlers"
)

func setupCategoriaRouter() (*http.ServeMux, *handlers.CategoriaHandler) {
	h := &handlers.CategoriaHandler{Queries: testQueries}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/categorias", h.List)
	mux.HandleFunc("GET /api/categorias/{id}", h.Get)
	mux.HandleFunc("POST /api/categorias", h.Create)
	mux.HandleFunc("PUT /api/categorias/{id}", h.Update)
	mux.HandleFunc("DELETE /api/categorias/{id}", h.Delete)
	return mux, h
}

func TestAPICreateCategoria(t *testing.T) {
	clearDB()
	mux, _ := setupCategoriaRouter()

	payload := []byte(`{"nombre":"Personal","descripcion":"Cosas personales"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/categorias", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Se esperaba estado 201, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var res db.Categorium
	json.NewDecoder(w.Body).Decode(&res)
	if res.Nombre != "Personal" {
		t.Errorf("Se esperaba nombre 'Personal', se obtuvo '%s'", res.Nombre)
	}
}

func TestAPIListCategorias(t *testing.T) {
	clearDB()
	mux, _ := setupCategoriaRouter()

	// Crear data dummy
	testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Cat A"})
	testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Cat B"})

	req := httptest.NewRequest(http.MethodGet, "/api/categorias", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d", w.Code)
	}

	var res []db.Categorium
	json.NewDecoder(w.Body).Decode(&res)
	if len(res) != 2 {
		t.Errorf("Se esperaban 2 categorias, se obtuvieron %d", len(res))
	}
}

func TestAPIGetCategoria(t *testing.T) {
	clearDB()
	mux, _ := setupCategoriaRouter()

	cat, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Unica"})

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/categorias/%d", cat.IDCategoria), nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d", w.Code)
	}

	var res db.Categorium
	json.NewDecoder(w.Body).Decode(&res)
	if res.Nombre != "Unica" {
		t.Errorf("Se esperaba 'Unica', se obtuvo '%s'", res.Nombre)
	}
}

func TestAPIUpdateCategoria(t *testing.T) {
	clearDB()
	mux, _ := setupCategoriaRouter()

	cat, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Vieja"})

	payload := []byte(`{"nombre":"Actualizada","descripcion":"Desc"}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categorias/%d", cat.IDCategoria), bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var res db.Categorium
	json.NewDecoder(w.Body).Decode(&res)
	if res.Nombre != "Actualizada" {
		t.Errorf("Se esperaba 'Actualizada', se obtuvo '%s'", res.Nombre)
	}
}

func TestAPIDeleteCategoria(t *testing.T) {
	clearDB()
	mux, _ := setupCategoriaRouter()

	cat, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Borrar"})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/categorias/%d", cat.IDCategoria), nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Se esperaba estado 204, se obtuvo %d", w.Code)
	}

	// Comprobar que realmente se borró
	_, err := testQueries.GetCategoria(context.Background(), cat.IDCategoria)
	if err == nil {
		t.Errorf("La categoría no fue eliminada de la base de datos")
	}
}
