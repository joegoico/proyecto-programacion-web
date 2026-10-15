#!/usr/bin/env bash

# Detener la ejecución si ocurre algún error
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

echo "=========================================="
echo " Generando código y corriendo tests...    "
echo "=========================================="

# 1. Generar código Go con sqlc vía Docker
echo "-> Generando código de acceso a datos con sqlc..."
docker run --rm -v "$PWD":/src -w /src sqlc/sqlc generate

# 2. Levantar la base de datos de testing
echo "-> Limpiando contenedores de testing previos..."
docker compose -f ${TEST_COMPOSE_FILE} down -v

echo "-> Levantando base de datos de testing..."
docker compose -f ${TEST_COMPOSE_FILE} up -d

echo "-> Esperando a que la base de datos de testing esté lista..."
# Hacemos un bucle para esperar a que postgres esté disponible
RETRIES=10
until docker exec gestor-notas-db-test pg_isready -U ${TEST_DB_USER} -d ${TEST_DB_NAME} > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
  echo "Esperando a postgres..."
  sleep 2
  ((RETRIES--))
done

if [ $RETRIES -eq 0 ]; then
  echo "Error: Base de datos de pruebas no se inició a tiempo."
  docker compose -f ${TEST_COMPOSE_FILE} down -v
  exit 1
fi

# 3. Ejecutar los tests
echo "-> Ejecutando tests en Go..."
go test -v ./tests/...

# 4. Bajar la base de datos de testing
echo "-> Limpiando base de datos de testing..."
docker compose -f ${TEST_COMPOSE_FILE} down -v

echo "=========================================="
echo " Tests superados! Levantando servidor...  "
echo "=========================================="

# 5. Detener y remover contenedor previo si ya estaba corriendo
echo "-> Limpiando contenedores de app previos..."
if [ "$(docker ps -aq -f name=^gestor-notas-app$)" ]; then
    docker rm -f gestor-notas-app > /dev/null 2>&1
fi
docker compose down -v

# 6. Levantar la aplicación con docker-compose
echo "-> Construyendo e iniciando contenedores de la app y base de datos..."
docker compose up -d --build

echo "=========================================="
echo " ¡Servidor y DB corriendo exitosamente!"
echo " Abrir en el navegador: http://localhost:${PORT}"
echo " La API responderá en: http://localhost:${PORT}/api/categorias"
echo "=========================================="
echo "-> Mostrando logs del servidor (Ctrl+C para salir de los logs):"
docker compose logs -f app