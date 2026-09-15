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
* [Docker](https://docs.docker.com/get-docker/) instalado y en ejecución en el sistema.
* [Go](https://golang.org/doc/install) instalado (para la ejecución de los tests locales).

---

### Paso a Paso

1. **Clonar el repositorio:**
   ```bash
   git clone <URL_DEL_REPOSITORIO>
   cd proyecto-programacion-web
   ```

2. **Configurar Variables de Entorno:**
   Debes crear un archivo `.env` en la raíz del proyecto para definir los datos de conexión a la base de datos de pruebas (la cual será ignorada por git). Puedes usar el siguiente template:
   ```env
   TEST_DB_USER=testuser
   TEST_DB_PASSWORD=testpassword
   TEST_DB_NAME=notas_db_test
   TEST_DB_PORT=5433
   ```

3. **Otorgar permisos de ejecución al script:**
   ```bash
   chmod +x run.sh
   ```

4. **Ejecutar tests y levantar la aplicación:**
   Ejecuta el script orquestador. El mismo se encargará automáticamente de:
   - Cargar tu `.env`.
   - Generar el código a nivel de base de datos usando `sqlc`.
   - Levantar una base de datos PostgreSQL de pruebas usando contenedores de Docker.
   - **Correr todos los tests automatizados** de CRUD (usando el paquete estandar `testing` de Go).
   - Eliminar los contenedores y volúmenes de pruebas post-ejecución.
   - Construir y levantar el servidor principal para la aplicación real.

   ```bash
   ./run.sh
   ```

5. **Acceder a la aplicación:**
   Si los tests pasan exitosamente, abre tu navegador web e ingresa a:
   👉 **[http://localhost:8080](http://localhost:8080)**

---

### 🛑 Detener el Contenedor

Para detener y eliminar el contenedor de la aplicación en cualquier momento, ejecuta:
```bash
docker stop gestor-notas-app && docker rm gestor-notas-app
```