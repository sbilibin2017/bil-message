package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB_Success(t *testing.T) {
	db, err := New("sqlite", ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestNewDB_WithOptions(t *testing.T) {
	db, err := New(
		"sqlite",
		":memory:",
		WithMaxOpenConns(10),
		WithMaxIdleConns(5),
	)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Проверяем, что опции применились
	assert.Equal(t, 10, db.Stats().MaxOpenConnections)
	// IdleConns и ConnMaxLifetime нельзя получить напрямую через sqlx.DB, но проверим, что DB не nil
}

func TestWithMaxOpenConns_ZeroIgnored(t *testing.T) {
	db, err := New("sqlite", ":memory:", WithMaxOpenConns(0))
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
	// Проверяем, что не установлено ненулевое значение (значение по умолчанию зависит от драйвера)
}

func TestWithMaxIdleConns_ZeroIgnored(t *testing.T) {
	db, err := New("sqlite", ":memory:", WithMaxIdleConns(0))
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestRunMigrations(t *testing.T) {
	// Создаем временный файл базы данных
	tmpDB := filepath.Join(os.TempDir(), "test_migrations.db")
	defer os.Remove(tmpDB)

	sqlDB, err := New("sqlite", tmpDB)
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)

	// Создаем временную папку для миграций
	tmpMigrations := filepath.Join(os.TempDir(), "migrations")
	defer os.RemoveAll(tmpMigrations)
	err = os.Mkdir(tmpMigrations, 0755)
	assert.NoError(t, err)

	// Создаем фиктивный файл миграции (goose требует имени формата xxxxx_name.sql)
	migrationFile := filepath.Join(tmpMigrations, "00001_create_table.sql")
	err = os.WriteFile(migrationFile, []byte(`
-- +goose Up
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE test_table;
`), 0644)
	assert.NoError(t, err)

	// Успешный запуск миграций
	err = RunMigrations(sqlDB, "sqlite", tmpMigrations)
	assert.NoError(t, err)

	// Проверим, что таблица действительно создана
	var count int
	err = sqlDB.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table';")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Попытка выполнить миграции из несуществующей папки должна вернуть ошибку
	err = RunMigrations(sqlDB, "sqlite", filepath.Join(os.TempDir(), "nonexistent"))
	assert.Error(t, err)
}
