# Etapa 1: Build del binario
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Instalamos templ para compilar las vistas
RUN go install github.com/a-h/templ/cmd/templ@v0.2.793

# Copiamos código fuente
COPY views/ ./views/
COPY backend/ ./backend/

# Generamos las vistas y compilamos de forma estática
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./backend

# Etapa 2: Imagen final minimalista
FROM alpine:latest

WORKDIR /app

# Copiamos el ejecutable desde la etapa builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
