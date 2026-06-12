#!/bin/sh

URL=${POSTGRES_URL:-postgres://authuser:authpass@postgres:5432/authdb?sslmode=disable}

goose -dir /migrations postgres "$URL" up

echo "Migrations completed"