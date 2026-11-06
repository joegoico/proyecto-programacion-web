package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/backend/static"
	"proyecto-programacion-web/views"
)

// NotaStore abstrae las operaciones necesarias sobre las entidades Nota y Categoria
type NotaStore interface {
	ListNotas(ctx context.Context) ([]db.Notum, error)
	GetNota(ctx context.Context, id int32) (db.Notum, error)
	CreateNota(ctx context.Context, arg db.CreateNotaParams) (db.Notum, error)
	UpdateNota(ctx context.Context, arg db.UpdateNotaParams) (db.Notum, error)
	DeleteNota(ctx context.Context, id int32) error

	ListCategorias(ctx context.Context) ([]db.Categorium, error)
	GetCategoria(ctx context.Context, id int32) (db.Categorium, error)
	CreateCategoria(ctx context.Context, arg db.CreateCategoriaParams) (db.Categorium, error)
	UpdateCategoria(ctx context.Context, arg db.UpdateCategoriaParams) (db.Categorium, error)
	DeleteCategoria(ctx context.Context, id int32) error

	ListCategoriasByNota(ctx context.Context, idNota int32) ([]db.Categorium, error)
	AssignCategoriaToNota(ctx context.Context, arg db.AssignCategoriaToNotaParams) error
	RemoveCategoriaFromNota(ctx context.Context, arg db.RemoveCategoriaFromNotaParams) error
}

// WebHandler maneja el renderizado de vistas y formularios clásicos
type WebHandler struct {
	Store NotaStore
}

// NewWebHandler crea una nueva instancia de WebHandler
func NewWebHandler(store NotaStore) *WebHandler {
	return &WebHandler{Store: store}
}

// Styles sirve la hoja de estilos embebida.
func Styles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(static.Styles)
}

// HandleIndex renderiza la página principal. El estado de los formularios se
// controla con la URL: ?nota=nueva | ?nota=ID y ?categoria=nueva | ?categoria=ID.
func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()

	notas, err := h.Store.ListNotas(ctx)
	if err != nil {
		http.Error(w, "Error al obtener las notas", http.StatusInternalServerError)
		return
	}
	categorias, err := h.Store.ListCategorias(ctx)
	if err != nil {
		http.Error(w, "Error al obtener las categorías", http.StatusInternalServerError)
		return
	}
	porNota := make(map[int32][]db.Categorium, len(notas))
	for _, n := range notas {
		cats, err := h.Store.ListCategoriasByNota(ctx, n.IDNota)
		if err != nil {
			http.Error(w, "Error al obtener las categorías de las notas", http.StatusInternalServerError)
			return
		}
		porNota[n.IDNota] = cats
	}

	data := views.IndexData{Notas: notas, Categorias: categorias, CategoriasPorNota: porNota}

	switch p := r.URL.Query().Get("nota"); {
	case p == "nueva":
		data.NotaAbierta = true
	case p != "":
		if id, err := strconv.Atoi(p); err == nil {
			for i := range notas {
				if notas[i].IDNota == int32(id) {
					data.NotaAbierta, data.NotaEditando = true, &notas[i]
				}
			}
		}
	}
	switch p := r.URL.Query().Get("categoria"); {
	case p == "nueva":
		data.CategoriaAbierta = true
	case p != "":
		if id, err := strconv.Atoi(p); err == nil {
			for i := range categorias {
				if categorias[i].IDCategoria == int32(id) {
					data.CategoriaAbierta, data.CategoriaEditando = true, &categorias[i]
				}
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.IndexPage(data).Render(ctx, w); err != nil {
		http.Error(w, "Error al renderizar la vista", http.StatusInternalServerError)
	}
}

// parseID lee el parámetro {id} del path. Responde 400 y devuelve false si es inválido.
func parseID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return 0, false
	}
	return int32(id), true
}

// writeStoreError traduce errores de la base a 404 (no existe) o 500.
func writeStoreError(w http.ResponseWriter, err error, notFoundMsg, internalMsg string) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, notFoundMsg, http.StatusNotFound)
		return
	}
	http.Error(w, internalMsg, http.StatusInternalServerError)
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// categoriasSeleccionadas lee los checkboxes "categorias" del formulario,
// descartando valores inválidos y duplicados.
func categoriasSeleccionadas(r *http.Request) []int32 {
	var ids []int32
	vistos := map[int32]bool{}
	for _, v := range r.Form["categorias"] {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || vistos[int32(n)] {
			continue
		}
		vistos[int32(n)] = true
		ids = append(ids, int32(n))
	}
	return ids
}

// HandleCreateNota procesa el POST del formulario de notas (PRG)
func (h *WebHandler) HandleCreateNota(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar el formulario", http.StatusBadRequest)
		return
	}
	titulo := strings.TrimSpace(r.FormValue("titulo"))
	if titulo == "" {
		http.Error(w, "El título es obligatorio", http.StatusBadRequest)
		return
	}

	nota, err := h.Store.CreateNota(r.Context(), db.CreateNotaParams{
		Titulo:     titulo,
		Contenido:  nullString(strings.TrimSpace(r.FormValue("contenido"))),
		IsArchived: false,
	})
	if err != nil {
		http.Error(w, "Error al guardar la nota", http.StatusInternalServerError)
		return
	}

	for _, catID := range categoriasSeleccionadas(r) {
		if err := h.Store.AssignCategoriaToNota(r.Context(), db.AssignCategoriaToNotaParams{IDNota: nota.IDNota, IDCategoria: catID}); err != nil {
			http.Error(w, "Error al asignar la categoría", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleUpdateNota procesa la edición de una nota (PRG). Conserva is_archived y
// reemplaza sus categorías por las marcadas en el formulario.
func (h *WebHandler) HandleUpdateNota(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar el formulario", http.StatusBadRequest)
		return
	}
	titulo := strings.TrimSpace(r.FormValue("titulo"))
	if titulo == "" {
		http.Error(w, "El título es obligatorio", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	existente, err := h.Store.GetNota(ctx, id)
	if err != nil {
		writeStoreError(w, err, "Nota no encontrada", "Error al obtener la nota")
		return
	}
	if _, err := h.Store.UpdateNota(ctx, db.UpdateNotaParams{
		IDNota:     id,
		Titulo:     titulo,
		Contenido:  nullString(strings.TrimSpace(r.FormValue("contenido"))),
		IsArchived: existente.IsArchived,
	}); err != nil {
		writeStoreError(w, err, "Nota no encontrada", "Error al actualizar la nota")
		return
	}

	// Sincronizar categorías: quitar las desmarcadas, asignar las nuevas
	actuales, err := h.Store.ListCategoriasByNota(ctx, id)
	if err != nil {
		http.Error(w, "Error al obtener las categorías de la nota", http.StatusInternalServerError)
		return
	}
	deseadas := map[int32]bool{}
	for _, c := range categoriasSeleccionadas(r) {
		deseadas[c] = true
	}
	tiene := map[int32]bool{}
	for _, c := range actuales {
		tiene[c.IDCategoria] = true
		if !deseadas[c.IDCategoria] {
			if err := h.Store.RemoveCategoriaFromNota(ctx, db.RemoveCategoriaFromNotaParams{IDNota: id, IDCategoria: c.IDCategoria}); err != nil {
				http.Error(w, "Error al quitar la categoría", http.StatusInternalServerError)
				return
			}
		}
	}
	for _, c := range categoriasSeleccionadas(r) {
		if !tiene[c] {
			if err := h.Store.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{IDNota: id, IDCategoria: c}); err != nil {
				http.Error(w, "Error al asignar la categoría", http.StatusInternalServerError)
				return
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleDeleteNota elimina una nota (PRG)
func (h *WebHandler) HandleDeleteNota(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.Store.DeleteNota(r.Context(), id); err != nil {
		http.Error(w, "Error al eliminar la nota", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleCreateCategoria procesa el POST de nueva categoría (PRG)
func (h *WebHandler) HandleCreateCategoria(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar el formulario", http.StatusBadRequest)
		return
	}
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	if nombre == "" {
		http.Error(w, "El nombre de la categoría es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := h.Store.CreateCategoria(r.Context(), db.CreateCategoriaParams{
		Nombre:      nombre,
		Descripcion: nullString(strings.TrimSpace(r.FormValue("descripcion"))),
	}); err != nil {
		http.Error(w, "Error al guardar la categoría", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleUpdateCategoria procesa la edición de una categoría (PRG).
// Si el formulario no envía "descripcion", conserva la existente.
func (h *WebHandler) HandleUpdateCategoria(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar el formulario", http.StatusBadRequest)
		return
	}
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	if nombre == "" {
		http.Error(w, "El nombre de la categoría es obligatorio", http.StatusBadRequest)
		return
	}
	existente, err := h.Store.GetCategoria(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "Categoría no encontrada", "Error al obtener la categoría")
		return
	}
	descripcion := existente.Descripcion
	if r.Form.Has("descripcion") {
		descripcion = nullString(strings.TrimSpace(r.FormValue("descripcion")))
	}
	if _, err := h.Store.UpdateCategoria(r.Context(), db.UpdateCategoriaParams{
		IDCategoria: id, Nombre: nombre, Descripcion: descripcion,
	}); err != nil {
		writeStoreError(w, err, "Categoría no encontrada", "Error al actualizar la categoría")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleDeleteCategoria elimina una categoría (PRG)
func (h *WebHandler) HandleDeleteCategoria(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.Store.DeleteCategoria(r.Context(), id); err != nil {
		http.Error(w, "Error al eliminar la categoría", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
