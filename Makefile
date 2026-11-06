.PHONY: help templ sqlc build up down run run-app test clean

TEMPL ?= $(shell which templ 2>/dev/null || echo $(HOME)/go/bin/templ)

help: ## Muestra los targets disponibles
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/ -/'

templ: ## Genera código Go a partir de las plantillas templ
	@if [ -x "$$(command -v $(TEMPL))" ]; then \
		$(TEMPL) generate; \
	else \
		go run github.com/a-h/templ/cmd/templ@v0.2.793 generate; \
	fi

sqlc: ## Genera el código Go con sqlc
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	elif [ -f "$(HOME)/go/bin/sqlc" ]; then \
		$(HOME)/go/bin/sqlc generate; \
	elif docker run --rm -v "$(CURDIR)":/src -w /src sqlc/sqlc generate 2>/dev/null; then \
		: ; \
	elif [ -f "backend/db/queries.sql.go" ]; then \
		echo "-> Aviso: sqlc vía Docker no disponible, usando código existente en backend/db"; \
	else \
		docker run --rm -v "$(CURDIR)":/src -w /src sqlc/sqlc generate; \
	fi

build: templ sqlc ## Genera con templ y sqlc y compila
	go build ./...

run: ## Ciclo completo: tareas previas, ejecución de la app y limpieza posterior
	./run.sh

run-app: run ## Alias de run

up: ## Levanta app + base de datos en segundo plano
	docker compose up -d

down: ## Borra contenedores y volúmenes
	-docker rm -f gestor-notas-app >/dev/null 2>&1 || true
	docker compose down -v

test: ## Tests de Go contra una base de datos de testing temporal
	-docker compose -f docker-compose.test.yml down -v
	docker compose -f docker-compose.test.yml up -d
	@until docker exec gestor-notas-db-test pg_isready -U $${TEST_DB_USER:-testuser} -d $${TEST_DB_NAME:-notas_db_test} >/dev/null 2>&1; do sleep 1; done
	@CONTAINER_IP=$$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' gestor-notas-db-test 2>/dev/null || echo ""); \
	if [ -n "$$CONTAINER_IP" ]; then \
		export TEST_DB_HOST="$$CONTAINER_IP" TEST_DB_PORT="5432"; \
	fi; \
	go test -v ./views/... ./backend/handlers/... ./tests/... ; status=$$?; docker compose -f docker-compose.test.yml down -v; exit $$status

clean: down ## Alias de down
