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

---

### Paso a Paso

1. **Clonar o ubicarse en la raíz del repositorio:**
   ```bash
   cd proyecto-programacion-web
   ```

2. **Otorgar permisos de ejecución al script (solo la primera vez si es necesario):**
   ```bash
   chmod +x run.sh
   ```

3. **Ejecutar la aplicación:**
   ```bash
   ./run.sh
   ```

4. **Acceder a la aplicación:**
   Abre tu navegador web e ingresa a:
   👉 **[http://localhost:8080](http://localhost:8080)**

---

### 🛑 Detener el Contenedor

Para detener y eliminar el contenedor en cualquier momento, ejecuta:
```bash
docker stop servidor-notas-app && docker rm servidor-notas-app
```