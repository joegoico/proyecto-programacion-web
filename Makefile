.PHONY: help sqlc build up down run test test-api clean

help: ## Muestra los targets disponibles
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/ -/'

sqlc: ## Genera el código Go con sqlc (vía Docker)
	docker run --rm -v "$(CURDIR)":/src -w /src sqlc/sqlc generate

build: sqlc ## Genera con sqlc y compila
	go build ./...

up: ## Levanta app + base de datos en segundo plano
	docker compose up -d

down: ## Borra contenedores y volúmenes
	-docker rm -f gestor-notas-app >/dev/null 2>&1
	docker compose down -v

run: ## Tests de Go + levanta la app en modo desarrollo (run.sh)
	./run.sh

test: ## Tests de Go contra una base de datos de testing temporal
	-docker compose -f docker-compose.test.yml down -v
	docker compose -f docker-compose.test.yml up -d
	@until docker exec gestor-notas-db-test pg_isready -U $${TEST_DB_USER:-testuser} -d $${TEST_DB_NAME:-notas_db_test} >/dev/null 2>&1; do sleep 1; done
	go test -v ./tests/... ; status=$$?; docker compose -f docker-compose.test.yml down -v; exit $$status

test-api: ## Todo: sqlc, build, levantar, probar cada endpoint con curl, limpiar
	./test-api.sh

clean: down ## Alias de down
