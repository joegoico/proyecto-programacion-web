package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/backend/handlers"
)

// mockNotaStore implementa handlers.NotaStore registrando las llamadas recibidas.
type mockNotaStore struct {
	notas      []db.Notum
	categorias []db.Categorium
	catsByNota map[int32][]db.Categorium

	createCalls     []db.CreateNotaParams
	updateNotaCalls []db.UpdateNotaParams
	deleteNotaCalls []int32
	assignCalls     []db.AssignCategoriaToNotaParams
	removeCalls     []db.RemoveCategoriaFromNotaParams
	createCatCalls  []db.CreateCategoriaParams
	updateCatCalls  []db.UpdateCategoriaParams
	deleteCatCalls  []int32

	listErr           error
	listCatErr        error
	listCatsByNotaErr error
	createErr         error
	createCatErr      error
	assignErr         error
	updateErr         error
	deleteErr         error
}

func (m *mockNotaStore) ListNotas(ctx context.Context) ([]db.Notum, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.notas, nil
}

func (m *mockNotaStore) GetNota(ctx context.Context, id int32) (db.Notum, error) {
	for _, n := range m.notas {
		if n.IDNota == id {
			return n, nil
		}
	}
	return db.Notum{}, sql.ErrNoRows
}

func (m *mockNotaStore) CreateNota(ctx context.Context, arg db.CreateNotaParams) (db.Notum, error) {
	if m.createErr != nil {
		return db.Notum{}, m.createErr
	}
	m.createCalls = append(m.createCalls, arg)
	n := db.Notum{
		IDNota: int32(len(m.notas) + 1), Titulo: arg.Titulo, Contenido: arg.Contenido,
		IsArchived: arg.IsArchived, CreatedAt: time.Now(),
	}
	m.notas = append(m.notas, n)
	return n, nil
}

func (m *mockNotaStore) UpdateNota(ctx context.Context, arg db.UpdateNotaParams) (db.Notum, error) {
	if m.updateErr != nil {
		return db.Notum{}, m.updateErr
	}
	m.updateNotaCalls = append(m.updateNotaCalls, arg)
	return db.Notum{IDNota: arg.IDNota, Titulo: arg.Titulo, Contenido: arg.Contenido, IsArchived: arg.IsArchived}, nil
}

func (m *mockNotaStore) DeleteNota(ctx context.Context, id int32) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteNotaCalls = append(m.deleteNotaCalls, id)
	return nil
}

func (m *mockNotaStore) ListCategorias(ctx context.Context) ([]db.Categorium, error) {
	if m.listCatErr != nil {
		return nil, m.listCatErr
	}
	return m.categorias, nil
}

func (m *mockNotaStore) GetCategoria(ctx context.Context, id int32) (db.Categorium, error) {
	for _, c := range m.categorias {
		if c.IDCategoria == id {
			return c, nil
		}
	}
	return db.Categorium{}, sql.ErrNoRows
}

func (m *mockNotaStore) CreateCategoria(ctx context.Context, arg db.CreateCategoriaParams) (db.Categorium, error) {
	if m.createCatErr != nil {
		return db.Categorium{}, m.createCatErr
	}
	m.createCatCalls = append(m.createCatCalls, arg)
	c := db.Categorium{IDCategoria: int32(len(m.categorias) + 1), Nombre: arg.Nombre, Descripcion: arg.Descripcion}
	m.categorias = append(m.categorias, c)
	return c, nil
}

func (m *mockNotaStore) UpdateCategoria(ctx context.Context, arg db.UpdateCategoriaParams) (db.Categorium, error) {
	if m.updateErr != nil {
		return db.Categorium{}, m.updateErr
	}
	m.updateCatCalls = append(m.updateCatCalls, arg)
	return db.Categorium{IDCategoria: arg.IDCategoria, Nombre: arg.Nombre, Descripcion: arg.Descripcion}, nil
}

func (m *mockNotaStore) DeleteCategoria(ctx context.Context, id int32) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCatCalls = append(m.deleteCatCalls, id)
	return nil
}

func (m *mockNotaStore) ListCategoriasByNota(ctx context.Context, idNota int32) ([]db.Categorium, error) {
	if m.listCatsByNotaErr != nil {
		return nil, m.listCatsByNotaErr
	}
	return m.catsByNota[idNota], nil
}

func (m *mockNotaStore) AssignCategoriaToNota(ctx context.Context, arg db.AssignCategoriaToNotaParams) error {
	if m.assignErr != nil {
		return m.assignErr
	}
	m.assignCalls = append(m.assignCalls, arg)
	return nil
}

func (m *mockNotaStore) RemoveCategoriaFromNota(ctx context.Context, arg db.RemoveCategoriaFromNotaParams) error {
	m.removeCalls = append(m.removeCalls, arg)
	return nil
}

// --- helpers ---

func newPost(path, id string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if id != "" {
		req.SetPathValue("id", id)
	}
	return req
}

func assertRedirectHome(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusSeeOther {
		t.Fatalf("se esperaba 303 (See Other), se obtuvo %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("se esperaba redirección a '/', se obtuvo %q", loc)
	}
}

func get(h *handlers.WebHandler, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.HandleIndex(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func storeConDatos() *mockNotaStore {
	return &mockNotaStore{
		notas: []db.Notum{{
			IDNota: 1, Titulo: "Nota Categorizada",
			Contenido: sql.NullString{String: "Contenido AAA", Valid: true}, CreatedAt: time.Now(),
		}},
		categorias: []db.Categorium{{IDCategoria: 1, Nombre: "Trabajo"}, {IDCategoria: 2, Nombre: "Estudio"}},
		catsByNota: map[int32][]db.Categorium{1: {{IDCategoria: 1, Nombre: "Trabajo"}}},
	}
}

// --- GET / ---

func TestHandleIndex_Success(t *testing.T) {
	// Arrange
	h := handlers.NewWebHandler(storeConDatos())

	// Act
	w := get(h, "/")

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("se esperaba text/html, se obtuvo %s", ct)
	}
	body := w.Body.String()
	for _, sub := range []string{
		"Nota Categorizada", "Contenido AAA",
		`<span class="chip">Trabajo</span>`, `<li class="chip"><span>Estudio</span>`,
	} {
		if !strings.Contains(body, sub) {
			t.Errorf("el body no contiene %q: %s", sub, body)
		}
	}
	if strings.Contains(body, `id="nota-form"`) || strings.Contains(body, `id="categoria-form"`) {
		t.Errorf("los formularios deben estar cerrados por defecto")
	}
}

func TestHandleIndex_NotFoundForSubpaths(t *testing.T) {
	// Arrange
	h := handlers.NewWebHandler(&mockNotaStore{})

	// Act
	w := get(h, "/subruta-inexistente")

	// Assert
	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, se obtuvo %d", w.Code)
	}
}

func TestHandleIndex_ErroresDeStore(t *testing.T) {
	casos := map[string]*mockNotaStore{
		"notas":              {listErr: errors.New("boom")},
		"categorias":         {listCatErr: errors.New("boom")},
		"categorias de nota": {notas: []db.Notum{{IDNota: 1}}, listCatsByNotaErr: errors.New("boom")},
	}
	for nombre, store := range casos {
		t.Run(nombre, func(t *testing.T) {
			// Arrange
			h := handlers.NewWebHandler(store)

			// Act
			w := get(h, "/")

			// Assert
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("se esperaba 500, se obtuvo %d", w.Code)
			}
		})
	}
}

func TestHandleIndex_AbreFormularioDeNuevaNotaYCategoria(t *testing.T) {
	// Arrange
	h := handlers.NewWebHandler(storeConDatos())

	// Act
	w := get(h, "/?nota=nueva&categoria=nueva")

	// Assert
	body := w.Body.String()
	for _, sub := range []string{`id="nota-form"`, `action="/notas"`, `id="categoria-form"`, `action="/categorias"`} {
		if !strings.Contains(body, sub) {
			t.Errorf("el body no contiene %q: %s", sub, body)
		}
	}
}

func TestHandleIndex_EditarNotaPrecargaDatosYCategorias(t *testing.T) {
	// Arrange
	h := handlers.NewWebHandler(storeConDatos())

	// Act
	w := get(h, "/?nota=1")

	// Assert
	body := w.Body.String()
	for _, sub := range []string{
		`action="/notas/1/editar" method="POST"`, `value="Nota Categorizada"`, "Contenido AAA</textarea>", `value="1" checked`,
	} {
		if !strings.Contains(body, sub) {
			t.Errorf("el body no contiene %q: %s", sub, body)
		}
	}
	if strings.Contains(body, `value="2" checked`) {
		t.Errorf("solo la categoría 1 debía estar marcada")
	}
}

func TestHandleIndex_EditarCategoriaPrecargaNombre(t *testing.T) {
	// Arrange
	h := handlers.NewWebHandler(storeConDatos())

	// Act
	w := get(h, "/?categoria=2")

	// Assert
	body := w.Body.String()
	for _, sub := range []string{`action="/categorias/2/editar" method="POST"`, `value="Estudio"`} {
		if !strings.Contains(body, sub) {
			t.Errorf("el body no contiene %q: %s", sub, body)
		}
	}
}

func TestHandleIndex_EditarInexistenteOInvalidoNoAbreFormulario(t *testing.T) {
	for _, target := range []string{"/?nota=999", "/?nota=abc", "/?categoria=999", "/?categoria=abc"} {
		t.Run(target, func(t *testing.T) {
			// Arrange
			h := handlers.NewWebHandler(storeConDatos())

			// Act
			w := get(h, target)

			// Assert
			if w.Code != http.StatusOK {
				t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
			}
			body := w.Body.String()
			if strings.Contains(body, `id="nota-form"`) || strings.Contains(body, `id="categoria-form"`) {
				t.Errorf("no debía abrirse ningún formulario")
			}
		})
	}
}

// --- estilos ---

func TestStyles_ServeCSS(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/styles.css", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.Styles(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Errorf("se esperaba text/css, se obtuvo %s", ct)
	}
	if !strings.Contains(w.Body.String(), "box-sizing") {
		t.Errorf("el CSS servido no parece ser styles.css")
	}
}

// --- POST /notas ---

func TestHandleCreateNota_Success_PRG(t *testing.T) {
	// Arrange
	store := &mockNotaStore{}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas", "", url.Values{"titulo": {"Comprar leche"}, "contenido": {"Descremada 1L"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateNota(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.createCalls) != 1 {
		t.Fatalf("se esperaba 1 llamada a CreateNota, hubo %d", len(store.createCalls))
	}
	got := store.createCalls[0]
	if got.Titulo != "Comprar leche" || !got.Contenido.Valid || got.Contenido.String != "Descremada 1L" {
		t.Errorf("parámetros inesperados: %+v", got)
	}
	if len(store.assignCalls) != 0 {
		t.Errorf("sin categorías seleccionadas no debía asignarse ninguna")
	}
}

func TestHandleCreateNota_EmptyTitle(t *testing.T) {
	// Arrange
	store := &mockNotaStore{}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas", "", url.Values{"titulo": {"   "}, "contenido": {"x"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateNota(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400, se obtuvo %d", w.Code)
	}
	if len(store.createCalls) != 0 {
		t.Errorf("no debía llamarse a CreateNota con título vacío")
	}
}

func TestHandleCreateNota_ConVariasCategorias(t *testing.T) {
	// Arrange
	store := &mockNotaStore{}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas", "", url.Values{"titulo": {"TP5"}, "categorias": {"1", "2", "abc", "0"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateNota(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.assignCalls) != 2 {
		t.Fatalf("se esperaban 2 asignaciones (los valores inválidos se ignoran), hubo %d", len(store.assignCalls))
	}
	for i, wantCat := range []int32{1, 2} {
		got := store.assignCalls[i]
		if got.IDNota != 1 || got.IDCategoria != wantCat {
			t.Errorf("asignación %d inesperada: %+v", i, got)
		}
	}
}

func TestHandleCreateNota_ErrorAlAsignarCategoria(t *testing.T) {
	// Arrange
	store := &mockNotaStore{assignErr: errors.New("boom")}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas", "", url.Values{"titulo": {"TP5"}, "categorias": {"1"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateNota(w, req)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, se obtuvo %d", w.Code)
	}
}

// --- POST /notas/{id}/editar y /eliminar ---

func TestHandleUpdateNota_Success_PRG(t *testing.T) {
	// Arrange
	store := &mockNotaStore{notas: []db.Notum{{IDNota: 5, Titulo: "Viejo", IsArchived: true}}}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas/5/editar", "5", url.Values{"titulo": {"  Nuevo titulo "}, "contenido": {"Nuevo contenido"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleUpdateNota(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.updateNotaCalls) != 1 {
		t.Fatalf("se esperaba 1 llamada a UpdateNota, hubo %d", len(store.updateNotaCalls))
	}
	got := store.updateNotaCalls[0]
	if got.IDNota != 5 || got.Titulo != "Nuevo titulo" || got.Contenido.String != "Nuevo contenido" || !got.Contenido.Valid {
		t.Errorf("parámetros inesperados: %+v", got)
	}
	if !got.IsArchived {
		t.Errorf("debía conservarse is_archived de la nota existente")
	}
}

func TestHandleUpdateNota_ReemplazaCategorias(t *testing.T) {
	// Arrange: la nota tiene {1,2}; se envían {2,3}
	store := &mockNotaStore{
		notas:      []db.Notum{{IDNota: 5, Titulo: "Viejo"}},
		catsByNota: map[int32][]db.Categorium{5: {{IDCategoria: 1}, {IDCategoria: 2}}},
	}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas/5/editar", "5", url.Values{"titulo": {"Viejo"}, "categorias": {"2", "3"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleUpdateNota(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.removeCalls) != 1 || store.removeCalls[0].IDCategoria != 1 || store.removeCalls[0].IDNota != 5 {
		t.Errorf("se esperaba quitar solo la categoría 1, hubo %+v", store.removeCalls)
	}
	if len(store.assignCalls) != 1 || store.assignCalls[0].IDCategoria != 3 || store.assignCalls[0].IDNota != 5 {
		t.Errorf("se esperaba asignar solo la categoría 3, hubo %+v", store.assignCalls)
	}
}

func TestHandleUpdateNota_SinCategoriasMarcadasLasQuitaTodas(t *testing.T) {
	// Arrange
	store := &mockNotaStore{
		notas:      []db.Notum{{IDNota: 5, Titulo: "Viejo"}},
		catsByNota: map[int32][]db.Categorium{5: {{IDCategoria: 1}, {IDCategoria: 2}}},
	}
	h := handlers.NewWebHandler(store)
	req := newPost("/notas/5/editar", "5", url.Values{"titulo": {"Viejo"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleUpdateNota(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.removeCalls) != 2 || len(store.assignCalls) != 0 {
		t.Errorf("se esperaba quitar 2 y asignar 0, hubo remove=%v assign=%v", store.removeCalls, store.assignCalls)
	}
}

func TestHandleUpdateNota_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		store  *mockNotaStore
		id     string
		form   url.Values
		status int
	}{
		{"titulo vacío", &mockNotaStore{notas: []db.Notum{{IDNota: 5}}}, "5", url.Values{"titulo": {" "}}, http.StatusBadRequest},
		{"id inválido", &mockNotaStore{}, "abc", url.Values{"titulo": {"X"}}, http.StatusBadRequest},
		{"no existe", &mockNotaStore{}, "99", url.Values{"titulo": {"X"}}, http.StatusNotFound},
		{"falla la base", &mockNotaStore{notas: []db.Notum{{IDNota: 5}}, updateErr: errors.New("boom")}, "5", url.Values{"titulo": {"X"}}, http.StatusInternalServerError},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			// Arrange
			h := handlers.NewWebHandler(c.store)
			req := newPost("/notas/"+c.id+"/editar", c.id, c.form)
			w := httptest.NewRecorder()

			// Act
			h.HandleUpdateNota(w, req)

			// Assert
			if w.Code != c.status {
				t.Fatalf("se esperaba %d, se obtuvo %d", c.status, w.Code)
			}
			if len(c.store.updateNotaCalls) != 0 {
				t.Errorf("no debía actualizarse la nota")
			}
		})
	}
}

func TestHandleDeleteNota(t *testing.T) {
	t.Run("ok redirige", func(t *testing.T) {
		// Arrange
		store := &mockNotaStore{}
		h := handlers.NewWebHandler(store)
		w := httptest.NewRecorder()

		// Act
		h.HandleDeleteNota(w, newPost("/notas/7/eliminar", "7", url.Values{}))

		// Assert
		assertRedirectHome(t, w)
		if len(store.deleteNotaCalls) != 1 || store.deleteNotaCalls[0] != 7 {
			t.Errorf("se esperaba DeleteNota(7), hubo %v", store.deleteNotaCalls)
		}
	})
	t.Run("id inválido", func(t *testing.T) {
		// Arrange
		store := &mockNotaStore{}
		h := handlers.NewWebHandler(store)
		w := httptest.NewRecorder()

		// Act
		h.HandleDeleteNota(w, newPost("/notas/abc/eliminar", "abc", url.Values{}))

		// Assert
		if w.Code != http.StatusBadRequest || len(store.deleteNotaCalls) != 0 {
			t.Errorf("se esperaba 400 sin borrar, se obtuvo %d y %v", w.Code, store.deleteNotaCalls)
		}
	})
	t.Run("falla la base", func(t *testing.T) {
		// Arrange
		h := handlers.NewWebHandler(&mockNotaStore{deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()

		// Act
		h.HandleDeleteNota(w, newPost("/notas/7/eliminar", "7", url.Values{}))

		// Assert
		if w.Code != http.StatusInternalServerError {
			t.Errorf("se esperaba 500, se obtuvo %d", w.Code)
		}
	})
}

// --- categorías ---

func TestHandleCreateCategoria_Success_PRG(t *testing.T) {
	// Arrange
	store := &mockNotaStore{}
	h := handlers.NewWebHandler(store)
	req := newPost("/categorias", "", url.Values{"nombre": {"Facultad"}})
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateCategoria(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.createCatCalls) != 1 || store.createCatCalls[0].Nombre != "Facultad" {
		t.Fatalf("parámetros inesperados: %+v", store.createCatCalls)
	}
	if store.createCatCalls[0].Descripcion.Valid {
		t.Errorf("sin descripción enviada no debía guardarse una")
	}
}

func TestHandleCreateCategoria_EmptyName(t *testing.T) {
	// Arrange
	store := &mockNotaStore{}
	h := handlers.NewWebHandler(store)
	w := httptest.NewRecorder()

	// Act
	h.HandleCreateCategoria(w, newPost("/categorias", "", url.Values{"nombre": {"   "}}))

	// Assert
	if w.Code != http.StatusBadRequest || len(store.createCatCalls) != 0 {
		t.Errorf("se esperaba 400 sin crear, se obtuvo %d y %d llamadas", w.Code, len(store.createCatCalls))
	}
}

func TestHandleUpdateCategoria_Success_PRG(t *testing.T) {
	// Arrange
	store := &mockNotaStore{categorias: []db.Categorium{{IDCategoria: 4, Nombre: "Vieja", Descripcion: sql.NullString{String: "Desc previa", Valid: true}}}}
	h := handlers.NewWebHandler(store)
	req := newPost("/categorias/4/editar", "4", url.Values{"nombre": {" Nueva "}})
	w := httptest.NewRecorder()

	// Act
	h.HandleUpdateCategoria(w, req)

	// Assert
	assertRedirectHome(t, w)
	if len(store.updateCatCalls) != 1 {
		t.Fatalf("se esperaba 1 llamada a UpdateCategoria, hubo %d", len(store.updateCatCalls))
	}
	got := store.updateCatCalls[0]
	if got.IDCategoria != 4 || got.Nombre != "Nueva" {
		t.Errorf("parámetros inesperados: %+v", got)
	}
	if !got.Descripcion.Valid || got.Descripcion.String != "Desc previa" {
		t.Errorf("debía conservarse la descripción existente: %+v", got.Descripcion)
	}
}

func TestHandleUpdateCategoria_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		store  *mockNotaStore
		id     string
		form   url.Values
		status int
	}{
		{"nombre vacío", &mockNotaStore{categorias: []db.Categorium{{IDCategoria: 4}}}, "4", url.Values{"nombre": {" "}}, http.StatusBadRequest},
		{"id inválido", &mockNotaStore{}, "x", url.Values{"nombre": {"X"}}, http.StatusBadRequest},
		{"no existe", &mockNotaStore{}, "99", url.Values{"nombre": {"X"}}, http.StatusNotFound},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			// Arrange
			h := handlers.NewWebHandler(c.store)
			w := httptest.NewRecorder()

			// Act
			h.HandleUpdateCategoria(w, newPost("/categorias/"+c.id+"/editar", c.id, c.form))

			// Assert
			if w.Code != c.status || len(c.store.updateCatCalls) != 0 {
				t.Errorf("se esperaba %d sin actualizar, se obtuvo %d", c.status, w.Code)
			}
		})
	}
}

func TestHandleDeleteCategoria(t *testing.T) {
	t.Run("ok redirige", func(t *testing.T) {
		// Arrange
		store := &mockNotaStore{}
		h := handlers.NewWebHandler(store)
		w := httptest.NewRecorder()

		// Act
		h.HandleDeleteCategoria(w, newPost("/categorias/3/eliminar", "3", url.Values{}))

		// Assert
		assertRedirectHome(t, w)
		if len(store.deleteCatCalls) != 1 || store.deleteCatCalls[0] != 3 {
			t.Errorf("se esperaba DeleteCategoria(3), hubo %v", store.deleteCatCalls)
		}
	})
	t.Run("id inválido", func(t *testing.T) {
		// Arrange
		h := handlers.NewWebHandler(&mockNotaStore{})
		w := httptest.NewRecorder()

		// Act
		h.HandleDeleteCategoria(w, newPost("/categorias/x/eliminar", "x", url.Values{}))

		// Assert
		if w.Code != http.StatusBadRequest {
			t.Errorf("se esperaba 400, se obtuvo %d", w.Code)
		}
	})
}
