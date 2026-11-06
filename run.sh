#!/usr/bin/env bash

# Detener la ejecución si ocurre algún error en las etapas previas
set -e

# Cargar variables de entorno si existe el archivo
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

export TEST_DB_USER="${TEST_DB_USER:-testuser}" TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-testpassword}" TEST_DB_NAME="${TEST_DB_NAME:-notas_db_test}" TEST_DB_PORT="${TEST_DB_PORT:-5433}"
IMAGE_NAME="notas-servidor-go"
CONTAINER_NAME="gestor-notas-app"
PORT="8080"
TEST_COMPOSE_FILE="docker-compose.test.yml"

# Tareas posteriores (Cleanup automático al detener con Ctrl+C)
cleanup() {
  echo ""
  echo "=========================================="
  echo " Limpiando contenedores y volúmenes...    "
  echo "=========================================="
  docker compose down -v >/dev/null 2>&1 || true
  rm -f server
  echo " Limpieza completada exitosamente."
}
trap cleanup EXIT INT TERM

echo "=========================================="
echo " Generando código y corriendo tests...    "
echo "=========================================="

# 1. Generar vistas con templ
echo "-> Generando vistas con templ..."
if which templ > /dev/null 2>&1; then
    templ generate
elif [ -f "$HOME/go/bin/templ" ]; then
    "$HOME/go/bin/templ" generate
else
    go run github.com/a-h/templ/cmd/templ@v0.2.793 generate
fi

# 2. Generar código Go con sqlc
echo "-> Generando código de acceso a datos con sqlc..."
if which sqlc > /dev/null 2>&1; then
    sqlc generate
elif [ -f "$HOME/go/bin/sqlc" ]; then
    "$HOME/go/bin/sqlc" generate
elif docker run --rm -v "$PWD":/src -w /src sqlc/sqlc generate 2>/dev/null; then
    :
elif [ -f "backend/db/queries.sql.go" ]; then
    echo "-> (Aviso: No se pudo descargar/ejecutar sqlc vía Docker, pero el código en backend/db ya existe. Continuando...)"
else
    echo "Error: sqlc no está disponible y no se encuentra código en backend/db"
    exit 1
fi

# 3. Levantar la base de datos de testing
echo "-> Limpiando contenedores de testing previos..."
docker compose -f ${TEST_COMPOSE_FILE} down -v

echo "-> Levantando base de datos de testing..."
docker compose -f ${TEST_COMPOSE_FILE} up -d

# Autodetectar la IP directa del contenedor para evitar fallos de docker-proxy en snap
CONTAINER_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' gestor-notas-db-test 2>/dev/null || echo "")
if [ -n "$CONTAINER_IP" ]; then
  export TEST_DB_HOST="$CONTAINER_IP"
  export TEST_DB_PORT="5432"
fi

echo "-> Esperando a que la base de datos de testing esté lista..."
RETRIES=15
until (docker exec gestor-notas-db-test pg_isready -U ${TEST_DB_USER} -d ${TEST_DB_NAME} > /dev/null 2>&1 && docker exec gestor-notas-db-test psql -U ${TEST_DB_USER} -d ${TEST_DB_NAME} -c "SELECT 1 FROM nota LIMIT 1;" > /dev/null 2>&1) || [ $RETRIES -eq 0 ]; do
  echo "Esperando a postgres y esquema inicial..."
  sleep 1
  ((RETRIES--))
done
sleep 1

if [ $RETRIES -eq 0 ]; then
  echo "Error: Base de datos de pruebas no se inició a tiempo."
  docker compose -f ${TEST_COMPOSE_FILE} down -v
  exit 1
fi

# 4. Ejecutar los tests (unitarios de vistas, handlers e integración)
echo "-> Ejecutando tests en Go..."
go test -v ./views/... ./backend/handlers/... ./tests/...

# 5. Bajar la base de datos de testing
echo "-> Limpiando base de datos de testing..."
docker compose -f ${TEST_COMPOSE_FILE} down -v

echo "=========================================="
echo " Tests superados! Levantando servidor...  "
echo "=========================================="

# 6. Levantar la base de datos principal PostgreSQL en Docker
echo "-> Levantando base de datos principal PostgreSQL..."
docker compose up -d db

echo "-> Esperando a que PostgreSQL esté listo..."
RETRIES=15
until (docker exec gestor-notas-db pg_isready -U user -d notas_db > /dev/null 2>&1 && docker exec gestor-notas-db psql -U user -d notas_db -c "SELECT 1 FROM nota LIMIT 1;" > /dev/null 2>&1) || [ $RETRIES -eq 0 ]; do
  echo "Esperando a postgres..."
  sleep 1
  ((RETRIES--))
done

# Obtener IP del contenedor para evitar problemas de docker-proxy en snap
DB_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' gestor-notas-db 2>/dev/null || echo "")
if [ -z "$DB_IP" ]; then
  DB_IP="localhost"
fi

# 7. Intentar levantar la app con Docker Compose; si el confinamiento de Snap bloquea la descarga de imágenes, fallback a ejecución nativa
echo "-> Iniciando aplicación..."
set +e
docker compose up -d --build app >/dev/null 2>&1
DOCKER_APP_STATUS=$?
set -e

if [ $DOCKER_APP_STATUS -eq 0 ]; then
    echo "=========================================="
    echo " ¡Servidor y DB corriendo exitosamente en Docker!"
    echo " Abrir en el navegador: http://localhost:${PORT}"
    echo " Aplicación SSR con Go y Templ lista"
    echo "=========================================="
    echo "-> Mostrando logs del servidor (Ctrl+C para salir de los logs):"
    docker compose logs -f app
else
    echo "-> Compilando binario de Go localmente..."
    go build -o server ./backend
    echo "=========================================="
    echo " ¡Servidor y DB corriendo exitosamente!"
    echo " Base de datos: PostgreSQL en Docker"
    echo " Backend: Go + Templ en http://localhost:${PORT}"
    echo "=========================================="
    DATABASE_URL="postgres://user:password@${DB_IP}:5432/notas_db?sslmode=disable" PORT="${PORT}" ./server
fi
