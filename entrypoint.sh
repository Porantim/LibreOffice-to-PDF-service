#!/bin/bash
set -e

echo "Iniciando Unoserver..."

python3 -m unoserver.server \
    --interface 127.0.0.1 \
    --port 2003 \
    --uno-interface 127.0.0.1 \
    --uno-port 2002 \
    --quiet &
UNOSERVER_PID=$!

# Espera breve para garantir que o listener esteja pronto
sleep 5

echo "Iniciando REST API..."
exec /app/converter