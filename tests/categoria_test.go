package tests

import (
	"context"
	"database/sql"
	"testing"

	"proyecto-programacion-web/backend/db"
)

func TestCreateCategoria(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	params := db.CreateCategoriaParams{
		Nombre:      "Trabajo",
		Descripcion: sql.NullString{String: "Tareas del trabajo", Valid: true},
	}

	// Act
	cat, err := testQueries.CreateCategoria(ctx, params)

	// Assert
	if err != nil {
		t.Fatalf("No se pudo crear la categoria: %v", err)
	}
	if cat.Nombre != params.Nombre {
		t.Errorf("Se esperaba %s, se obtuvo %s", params.Nombre, cat.Nombre)
	}
}

func TestGetCategoria(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{
		Nombre:      "Personal",
		Descripcion: sql.NullString{String: "Cosas personales", Valid: true},
	})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	// Act
	catObtenida, err := testQueries.GetCategoria(ctx, catID)

	// Assert
	if err != nil {
		t.Fatalf("No se pudo obtener la categoria: %v", err)
	}
	if catObtenida.Nombre != "Personal" {
		t.Errorf("Se esperaba Personal, se obtuvo %s", catObtenida.Nombre)
	}
}

func TestUpdateCategoria(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{
		Nombre: "Viejo",
	})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	updateParams := db.UpdateCategoriaParams{
		IDCategoria: catID,
		Nombre:      "Nuevo",
		Descripcion: sql.NullString{String: "Desc nueva", Valid: true},
	}

	// Act
	catUpdate, err := testQueries.UpdateCategoria(ctx, updateParams)

	// Assert
	if err != nil {
		t.Fatalf("No se pudo actualizar la categoria: %v", err)
	}
	if catUpdate.Nombre != "Nuevo" {
		t.Errorf("Se esperaba Nuevo, se obtuvo %s", catUpdate.Nombre)
	}
}

func TestDeleteCategoria(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{
		Nombre: "Para Borrar",
	})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	// Act
	err := testQueries.DeleteCategoria(ctx, catID)

	// Assert
	if err != nil {
		t.Fatalf("No se pudo borrar la categoria: %v", err)
	}
	_, errGet := testQueries.GetCategoria(ctx, catID)
	if errGet == nil {
		t.Errorf("Se esperaba error al obtener categoria borrada, pero no hubo error")
	}
}

func TestListCategorias(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat A"})
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat B"})

	// Act
	categorias, err := testQueries.ListCategorias(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Error al listar categorias: %v", err)
	}
	if len(categorias) != 2 {
		t.Errorf("Se esperaban 2 categorias, se obtuvieron %d", len(categorias))
	}
}
