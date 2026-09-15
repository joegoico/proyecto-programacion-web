-- 1. Tabla de Categorías
CREATE TABLE categoria (
    id_categoria SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    descripcion TEXT
);

-- 2. Tabla de Notas
CREATE TABLE nota (
    id_nota SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    contenido TEXT,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE
);

-- 3. Tabla Intermedia (Relación N:M)
CREATE TABLE nota_categoria (
    id_nota INT NOT NULL,
    id_categoria INT NOT NULL,
    PRIMARY KEY (id_nota, id_categoria),
    CONSTRAINT fk_nota FOREIGN KEY (id_nota) REFERENCES nota (id_nota) ON DELETE CASCADE,
    CONSTRAINT fk_categoria FOREIGN KEY (id_categoria) REFERENCES categoria (id_categoria) ON DELETE CASCADE
);