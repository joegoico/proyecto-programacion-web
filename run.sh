#!/usr/bin/env bash

# Detener la ejecución si ocurre algún error
set -e

IMAGE_NAME="notas-servidor-go"
CONTAINER_NAME="gestor-notas-app"
PORT="8080"

echo "=========================================="
echo " Levantando la aplicación con Docker...   "
echo "=========================================="

# 1. Detener y remover contenedor previo si ya estaba corriendo
if [ "$(docker ps -aq -f name=^${CONTAINER_NAME}$)" ]; then
    echo "-> Limpiando contenedor previo..."
    docker rm -f ${CONTAINER_NAME} > /dev/null 2>&1
fi

# 2. Construir la imagen Docker
echo "-> Construyendo la imagen..."
docker build -t ${IMAGE_NAME} .

# 3. Iniciar el contenedor
echo "-> Iniciando contenedor en el puerto ${PORT}..."
docker run -d -p ${PORT}:${PORT} --name ${CONTAINER_NAME} ${IMAGE_NAME}

echo "=========================================="
echo " ¡Servidor corriendo exitosamente!"
echo " Abrir en el navegador: http://localhost:${PORT}"
echo "=========================================="
echo "-> Mostrando logs del servidor (Ctrl+C para salir de los logs):"
docker logs -f ${CONTAINER_NAME}