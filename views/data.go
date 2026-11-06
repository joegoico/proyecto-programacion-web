package views

import (
	"strconv"

	"proyecto-programacion-web/backend/db"
)

// IndexData agrupa todo lo que necesita la página principal.
// El estado de los formularios (abierto/cerrado, creación/edición) lo decide
// el servidor a partir de la URL, ya que no hay JavaScript en el cliente.
type IndexData struct {
	Notas             []db.Notum
	Categorias        []db.Categorium
	CategoriasPorNota map[int32][]db.Categorium

	NotaAbierta       bool
	NotaEditando      *db.Notum // nil => formulario de creación
	CategoriaAbierta  bool
	CategoriaEditando *db.Categorium // nil => formulario de creación
}

// seleccionadas devuelve las categorías de la nota en edición (checkboxes marcados).
func (d IndexData) seleccionadas() []db.Categorium {
	if d.NotaEditando == nil {
		return nil
	}
	return d.CategoriasPorNota[d.NotaEditando.IDNota]
}

func idStr(id int32) string { return strconv.Itoa(int(id)) }

func notaAction(n *db.Notum) string {
	if n == nil {
		return "/notas"
	}
	return "/notas/" + idStr(n.IDNota) + "/editar"
}

func notaTitulo(n *db.Notum) string {
	if n == nil {
		return ""
	}
	return n.Titulo
}

func notaContenido(n *db.Notum) string {
	if n == nil {
		return ""
	}
	return n.Contenido.String
}

func categoriaAction(c *db.Categorium) string {
	if c == nil {
		return "/categorias"
	}
	return "/categorias/" + idStr(c.IDCategoria) + "/editar"
}

func categoriaNombre(c *db.Categorium) string {
	if c == nil {
		return ""
	}
	return c.Nombre
}

func estaSeleccionada(sel []db.Categorium, id int32) bool {
	for _, c := range sel {
		if c.IDCategoria == id {
			return true
		}
	}
	return false
}
