-- name: CreateCategoria :one
INSERT INTO
    categoria (nombre, descripcion)
VALUES ($1, $2) RETURNING *;

-- name: GetCategoria :one
SELECT * FROM categoria WHERE id_categoria = $1 LIMIT 1;

-- name: ListCategorias :many
SELECT * FROM categoria ORDER BY nombre ASC;

-- name: UpdateCategoria :one
UPDATE categoria
SET
    nombre = $2,
    descripcion = $3
WHERE
    id_categoria = $1 RETURNING *;

-- name: DeleteCategoria :exec
DELETE FROM categoria WHERE id_categoria = $1;

-- name: CreateNota :one
INSERT INTO
    nota (
        titulo,
        contenido,
        is_archived
    )
VALUES ($1, $2, $3) RETURNING *;

-- name: GetNota :one
SELECT * FROM nota WHERE id_nota = $1 LIMIT 1;

-- name: ListNotas :many
SELECT * FROM nota ORDER BY id_nota DESC;

-- name: UpdateNota :one
UPDATE nota
SET
    titulo = $2,
    contenido = $3,
    is_archived = $4
WHERE
    id_nota = $1 RETURNING *;

-- name: ArchiveNote :one
UPDATE nota SET is_archived = $2 WHERE id_nota = $1 RETURNING *;

-- name: DeleteNota :exec
DELETE FROM nota WHERE id_nota = $1;

-- name: ListNotasByArchived :many
SELECT * FROM nota WHERE is_archived = $1 ORDER BY id_nota DESC;

-- name: AssignCategoriaToNota :exec
INSERT INTO
    nota_categoria (id_nota, id_categoria)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveCategoriaFromNota :exec
DELETE FROM nota_categoria WHERE id_nota = $1 AND id_categoria = $2;

-- name: ListNotasByCategoria :many
SELECT n.*
FROM nota n
    JOIN nota_categoria nc ON n.id_nota = nc.id_nota
WHERE
    nc.id_categoria = $1
    AND n.is_archived = $2
ORDER BY n.id_nota DESC;

-- name: ListCategoriasByNota :many
SELECT c.*
FROM
    categoria c
    JOIN nota_categoria nc ON c.id_categoria = nc.id_categoria
WHERE
    nc.id_nota = $1
ORDER BY c.nombre ASC;