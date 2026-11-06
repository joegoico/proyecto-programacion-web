package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/backend/handlers"
)

func setupWebRouter() *http.ServeMux {
	webHandler := handlers.NewWebHandler(testQueries)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /styles.css", handlers.Styles)
	mux.HandleFunc("GET /", webHandler.HandleIndex)
	mux.HandleFunc("POST /notas", webHandler.HandleCreateNota)
	mux.HandleFunc("POST /notas/{id}/editar", webHandler.HandleUpdateNota)
	mux.HandleFunc("POST /notas/{id}/eliminar", webHandler.HandleDeleteNota)
	mux.HandleFunc("POST /categorias", webHandler.HandleCreateCategoria)
	mux.HandleFunc("POST /categorias/{id}/editar", webHandler.HandleUpdateCategoria)
	mux.HandleFunc("POST /categorias/{id}/eliminar", webHandler.HandleDeleteCategoria)
	return mux
}

func postForm(mux *http.ServeMux, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func assertSeeOther(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusSeeOther {
		t.Fatalf("se esperaba 303 (See Other), se obtuvo %d. Body: %s", w.Code, w.Body.String())
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("se esperaba redirección a '/', se obtuvo %q", loc)
	}
}

func catIDs(t *testing.T, idNota int32) map[int32]bool {
	t.Helper()
	cats, err := testQueries.ListCategoriasByNota(context.Background(), idNota)
	if err != nil {
		t.Fatalf("error consultando categorías de la nota: %v", err)
	}
	ids := map[int32]bool{}
	for _, c := range cats {
		ids[c.IDCategoria] = true
	}
	return ids
}

func TestWeb_Integration_HandleIndex(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	ctx := context.Background()
	nota, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota Integración SSR"})
	cat, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Chip SSR"})
	if err := testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: nota.IDNota, IDCategoria: cat.IDCategoria}); err != nil {
		t.Fatalf("error asignando categoría: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}
	body := w.Body.String()
	for _, sub := range []string{"Nota Integración SSR", `<span class="chip">Chip SSR</span>`, `href="/styles.css"`} {
		if !strings.Contains(body, sub) {
			t.Errorf("se esperaba encontrar %q en el HTML renderizado", sub)
		}
	}
}

func TestWeb_Integration_EditarNotaPrecargaFormulario(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	nota, _ := testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "Para editar"})
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?nota=%d", nota.IDNota), nil)
	w := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(w, req)

	// Assert
	body := w.Body.String()
	for _, sub := range []string{fmt.Sprintf(`action="/notas/%d/editar"`, nota.IDNota), `value="Para editar"`} {
		if !strings.Contains(body, sub) {
			t.Errorf("se esperaba encontrar %q en el HTML", sub)
		}
	}
}

func TestWeb_Integration_Styles(t *testing.T) {
	// Arrange
	mux := setupWebRouter()
	req := httptest.NewRequest(http.MethodGet, "/styles.css", nil)
	w := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("se esperaba 200 text/css, se obtuvo %d %q", w.Code, w.Header().Get("Content-Type"))
	}
}

func TestWeb_Integration_CreateNotaConCategorias_PRG(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	ctx := context.Background()
	c1, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "A"})
	c2, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "B"})
	form := url.Values{
		"titulo":     {"Nota desde Formulario PRG"},
		"contenido":  {"Contenido guardado en Postgres"},
		"categorias": {fmt.Sprint(c1.IDCategoria), fmt.Sprint(c2.IDCategoria)},
	}

	// Act
	w := postForm(mux, "/notas", form)

	// Assert
	assertSeeOther(t, w)
	notas, err := testQueries.ListNotas(ctx)
	if err != nil || len(notas) != 1 {
		t.Fatalf("se esperaba 1 nota en la BD, hubo %d (err=%v)", len(notas), err)
	}
	if notas[0].Titulo != "Nota desde Formulario PRG" || notas[0].Contenido.String != "Contenido guardado en Postgres" {
		t.Errorf("nota inesperada: %+v", notas[0])
	}
	ids := catIDs(t, notas[0].IDNota)
	if len(ids) != 2 || !ids[c1.IDCategoria] || !ids[c2.IDCategoria] {
		t.Errorf("se esperaban las categorías %d y %d, hubo %v", c1.IDCategoria, c2.IDCategoria, ids)
	}
}

func TestWeb_Integration_UpdateNota_ReemplazaCategorias(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	ctx := context.Background()
	nota, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Original"})
	c1, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "A"})
	c2, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "B"})
	c3, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "C"})
	for _, c := range []int32{c1.IDCategoria, c2.IDCategoria} {
		if err := testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: nota.IDNota, IDCategoria: c}); err != nil {
			t.Fatalf("error asignando categoría: %v", err)
		}
	}
	form := url.Values{
		"titulo":     {"Editada"},
		"contenido":  {"Nuevo contenido"},
		"categorias": {fmt.Sprint(c2.IDCategoria), fmt.Sprint(c3.IDCategoria)},
	}

	// Act
	w := postForm(mux, fmt.Sprintf("/notas/%d/editar", nota.IDNota), form)

	// Assert
	assertSeeOther(t, w)
	got, err := testQueries.GetNota(ctx, nota.IDNota)
	if err != nil {
		t.Fatalf("error consultando nota: %v", err)
	}
	if got.Titulo != "Editada" || got.Contenido.String != "Nuevo contenido" {
		t.Errorf("la nota no se actualizó: %+v", got)
	}
	ids := catIDs(t, nota.IDNota)
	if len(ids) != 2 || ids[c1.IDCategoria] || !ids[c2.IDCategoria] || !ids[c3.IDCategoria] {
		t.Errorf("se esperaban las categorías %d y %d, hubo %v", c2.IDCategoria, c3.IDCategoria, ids)
	}
}

func TestWeb_Integration_DeleteNota_PRG(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	nota, _ := testQueries.CreateNota(context.Background(), db.CreateNotaParams{Titulo: "A borrar"})

	// Act
	w := postForm(mux, fmt.Sprintf("/notas/%d/eliminar", nota.IDNota), url.Values{})

	// Assert
	assertSeeOther(t, w)
	if notas, _ := testQueries.ListNotas(context.Background()); len(notas) != 0 {
		t.Errorf("se esperaba 0 notas, hay %d", len(notas))
	}
}

func TestWeb_Integration_CreateCategoria_PRG(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()

	// Act
	w := postForm(mux, "/categorias", url.Values{"nombre": {"Categoria Integración"}})

	// Assert
	assertSeeOther(t, w)
	cats, err := testQueries.ListCategorias(context.Background())
	if err != nil || len(cats) != 1 || cats[0].Nombre != "Categoria Integración" {
		t.Fatalf("categoría inesperada: %+v (err=%v)", cats, err)
	}
}

func TestWeb_Integration_UpdateCategoria_PRG(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	cat, _ := testQueries.CreateCategoria(context.Background(), db.CreateCategoriaParams{Nombre: "Vieja"})

	// Act
	w := postForm(mux, fmt.Sprintf("/categorias/%d/editar", cat.IDCategoria), url.Values{"nombre": {"Renombrada"}})

	// Assert
	assertSeeOther(t, w)
	got, err := testQueries.GetCategoria(context.Background(), cat.IDCategoria)
	if err != nil || got.Nombre != "Renombrada" {
		t.Errorf("la categoría no se actualizó: %+v (err=%v)", got, err)
	}
}

func TestWeb_Integration_DeleteCategoria_QuitaRelacionConNotas(t *testing.T) {
	// Arrange
	clearDB()
	mux := setupWebRouter()
	ctx := context.Background()
	nota, _ := testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Con categoría"})
	cat, _ := testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Temporal"})
	if err := testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: nota.IDNota, IDCategoria: cat.IDCategoria}); err != nil {
		t.Fatalf("error asignando categoría: %v", err)
	}

	// Act
	w := postForm(mux, fmt.Sprintf("/categorias/%d/eliminar", cat.IDCategoria), url.Values{})

	// Assert
	assertSeeOther(t, w)
	if cats, _ := testQueries.ListCategorias(ctx); len(cats) != 0 {
		t.Errorf("se esperaba 0 categorías, hay %d", len(cats))
	}
	if _, err := testQueries.GetNota(ctx, nota.IDNota); err != nil {
		t.Errorf("la nota no debía borrarse al eliminar su categoría: %v", err)
	}
	if ids := catIDs(t, nota.IDNota); len(ids) != 0 {
		t.Errorf("la nota no debía conservar la relación: %v", ids)
	}
}
