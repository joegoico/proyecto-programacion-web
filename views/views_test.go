package views_test

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"proyecto-programacion-web/backend/db"
	"proyecto-programacion-web/views"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	buf := new(bytes.Buffer)
	if err := c.Render(context.Background(), buf); err != nil {
		t.Fatalf("error al renderizar: %v", err)
	}
	return buf.String()
}

func assertContains(t *testing.T, html string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(html, sub) {
			t.Errorf("se esperaba encontrar %q en: %s", sub, html)
		}
	}
}

func assertNotContains(t *testing.T, html string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if strings.Contains(html, sub) {
			t.Errorf("no se esperaba encontrar %q en: %s", sub, html)
		}
	}
}

func TestLayout(t *testing.T) {
	// Arrange
	component := views.Layout("App de Notas")

	// Act
	html := render(t, component)

	// Assert
	assertContains(t, html,
		`<html lang="es">`,
		"<title>App de Notas</title>",
		`<link rel="stylesheet" href="/styles.css">`,
	)
	assertNotContains(t, html, "pico")
}

func TestEntityList_SinNotas(t *testing.T) {
	// Arrange
	component := views.EntityList([]db.Notum{}, nil)

	// Act
	html := render(t, component)

	// Assert
	assertContains(t, html, "<h2>Notas</h2>", `<ul id="lista-notas">`)
	assertNotContains(t, html, "<li")
}

func TestEntityList_ConNotasChipsYAcciones(t *testing.T) {
	// Arrange
	notas := []db.Notum{
		{IDNota: 7, Titulo: "Primera Nota", Contenido: sql.NullString{String: "Contenido 1", Valid: true}},
		{IDNota: 8, Titulo: "Sin contenido"},
	}
	cats := map[int32][]db.Categorium{
		7: {{IDCategoria: 1, Nombre: "Trabajo"}},
	}

	// Act
	html := render(t, views.EntityList(notas, cats))

	// Assert
	assertContains(t, html,
		"<strong>Primera Nota</strong>",
		"<p>Contenido 1</p>",
		`<span class="chip">Trabajo</span>`,
		"<strong>Sin contenido</strong>",
		`<div class="acciones">`,
		`<input type="hidden" name="nota" value="7">`,
		`<button type="submit" class="editar">Editar</button>`,
		`action="/notas/7/eliminar" method="POST"`,
		`<button type="submit" class="eliminar">Eliminar</button>`,
		`action="/notas/8/eliminar" method="POST"`,
	)
}

func TestEntityForm_Creacion(t *testing.T) {
	// Arrange
	categorias := []db.Categorium{{IDCategoria: 1, Nombre: "Trabajo"}, {IDCategoria: 2, Nombre: "Personal"}}

	// Act
	html := render(t, views.EntityForm(categorias, nil, nil))

	// Assert
	assertContains(t, html,
		`<form id="nota-form" action="/notas" method="POST">`,
		"<h2>Nueva nota</h2>",
		`name="titulo"`,
		`placeholder="Título"`,
		`name="contenido"`,
		"<legend>Categorías</legend>",
		`<div id="categorias-select">`,
		`<input type="checkbox" name="categorias" value="1"><span>Trabajo</span>`,
		`<input type="checkbox" name="categorias" value="2"><span>Personal</span>`,
		">Crear</button>",
	)
	assertNotContains(t, html, "checked")
}

func TestEntityForm_Edicion(t *testing.T) {
	// Arrange
	categorias := []db.Categorium{{IDCategoria: 1, Nombre: "Trabajo"}, {IDCategoria: 2, Nombre: "Personal"}}
	nota := &db.Notum{IDNota: 5, Titulo: "Titulo previo", Contenido: sql.NullString{String: "Contenido previo", Valid: true}}
	seleccionadas := []db.Categorium{{IDCategoria: 2, Nombre: "Personal"}}

	// Act
	html := render(t, views.EntityForm(categorias, nota, seleccionadas))

	// Assert
	assertContains(t, html,
		`action="/notas/5/editar" method="POST"`,
		`value="Titulo previo"`,
		"Contenido previo</textarea>",
		`value="2" checked`,
	)
	assertNotContains(t, html, `value="1" checked`)
}

func TestCategoriaForm_Creacion(t *testing.T) {
	// Act
	html := render(t, views.CategoriaForm(nil))

	// Assert
	assertContains(t, html,
		`<form id="categoria-form" action="/categorias" method="POST" class="inline">`,
		`name="nombre"`,
		`placeholder="Nombre de la categoría"`,
		">Guardar</button>",
	)
}

func TestCategoriaForm_Edicion(t *testing.T) {
	// Arrange
	cat := &db.Categorium{IDCategoria: 9, Nombre: "Estudio"}

	// Act
	html := render(t, views.CategoriaForm(cat))

	// Assert
	assertContains(t, html, `action="/categorias/9/editar" method="POST"`, `value="Estudio"`)
}

func TestCategoriaList(t *testing.T) {
	// Arrange
	categorias := []db.Categorium{{IDCategoria: 3, Nombre: "Urgente"}}

	// Act
	html := render(t, views.CategoriaList(categorias))

	// Assert
	assertContains(t, html,
		`<ul id="lista-categorias" class="chips">`,
		`<li class="chip"><span>Urgente</span>`,
		`<input type="hidden" name="categoria" value="3">`,
		`class="editar-cat">✎</button>`,
		`action="/categorias/3/eliminar" method="POST"`,
		`class="eliminar-cat">×</button>`,
	)
}

func indexData() views.IndexData {
	return views.IndexData{
		Notas:             []db.Notum{{IDNota: 10, Titulo: "Nota Integración"}},
		Categorias:        []db.Categorium{{IDCategoria: 1, Nombre: "Categoría En Index"}},
		CategoriasPorNota: map[int32][]db.Categorium{10: {{IDCategoria: 1, Nombre: "Categoría En Index"}}},
	}
}

func TestIndexPage_FormulariosCerradosPorDefecto(t *testing.T) {
	// Arrange
	data := indexData()

	// Act
	html := render(t, views.IndexPage(data))

	// Assert
	assertContains(t, html,
		"<!doctype html>",
		"<h1>Aplicación de Gestión de Notas Personales</h1>",
		`<section id="categorias">`,
		`id="btn-agregar-categoria"`,
		`id="btn-agregar-nota"`,
		`<input type="hidden" name="nota" value="nueva">`,
		`<input type="hidden" name="categoria" value="nueva">`,
		"Nota Integración",
		"Categoría En Index",
	)
	assertNotContains(t, html, `id="nota-form"`, `id="categoria-form"`)
}

func TestIndexPage_FormulariosAbiertos(t *testing.T) {
	// Arrange
	data := indexData()
	data.NotaAbierta = true
	data.CategoriaAbierta = true

	// Act
	html := render(t, views.IndexPage(data))

	// Assert: abiertos, y el botón pasa a cerrar (sin parámetro "nueva")
	assertContains(t, html, `id="nota-form"`, `id="categoria-form"`, `action="/notas"`, `action="/categorias"`)
	assertNotContains(t, html, `value="nueva"`)
}

func TestIndexPage_EdicionDeNotaYCategoria(t *testing.T) {
	// Arrange
	data := indexData()
	nota := data.Notas[0]
	cat := data.Categorias[0]
	data.NotaAbierta, data.NotaEditando = true, &nota
	data.CategoriaAbierta, data.CategoriaEditando = true, &cat

	// Act
	html := render(t, views.IndexPage(data))

	// Assert
	assertContains(t, html,
		`action="/notas/10/editar" method="POST"`,
		`value="Nota Integración"`,
		`value="1" checked`,
		`action="/categorias/1/editar" method="POST"`,
		`value="Categoría En Index"`,
	)
}
