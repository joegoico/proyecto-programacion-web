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

func setupNotaRouter() (*http.ServeMux, *handlers.NotaHandler) {
	h := &handlers.NotaHandler{Queries: testQueries}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/notas", h.List)
	mux.HandleFunc("GET /api/notas/{id}", h.Get)
	mux.HandleFunc("POST /api/notas", h.Create)
	mux.HandleFunc("PUT /api/notas/{id}", h.Update)
	mux.HandleFunc("DELETE /api/notas/{id}", h.Delete)
	return mux, h
}

func TestAPICreateNota(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()

	payload := []byte(`{"titulo":"Nota 1","contenido":"Mi primera nota","is_archived":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notas", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Se esperaba estado 201, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var res db.Notum
	json.NewDecoder(w.Body).Decode(&res)
	if res.Titulo != "Nota 1" {
		t.Errorf("Se esperaba titulo 'Nota 1', se obtuvo '%s'", res.Titulo)
	}
}

func TestAPIListNotas(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()

	testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "A", IsArchived: false})
	testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "B", IsArchived: false})

	req := httptest.NewRequest(http.MethodGet, "/api/notas", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d", w.Code)
	}

	var res []db.Notum
	json.NewDecoder(w.Body).Decode(&res)
	if len(res) != 2 {
		t.Errorf("Se esperaban 2 notas, se obtuvieron %d", len(res))
	}
}

func TestAPIGetNota(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()

	nota, _ := testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "Busqueda", IsArchived: false})

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/notas/%d", nota.IDNota), nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d", w.Code)
	}

	var res db.Notum
	json.NewDecoder(w.Body).Decode(&res)
	if res.Titulo != "Busqueda" {
		t.Errorf("Se esperaba 'Busqueda', se obtuvo '%s'", res.Titulo)
	}
}

func TestAPIUpdateNota(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()

	nota, _ := testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "Vieja", IsArchived: false})

	payload := []byte(`{"titulo":"Actualizada","contenido":"Desc","is_archived":true}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/notas/%d", nota.IDNota), bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba estado 200, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var res db.Notum
	json.NewDecoder(w.Body).Decode(&res)
	if res.Titulo != "Actualizada" {
		t.Errorf("Se esperaba 'Actualizada', se obtuvo '%s'", res.Titulo)
	}
	if res.IsArchived != true {
		t.Errorf("Se esperaba que estuviese archivada")
	}
}

func TestAPIDeleteNota(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()

	nota, _ := testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "Borrar", IsArchived: false})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/notas/%d", nota.IDNota), nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Se esperaba estado 204, se obtuvo %d", w.Code)
	}

	_, err := testQueries.GetNota(context.Background(), nota.IDNota)
	if err == nil {
		t.Errorf("La nota no fue eliminada de la base de datos")
	}
}
