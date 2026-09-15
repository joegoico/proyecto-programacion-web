package tests

import (
	"context"
	"database/sql"
	"testing"

	"proyecto-programacion-web/backend/db"
)

func TestCreateNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	params := db.CreateNotaParams{
		Titulo:     "Nota 1",
		Contenido:  sql.NullString{String: "Contenido 1", Valid: true},
		IsArchived: false,
	}

	// Act
	nota, err := testQueries.CreateNota(ctx, params)

	// Assert
	if err != nil {
		t.Fatalf("Error al crear nota: %v", err)
	}
	if nota.Titulo != params.Titulo {
		t.Errorf("Se esperaba Titulo %s, se obtuvo %s", params.Titulo, nota.Titulo)
	}
}

func TestGetNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo: "Nota Get",
	})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	// Act
	notaObtenida, err := testQueries.GetNota(ctx, notaID)

	// Assert
	if err != nil {
		t.Fatalf("Error al obtener nota: %v", err)
	}
	if notaObtenida.Titulo != "Nota Get" {
		t.Errorf("Se esperaba Nota Get, se obtuvo %s", notaObtenida.Titulo)
	}
}

func TestUpdateNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo: "Nota Vieja",
	})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	updateParams := db.UpdateNotaParams{
		IDNota:     notaID,
		Titulo:     "Nota Modificada",
		Contenido:  sql.NullString{String: "Contenido modificado", Valid: true},
		IsArchived: false,
	}

	// Act
	notaUpdate, err := testQueries.UpdateNota(ctx, updateParams)

	// Assert
	if err != nil {
		t.Fatalf("Error al actualizar nota: %v", err)
	}
	if notaUpdate.Titulo != "Nota Modificada" {
		t.Errorf("Se esperaba Nota Modificada, se obtuvo %s", notaUpdate.Titulo)
	}
}

func TestArchiveNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo:     "Nota para Archivar",
		IsArchived: false,
	})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	archiveParams := db.ArchiveNoteParams{
		IDNota:     notaID,
		IsArchived: true,
	}

	// Act
	notaArchive, err := testQueries.ArchiveNote(ctx, archiveParams)

	// Assert
	if err != nil {
		t.Fatalf("Error al archivar nota: %v", err)
	}
	if !notaArchive.IsArchived {
		t.Errorf("Se esperaba is_archived = true")
	}
}

func TestDeleteNota(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo: "Nota a Borrar",
	})
	notas, _ := testQueries.ListNotas(ctx)
	notaID := notas[0].IDNota

	// Act
	err := testQueries.DeleteNota(ctx, notaID)

	// Assert
	if err != nil {
		t.Fatalf("Error al borrar nota: %v", err)
	}
	_, errGet := testQueries.GetNota(ctx, notaID)
	if errGet == nil {
		t.Errorf("Se esperaba error al obtener nota borrada")
	}
}

func TestListNotasByArchived(t *testing.T) {
	// Arrange
	clearDB()
	ctx := context.Background()
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo:     "Nota Activa",
		IsArchived: false,
	})
	testQueries.CreateNota(ctx, db.CreateNotaParams{
		Titulo:     "Nota Archivada",
		IsArchived: true,
	})

	// Act
	activas, errAct := testQueries.ListNotasByArchived(ctx, false)
	archivadas, errArch := testQueries.ListNotasByArchived(ctx, true)

	// Assert
	if errAct != nil || errArch != nil {
		t.Fatalf("Errores al listar: %v / %v", errAct, errArch)
	}
	if len(activas) != 1 || activas[0].Titulo != "Nota Activa" {
		t.Errorf("Se esperaba 1 nota activa 'Nota Activa'")
	}
	if len(archivadas) != 1 || archivadas[0].Titulo != "Nota Archivada" {
		t.Errorf("Se esperaba 1 nota archivada 'Nota Archivada'")
	}
}
