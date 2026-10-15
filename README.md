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


  ## 🚀 Instrucciones de Ejecución

### Requisitos Previos
* [Docker](https://docs.docker.com/get-docker/) instalado y en ejecución en el sistema (con el plugin `docker compose`).
* [Go](https://golang.org/doc/install) instalado (para la ejecución de los tests locales y la compilación).
* `make` y `curl` (para `make test-api`).

> El daemon de Docker debe estar corriendo y tu usuario debe poder usarlo sin `sudo`. El **puerto 8080** debe estar libre.

---

### Paso a Paso

1. **Clonar el repositorio** (el TP3 está en la branch `tp3`):
   ```bash
   git clone <URL_DEL_REPOSITORIO>
   cd proyecto-programacion-web
   git checkout tp3
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
   chmod +x run.sh test-api.sh
   ```

4. **Ejecutar tests y levantar la aplicación:**
   Ejecuta el script orquestador. El mismo se encargará automáticamente de:
   - Cargar tu `.env` (si existe).
   - Generar el código a nivel de base de datos usando `sqlc`.
   - Levantar una base de datos PostgreSQL de pruebas usando contenedores de Docker.
   - **Correr todos los tests automatizados** de CRUD (usando el paquete estandar `testing` de Go).
   - Eliminar los contenedores y volúmenes de pruebas post-ejecución.
   - Construir y levantar el servidor principal (app + base de datos con `docker compose`).

   ```bash
   ./run.sh
   ```

5. **Acceder a la aplicación:**
   Si los tests pasan exitosamente, abre tu navegador web e ingresa a:
   👉 **[http://localhost:8080](http://localhost:8080)**

---

### 🛑 Detener el Contenedor

Para detener y eliminar el contenedor de la aplicación y base de datos, ejecuta:
```bash
docker compose down -v
```

---

## 🧪 Pruebas de la API (TP3)

Se implementó una API REST para gestionar categorías y notas, bajo la ruta base `/api`. La lógica está modulada en `backend/handlers` (capa de lógica de negocio).

### Probar todo con un solo comando (`make test-api`)

Después de clonar el repo y hacer `git checkout tp3`, ejecutar:

```bash
make test-api
```

No hace falta crear un `.env` ni instalar nada más. `make test-api` ejecuta [test-api.sh](test-api.sh), que hace automáticamente:

1. **Tareas previas**
   1. Verifica que existan `docker`, `go` y `curl`.
   2. Ejecuta **sqlc** (`sqlc generate`, vía Docker) para generar `backend/db`.
   3. **Compila** el proyecto (`go build ./...`).
   4. **Borra** contenedores y volúmenes anteriores (`docker compose down -v`).
   5. **Levanta** los contenedores: base de datos PostgreSQL y aplicación.
   6. **Espera** a que PostgreSQL esté listo (`pg_isready`) y a que la API responda.
      La primera vez puede tardar uno o dos minutos porque el contenedor descarga e instala `air`.
2. **Prueba de cada endpoint** con `curl`: se verifica el código HTTP y parte del cuerpo de cada respuesta.
   Se prueban los 10 endpoints (5 de categorías y 5 de notas) y los casos de error (400 y 404). Cada prueba muestra ✅ o ❌.
3. **Tareas posteriores**: borra contenedores y volúmenes. Esto se hace siempre, incluso si alguna prueba falla.

El resultado final se ve así, y el comando termina con código `0` si todo pasó (distinto de `0` si algo falló):

```
 Resultado: 30 OK, 0 fallidas
```

### Otros comandos (`make help` los lista)

| Comando | Descripción |
| :--- | :--- |
| `make test-api` | Todo el flujo de pruebas de la API (arriba) |
| `make sqlc` | Solo genera el código con sqlc |
| `make build` | sqlc + compilación |
| `make up` | Levanta app + base de datos (queda corriendo en http://localhost:8080) |
| `make down` | Borra contenedores y volúmenes |
| `make test` | Tests de Go (carpeta `tests/`) contra una base de datos de testing temporal (puerto 5433) |
| `make run` / `./run.sh` | Tests de Go y luego levanta la app con logs en pantalla |

### Ejecutar la aplicación manualmente

```bash
make up                                   # levantar
curl http://localhost:8080/api/categorias # probar
make down                                 # bajar y borrar datos
```

La app corre con [air](https://github.com/air-verse/air) (configurado en `.air.toml`), por lo que recompila sola al guardar cambios en el código.

---

## Endpoints de la API REST

Ruta base: `http://localhost:8080/api`. Las respuestas con datos son JSON.

### Categorías

| Método | Ruta | Body (JSON) | Respuestas |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/categorias` | `{"nombre":"Trabajo","descripcion":"..."}` | `201` con la categoría creada · `400` si el nombre está vacío o el JSON es inválido |
| `GET` | `/api/categorias` | – | `200` con un array |
| `GET` | `/api/categorias/{id}` | – | `200` · `400` id inválido · `404` no existe |
| `PUT` | `/api/categorias/{id}` | `{"nombre":"...","descripcion":"..."}` | `200` · `400` · `404` |
| `DELETE` | `/api/categorias/{id}` | – | `204` · `400` |

### Notas

| Método | Ruta | Body (JSON) | Respuestas |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/notas` | `{"titulo":"...","contenido":"...","is_archived":false}` | `201` con la nota creada · `400` si el título está vacío o el JSON es inválido |
| `GET` | `/api/notas` | – | `200` con un array |
| `GET` | `/api/notas/{id}` | – | `200` · `400` id inválido · `404` no existe |
| `PUT` | `/api/notas/{id}` | `{"titulo":"...","contenido":"...","is_archived":true}` | `200` · `400` · `404` |
| `DELETE` | `/api/notas/{id}` | – | `204` · `400` |

Ejemplo manual:

```bash
curl -X POST http://localhost:8080/api/categorias \
  -H "Content-Type: application/json" \
  -d '{"nombre":"Trabajo","descripcion":"Tareas del trabajo"}'
```

---

## Estructura del proyecto

```
.
├── backend/
│   ├── main.go             # Conexión a la DB, rutas y arranque del servidor
│   ├── handlers/           # Capa de lógica de negocio: handlers HTTP CRUD (categorias.go, notas.go)
│   ├── db/                 # Código generado por sqlc (NO editar a mano)
│   └── sql/                # schema.sql y queries.sql (fuente para sqlc)
├── frontend/               # Interfaz web
├── tests/                  # Tests de Go de la capa de datos y de los handlers
├── sqlc.yaml               # Configuración de sqlc
├── docker-compose.yml      # App (con air) + PostgreSQL
├── docker-compose.test.yml # PostgreSQL temporal para los tests de Go
├── Dockerfile              # Imagen de producción de la app
├── .air.toml               # Recarga automática en desarrollo
├── Makefile                # make test-api, make up, make down, ...
├── test-api.sh             # Pruebas de la API con curl (flujo completo)
└── run.sh                  # Tests de Go + levantar la app
```

**Arquitectura por capas:** `main.go` (rutas) → `handlers` (validación y lógica HTTP) → `db` (consultas generadas por sqlc) → PostgreSQL.
