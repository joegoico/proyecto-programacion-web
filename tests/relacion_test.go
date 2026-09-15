package tests

import (
	"context"
	"testing"

	"proyecto-programacion-web/backend/db"
)

func TestAssignCategoriaToNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota"})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat"})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	// Act
	err := testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notaID,
		IDCategoria: catID,
	})

	// Assert
	if err != nil {
		t.Fatalf("Error al asignar categoria a nota: %v", err)
	}
	catAsignadas, _ := testQueries.ListCategoriasByNota(ctx, notaID)
	if len(catAsignadas) != 1 || catAsignadas[0].IDCategoria != catID {
		t.Errorf("Se esperaba que la nota tuviera 1 categoria asignada")
	}
}

func TestRemoveCategoriaFromNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota"})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat"})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notaID,
		IDCategoria: catID,
	})

	// Act
	err := testQueries.RemoveCategoriaFromNota(ctx, db.RemoveCategoriaFromNotaParams{
		IDNota:      notaID,
		IDCategoria: catID,
	})

	// Assert
	if err != nil {
		t.Fatalf("Error al remover categoria de nota: %v", err)
	}
	catAsignadas, _ := testQueries.ListCategoriasByNota(ctx, notaID)
	if len(catAsignadas) != 0 {
		t.Errorf("Se esperaba que la nota no tuviera categorias asignadas")
	}
}

func TestListCategoriasByNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota Múltiple"})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat1"})
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat2"})
	categorias, _ := testQueries.ListCategorias(ctx)
	
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notaID,
		IDCategoria: categorias[0].IDCategoria,
	})
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notaID,
		IDCategoria: categorias[1].IDCategoria,
	})

	// Act
	catAsignadas, err := testQueries.ListCategoriasByNota(ctx, notaID)

	// Assert
	if err != nil {
		t.Fatalf("Error al listar categorias por nota: %v", err)
	}
	if len(catAsignadas) != 2 {
		t.Fatalf("Se esperaban 2 categorias, se obtuvieron %d", len(catAsignadas))
	}
}

func TestListNotasByCategoria(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateCategoria(ctx, db.CreateCategoriaParams{Nombre: "Cat Común"})
	categorias, _ := testQueries.ListCategorias(ctx)
	catID := categorias[0].IDCategoria

	testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota 1", IsArchived: false})
	testQueries.CreateNota(ctx, db.CreateNotaParams{Titulo: "Nota 2", IsArchived: false})
	notas, _ := testQueries.ListNotas(ctx)
	
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notas[0].IDNota,
		IDCategoria: catID,
	})
	testQueries.AssignCategoriaToNota(ctx, db.AssignCategoriaToNotaParams{
		IDNota:      notas[1].IDNota,
		IDCategoria: catID,
	})

	// Act
	notasCat, err := testQueries.ListNotasByCategoria(ctx, db.ListNotasByCategoriaParams{
		IDCategoria: catID,
		IsArchived:  false,
	})

	// Assert
	if err != nil {
		t.Fatalf("Error al listar notas por categoria: %v", err)
	}
	if len(notasCat) != 2 {
		t.Fatalf("Se esperaban 2 notas para Cat Común, se obtuvieron %d", len(notasCat))
	}
}
