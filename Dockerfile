# Etapa 1: Build del binario
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

# Copiamos solo el código fuente del backend para compilar
COPY backend/ ./backend/

# Compilamos el binario de forma estática
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./backend

# Etapa 2: Imagen final minimalista
FROM alpine:latest

WORKDIR /app

# Copiamos el ejecutable desde la etapa builder
COPY --from=builder /app/server .

# Copiamos la carpeta frontend directamente desde el contexto raíz
COPY frontend/ ./frontend/

EXPOSE 8080

CMD ["./server"]