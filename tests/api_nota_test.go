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

func TestAPICreateNotaConCategoriasYListar(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()
	c1, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Trabajo"})
	c2, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Casa"})

	payload := []byte(fmt.Sprintf(`{"titulo":"N","categorias":[%d,%d]}`, c1.IDCategoria, c2.IDCategoria))
	req := httptest.NewRequest(http.MethodPost, "/api/notas", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Se esperaba 201, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/notas", nil))
	var res []struct {
		Titulo     string `json:"titulo"`
		Categorias []struct {
			Nombre string `json:"nombre"`
		} `json:"categorias"`
	}
	json.NewDecoder(w.Body).Decode(&res)
	if len(res) != 1 || len(res[0].Categorias) != 2 {
		t.Fatalf("Se esperaba 1 nota con 2 categorías, se obtuvo %+v", res)
	}
	if res[0].Categorias[0].Nombre != "Casa" || res[0].Categorias[1].Nombre != "Trabajo" {
		t.Errorf("Categorías inesperadas: %+v", res[0].Categorias)
	}
}

func TestAPIListNotasSinCategoriasDevuelveArrayVacio(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()
	testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "A"})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/notas", nil))
	if !bytes.Contains(w.Body.Bytes(), []byte(`"categorias":[]`)) {
		t.Errorf("Se esperaba categorias vacío, body: %s", w.Body.String())
	}
}

func TestAPIUpdateNotaReemplazaCategorias(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()
	ctx := context.Background()
	c1, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "A"})
	c2, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "B"})
	n, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "N"})
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: n.IDNota, IDCategoria: c1.IDCategoria})

	payload := []byte(fmt.Sprintf(`{"titulo":"N2","categorias":[%d]}`, c2.IDCategoria))
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/notas/%d", n.IDNota), bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Se esperaba 200, se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}
	cats, _ := testQueries.ListCategoriasByNota(ctx, n.IDNota)
	if len(cats) != 1 || cats[0].Nombre != "B" {
		t.Errorf("Se esperaba solo la categoría B, se obtuvo %+v", cats)
	}
}

func TestAPIUpdateNotaSinCampoCategoriasLasConserva(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()
	ctx := context.Background()
	c1, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "A"})
	n, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "N"})
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: n.IDNota, IDCategoria: c1.IDCategoria})

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/notas/%d", n.IDNota), bytes.NewBufferString(`{"titulo":"N2"}`))
	mux.ServeHTTP(httptest.NewRecorder(), req)
	cats, _ := testQueries.ListCategoriasByNota(ctx, n.IDNota)
	if len(cats) != 1 {
		t.Errorf("Se esperaba conservar 1 categoría, se obtuvo %d", len(cats))
	}
}

func TestAPIUpdateNotaConListaVaciaQuitaCategorias(t *testing.T) {
	clearDB()
	mux, _ := setupNotaRouter()
	ctx := context.Background()
	c1, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "A"})
	n, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "N"})
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: n.IDNota, IDCategoria: c1.IDCategoria})

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/notas/%d", n.IDNota), bytes.NewBufferString(`{"titulo":"N2","categorias":[]}`))
	mux.ServeHTTP(httptest.NewRecorder(), req)
	cats, _ := testQueries.ListCategoriasByNota(ctx, n.IDNota)
	if len(cats) != 0 {
		t.Errorf("Se esperaban 0 categorías, se obtuvo %d", len(cats))
	}
}
