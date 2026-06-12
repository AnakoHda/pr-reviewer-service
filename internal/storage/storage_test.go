package storage

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	// 1. Поднимаем контейнер Postgres
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// 2. Подключаемся к базе
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	// 3. Находим путь к миграциям
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsDir := filepath.Join(basepath, "../../migrations")

	// 4. Накатываем миграции через goose
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Функция для очистки ресурсов
	cleanup := func() {
		_ = db.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestConnection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	var result int
	err := db.Get(&result, "SELECT 1")
	if err != nil {
		t.Errorf("failed to execute simple query: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}
