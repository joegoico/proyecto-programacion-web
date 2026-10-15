#!/usr/bin/env bash


set -uo pipefail
cd "$(dirname "$0")"

BASE_URL="http://localhost:8080/api"
DB_CONTAINER="gestor-notas-db"
APP_CONTAINER="gestor-notas-app"
WAIT_APP_SECONDS=180
BODY_FILE="$(mktemp)"
PASSED=0
FAILED=0

# ---------------------------------------------------------------- utilidades
cleanup() {
  rm -f "$BODY_FILE"
  echo
  echo "==> [POST] Borrando contenedores y volúmenes..."
  docker rm -f "$APP_CONTAINER" >/dev/null 2>&1 || true
  docker compose down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

die() { echo "❌ $*" >&2; exit 1; }

# check "descripción" CODIGO_ESPERADO METODO URL [json_body]
# Deja el cuerpo de la respuesta en $BODY_FILE y el código en $CODE.
check() {
  local desc="$1" expected="$2" method="$3" url="$4" data="${5:-}"
  if [ -n "$data" ]; then
    CODE=$(curl -s -o "$BODY_FILE" -w '%{http_code}' -X "$method" "$url" \
      -H 'Content-Type: application/json' -d "$data")
  else
    CODE=$(curl -s -o "$BODY_FILE" -w '%{http_code}' -X "$method" "$url")
  fi
  if [ "$CODE" = "$expected" ]; then
    echo "  ✅ $desc  ($method ${url#http://localhost:8080} -> $CODE)"
    PASSED=$((PASSED + 1))
  else
    echo "  ❌ $desc  ($method ${url#http://localhost:8080}) esperado=$expected obtenido=$CODE"
    echo "     respuesta: $(cat "$BODY_FILE")"
    FAILED=$((FAILED + 1))
  fi
}

# extract_id CAMPO  -> lee un entero del JSON guardado en $BODY_FILE
extract_id() { grep -o "\"$1\":[0-9]*" "$BODY_FILE" | head -1 | cut -d: -f2; }

# body_has "texto" -> verifica que la última respuesta contenga el texto
body_has() {
  if grep -q "$1" "$BODY_FILE"; then
    echo "  ✅ la respuesta contiene $1"
    PASSED=$((PASSED + 1))
  else
    echo "  ❌ la respuesta NO contiene $1: $(cat "$BODY_FILE")"
    FAILED=$((FAILED + 1))
  fi
}

# ------------------------------------------------------------ tareas previas
echo "==> [PRE] Verificando herramientas..."
for tool in docker go curl; do
  command -v "$tool" >/dev/null 2>&1 || die "Falta '$tool' en el PATH."
done
docker compose version >/dev/null 2>&1 || die "Falta el plugin 'docker compose'."

echo "==> [PRE] Generando código con sqlc..."
docker run --rm -v "$PWD":/src -w /src sqlc/sqlc generate || die "sqlc falló."

echo "==> [PRE] Compilando..."
go build ./... || die "La compilación falló."

echo "==> [PRE] Borrando contenedores y volúmenes previos..."
docker rm -f "$APP_CONTAINER" >/dev/null 2>&1 || true
docker compose down -v >/dev/null 2>&1 || true

echo "==> [PRE] Levantando contenedores (app + base de datos)..."
docker compose up -d || die "docker compose up falló."

echo "==> [PRE] Esperando a que PostgreSQL esté listo..."
for i in $(seq 1 30); do
  docker exec "$DB_CONTAINER" pg_isready -U user -d notas_db >/dev/null 2>&1 && break
  [ "$i" = 30 ] && die "PostgreSQL no arrancó a tiempo."
  sleep 2
done
echo "    PostgreSQL listo."

echo "==> [PRE] Esperando a que la aplicación responda (hasta ${WAIT_APP_SECONDS}s)..."
READY=0
for i in $(seq 1 $((WAIT_APP_SECONDS / 2))); do
  if [ "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/categorias")" = "200" ]; then
    READY=1; break
  fi
  sleep 2
done
if [ "$READY" != 1 ]; then
  docker compose logs --tail 30 app
  die "La aplicación no respondió a tiempo."
fi
echo "    Aplicación lista."

# -------------------------------------------------------------------- pruebas
echo
echo "==> [TEST] Categorías"
check "Crear categoría"            201 POST   "$BASE_URL/categorias" '{"nombre":"Trabajo","descripcion":"Tareas del trabajo"}'
CAT_ID=$(extract_id id_categoria)
[ -n "$CAT_ID" ] || die "No se pudo obtener el id de la categoría creada."
body_has '"nombre":"Trabajo"'
check "Crear segunda categoría"    201 POST   "$BASE_URL/categorias" '{"nombre":"Personal","descripcion":"Tareas personales"}'
check "Listar categorías"          200 GET    "$BASE_URL/categorias"
body_has '"Personal"'
check "Obtener categoría por id"   200 GET    "$BASE_URL/categorias/$CAT_ID"
check "Actualizar categoría"       200 PUT    "$BASE_URL/categorias/$CAT_ID" '{"nombre":"Trabajo Actualizado","descripcion":"Nueva descripcion"}'
body_has '"Trabajo Actualizado"'
check "Crear sin nombre (validación)" 400 POST "$BASE_URL/categorias" '{"nombre":"","descripcion":"x"}'
check "JSON inválido"              400 POST   "$BASE_URL/categorias" 'esto no es json'
check "ID no numérico"             400 GET    "$BASE_URL/categorias/abc"
check "ID inexistente"             404 GET    "$BASE_URL/categorias/999999"
check "Actualizar inexistente"     404 PUT    "$BASE_URL/categorias/999999" '{"nombre":"X","descripcion":"Y"}'
check "Eliminar categoría"         204 DELETE "$BASE_URL/categorias/$CAT_ID"
check "Obtener eliminada"          404 GET    "$BASE_URL/categorias/$CAT_ID"

echo
echo "==> [TEST] Notas"
check "Crear nota"                 201 POST   "$BASE_URL/notas" '{"titulo":"Comprar leche","contenido":"Ir al supermercado","is_archived":false}'
NOTA_ID=$(extract_id id_nota)
[ -n "$NOTA_ID" ] || die "No se pudo obtener el id de la nota creada."
body_has '"titulo":"Comprar leche"'
check "Listar notas"               200 GET    "$BASE_URL/notas"
body_has '"Comprar leche"'
check "Obtener nota por id"        200 GET    "$BASE_URL/notas/$NOTA_ID"
check "Actualizar nota"            200 PUT    "$BASE_URL/notas/$NOTA_ID" '{"titulo":"Comprar leche y pan","contenido":"Supermercado en la tarde","is_archived":true}'
body_has '"titulo":"Comprar leche y pan"'
body_has '"is_archived":true'
check "Crear sin título (validación)" 400 POST "$BASE_URL/notas" '{"titulo":"","contenido":"x"}'
check "JSON inválido"              400 POST   "$BASE_URL/notas" 'esto no es json'
check "ID no numérico"             400 GET    "$BASE_URL/notas/abc"
check "ID inexistente"             404 GET    "$BASE_URL/notas/999999"
check "Actualizar inexistente"     404 PUT    "$BASE_URL/notas/999999" '{"titulo":"X","contenido":"Y","is_archived":false}'
check "Eliminar nota"              204 DELETE "$BASE_URL/notas/$NOTA_ID"
check "Obtener eliminada"          404 GET    "$BASE_URL/notas/$NOTA_ID"

# ------------------------------------------------------------------ resultado
echo
echo "=========================================="
echo " Resultado: $PASSED OK, $FAILED fallidas"
echo "=========================================="
[ "$FAILED" -eq 0 ]
