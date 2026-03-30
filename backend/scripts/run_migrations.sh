#!/bin/bash

# Script para rodar migrations do banco de dados
# Usage: ./scripts/run_migrations.sh

set -e

cd "$(dirname "$0")/.."

echo "🔍 Verificando variáveis de ambiente..."

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    echo "Please create .env file with database credentials."
    exit 1
fi

echo "✅ .env file found"

echo "🚀 Running database migrations..."

# Run migrations using the Go binary
go run cmd/api/main.go --migrations-only

echo "✅ Migrations completed!"
