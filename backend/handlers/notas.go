package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"proyecto-programacion-web/backend/db"
)

type NotaHandler struct {
	Queries *db.Queries
}

func (h *NotaHandler) List(w http.ResponseWriter, r *http.Request) {
	notas, err := h.Queries.ListNotas(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notas)
}

func (h *NotaHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	nota, err := h.Queries.GetNota(r.Context(), int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nota no encontrada", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nota)
}

func (h *NotaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Titulo     string `json:"titulo"`
		Contenido  string `json:"contenido"`
		IsArchived bool   `json:"is_archived"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	if input.Titulo == "" {
		http.Error(w, "El campo 'titulo' no puede estar vacío", http.StatusBadRequest)
		return
	}

	params := db.CreateNotaParams{
		Titulo:     input.Titulo,
		IsArchived: input.IsArchived,
	}
	if input.Contenido != "" {
		params.Contenido = sql.NullString{String: input.Contenido, Valid: true}
	}

	nota, err := h.Queries.CreateNota(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nota)
}

func (h *NotaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Titulo     string `json:"titulo"`
		Contenido  string `json:"contenido"`
		IsArchived bool   `json:"is_archived"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Cuerpo de la petición inválido", http.StatusBadRequest)
		return
	}

	if input.Titulo == "" {
		http.Error(w, "El campo 'titulo' no puede estar vacío", http.StatusBadRequest)
		return
	}

	params := db.UpdateNotaParams{
		IDNota:     int32(id),
		Titulo:     input.Titulo,
		IsArchived: input.IsArchived,
	}
	if input.Contenido != "" {
		params.Contenido = sql.NullString{String: input.Contenido, Valid: true}
	}

	nota, err := h.Queries.UpdateNota(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nota no encontrada", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nota)
}

func (h *NotaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	err = h.Queries.DeleteNota(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
