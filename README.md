# proyecto-programacion-web
# Especificación del Sistema de Gestión de Notas

## 1. Descripción General
La aplicación está diseñada para la gestión eficiente de notas personales, permitiendo organizar la información mediante un sistema flexible de categorización y estados de archivo.

---

## 2. Funcionalidades Principales

* **Organización por Categorías:**
  * Cada nota puede asociarse a **una o más categorías** (relación muchos a muchos).
  * Filtrado dinámico de notas según las categorías seleccionadas.

* **Control de Estado (Archivado):**
  * Posibilidad de **archivar** y **desarchivar** notas de forma individual.
  * Vistas diferenciadas para consultar notas activas frente a notas archivadas.

* **Gestión de Contenido:**
  * Creación, edición y visualización de notas personales estructuradas.

---

## 3. Modelo de Dominio

### **Entidad: Nota**
Representa la unidad básica de información registrada por el usuario.

| Atributo | Tipo de Dato | Descripción |
| :--- | :--- | :--- |
| `id_nota` | Clave Primaria (PK) | Identificador único de la nota |
| `titulo` | Texto / String | Título descriptivo de la nota |
| `contenido` | Texto / Text | Cuerpo o detalle de la nota |
| `is_archived` | Booleano (`true`/`false`) | Indicador de estado (archivada o activa) |

---

### **Entidad: Categoría**
Etiqueta temática utilizada para clasificar y filtrar las notas.

| Atributo | Tipo de Dato | Descripción |
| :--- | :--- | :--- |
| `id_categoria` | Clave Primaria (PK) | Identificador único de la categoría |
| `nombre` | Texto / String | Nombre de la categoría (ej. *Trabajo*, *Personal*) |
| `descripcion` | Texto / String | Breve detalle sobre el propósito de la categoría |

---

## 4. Relación entre Entidades
* **Nota ↔ Categoría:** Relación **N:M (Muchos a Muchos)**. 
  * Una nota puede contener múltiples categorías.
  * Una categoría puede estar presente en múltiples notas.

---

## 🚀 Instrucciones de Ejecución

### Requisitos Previos
* [Docker](https://docs.docker.com/get-docker/) instalado y en ejecución en el sistema (con el plugin `docker compose`).
* [Go](https://golang.org/doc/install) instalado (para la ejecución de los tests locales y la compilación).
* `make` (estándar en Linux / macOS).

> El daemon de Docker debe estar corriendo y tu usuario debe poder usarlo sin `sudo`. El **puerto 8080** debe estar libre.

---

### Paso a Paso

1. **Clonar el repositorio y posicionarse en la rama `tp5`:**
   ```bash
   git clone <URL_DEL_REPOSITORIO>
   cd proyecto-programacion-web
   git checkout tp5
   ```

2. **(Opcional) Configurar Variables de Entorno:**
   Ya no es obligatorio crear un `.env`: hay valores por defecto. Si querés cambiar los datos de conexión de la base de datos de pruebas, creá un archivo `.env` en la raíz (ignorado por git):
   ```env
   TEST_DB_USER=testuser
   TEST_DB_PASSWORD=testpassword
   TEST_DB_NAME=notas_db_test
   TEST_DB_PORT=5433
   ```

3. **Otorgar permisos de ejecución a los scripts** (solo si no los tienen):
   ```bash
   chmod +x run.sh
   ```

4. **Ejecutar tests y levantar la aplicación:**
   Ejecuta el script orquestador. El mismo se encargará automáticamente del ciclo completo:
   - **Tareas previas:**
     - Limpieza de contenedores y volúmenes previos.
     - Generación de vistas con `templ generate`.
     - Generación de acceso a base de datos con `sqlc`.
     - Levantado de PostgreSQL temporal y ejecución de la suite de tests (unitarios e integración).
     - Compilación y levantado de PostgreSQL con Docker.
   - **Ejecución:**
     - Inicia el servidor web en http://localhost:8080.
   - **Tareas posteriores (Cleanup):**
     - Al presionar `Ctrl+C`, el manejador `trap` detiene y borra automáticamente los contenedores y volúmenes (`docker compose down -v`).

   ```bash
   ./run.sh
   # o alternativamente:
   make run
   ```

5. **Acceder a la aplicación:**
   Abre tu navegador web e ingresa a:
   👉 **[http://localhost:8080](http://localhost:8080)**

---

### 🛑 Detener el Contenedor
Para detener manualmente y eliminar los contenedores y volúmenes:
```bash
make down
# o:
docker compose down -v
```

---

### Comandos disponibles (`make help` los lista)

| Comando | Descripción |
| :--- | :--- |
| `make run` / `./run.sh` | Ciclo completo: tareas previas (templ, sqlc, tests), ejecución de la app y limpieza posterior |
| `make templ` | Genera código Go a partir de las plantillas templ (`templ generate`) |
| `make sqlc` | Genera el código de acceso a datos con sqlc |
| `make build` | templ + sqlc + compilación de Go |
| `make test` | Tests de Go (unitarios de vistas, handlers e integración con PostgreSQL) |
| `make up` | Levanta PostgreSQL y la app en segundo plano |
| `make down` / `make clean` | Detiene y borra contenedores y volúmenes |

---

## 🌐 Renderizado del Lado del Servidor con Templ (TP5)

En esta entrega se reemplazó la arquitectura de frontend separada (SPA con HTML/JS/CSS) por una aplicación renderizada en el servidor (SSR) mediante **`templ`** en Go:

- **Sin JavaScript en el cliente:** Se eliminaron todos los archivos `.js`. La apertura/cierre de formularios y el modo edición se controlan mediante estado semántico en la URL (`?nota=nueva`, `?nota=ID`, `?categoria=nueva`, `?categoria=ID`).
- **Patrón Post/Redirect/Get (PRG):** Los formularios clásicos envían peticiones `POST` (`/notas`, `/notas/{id}/editar`, `/notas/{id}/eliminar`, `/categorias`, `/categorias/{id}/editar`, `/categorias/{id}/eliminar`) y el servidor responde con redirección `303 See Other` hacia `/`, evitando reenvíos accidentales al refrescar el navegador.
- **Componentes lógicos en `views/`:**
  - `layout.templ`: Estructura HTML base con estilos embebidos.
  - `entity_list.templ`: Listado de notas con chips de categorías y acciones (Editar / Eliminar).
  - `entity_form.templ`: Formulario de creación/edición de notas con selección múltiple de categorías.
  - `categoria_list.templ`: Listado de categorías con opciones para renombrar y eliminar.
  - `categoria_form.templ`: Formulario inline para añadir o editar categorías.
  - `index.templ`: Orquestador principal de la vista.
  - `data.go`: Tipado estático con `IndexData`.
- **Estilos Embebidos:** La hoja de estilos `styles.css` se incluye directamente en el binario compilado mediante `//go:embed` en `backend/static/static.go`.

---

## 🧪 Pruebas Automatizadas (TDD con Patrón AAA)

Para ejecutar la batería completa de pruebas unitarias y de integración:

```bash
make test
```

Cubre:
1. **Renderizado de Vistas (`./views/...`):** Comprueba que cada componente `templ` genere el HTML semántico esperado.
2. **Handlers Web (`./backend/handlers/...`):** Pruebas unitarias con mocks en memoria verificando validaciones (400), no existencia (404), sincronización de categorías y redirecciones PRG (303).
3. **Integración (`./tests/...`):** Pruebas de extremo a extremo contra una base de datos PostgreSQL real temporal.

---

## Estructura del proyecto

```
.
├── backend/
│   ├── main.go             # Conexión a la DB, rutas web SSR y arranque del servidor
│   ├── handlers/           # Handlers HTTP SSR (web.go y tests web_test.go)
│   ├── db/                 # Código generado por sqlc (ignorado en git)
│   ├── static/             # Hoja de estilos styles.css embebida con go:embed
│   └── sql/                # schema.sql y queries.sql (fuente para sqlc)
├── views/                  # Componentes Templ (*.templ y *_templ.go)
│   ├── layout.templ        # Estructura base HTML
│   ├── entity_list.templ   # Lista de notas con chips y botones
│   ├── entity_form.templ   # Formulario de alta y edición de notas
│   ├── categoria_list.templ# Lista de categorías
│   ├── categoria_form.templ# Formulario de categoría
│   ├── index.templ         # Ensamblador principal
│   └── views_test.go       # Tests unitarios de vistas
├── tests/                  # Tests de integración contra PostgreSQL temporal
├── sqlc.yaml               # Configuración de sqlc
├── docker-compose.yml      # Base de datos PostgreSQL principal
├── docker-compose.test.yml # PostgreSQL temporal para los tests de Go
├── Dockerfile              # Imagen de producción de la app
├── .air.toml               # Recarga automática en desarrollo
├── Makefile                # make run, make templ, make test, make down, ...
└── run.sh                  # Ciclo completo: tareas previas, ejecución y limpieza
```

**Arquitectura por capas:** `main.go` (enrutador HTTP) → `handlers` (lógica web SSR, PRG y validación) → `views` (componentes templ fuertemente tipados) → `db` (consultas sqlc) → PostgreSQL.
