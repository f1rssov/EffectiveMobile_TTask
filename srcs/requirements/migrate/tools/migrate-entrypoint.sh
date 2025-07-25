#!/bin/sh
set -e

until pg_isready -h "$DB_HOST" -U "$DB_USER"; do
  echo "PostgreSQL пока не доступен, ждем..."
  sleep 2
done

echo "PostgreSQL доступен, запускаем миграции..."

migrate -path=/migrations -database "postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable" up

echo "Миграции выполнены, завершаем контейнер."
