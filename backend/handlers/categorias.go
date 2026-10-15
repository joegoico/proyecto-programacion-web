package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"proyecto-programacion-web/backend/db"
)

type CategoriaHandler struct {
	Queries *db.Queries
}

func (h *CategoriaHandler) List(w http.ResponseWriter, r *http.Request) {
	categorias, err := h.Queries.ListCategorias(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categorias)
}

func (h *CategoriaHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	categoria, err := h.Queries.GetCategoria(r.Context(), int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categoria)
}

func (h *CategoriaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	if input.Nombre == "" {
		http.Error(w, "El campo 'nombre' no puede estar vacío", http.StatusBadRequest)
		return
	}

	params := db.CreateCategoriaParams{
		Nombre: input.Nombre,
	}
	if input.Descripcion != "" {
		params.Descripcion = sql.NullString{String: input.Descripcion, Valid: true}
	}

	categoria, err := h.Queries.CreateCategoria(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(categoria)
}

func (h *CategoriaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	if input.Nombre == "" {
		http.Error(w, "El campo 'nombre' no puede estar vacío", http.StatusBadRequest)
		return
	}

	params := db.UpdateCategoriaParams{
		IDCategoria: int32(id),
		Nombre:      input.Nombre,
	}
	if input.Descripcion != "" {
		params.Descripcion = sql.NullString{String: input.Descripcion, Valid: true}
	}

	categoria, err := h.Queries.UpdateCategoria(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categoria)
}

func (h *CategoriaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	err = h.Queries.DeleteCategoria(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
