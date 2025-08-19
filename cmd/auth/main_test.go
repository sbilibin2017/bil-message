package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/configs/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPrintBuildInfo(t *testing.T) {
	// Перенаправим stdout в буфер
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	buildVersion = "v1.2.3"
	buildCommit = "abc123"
	buildDate = "2025-08-19"

	printBuildInfo()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	assert.Contains(t, output, "Version : v1.2.3")
	assert.Contains(t, output, "Commit  : abc123")
	assert.Contains(t, output, "Date    : 2025-08-19")
}

func TestParseFlags(t *testing.T) {
	// Сохраним оригинальные os.Args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"cmd",
		"--address", "127.0.0.1:9000",
		"--dsn", "sqlite://:memory:",
		"--jwt-key", "testkey",
		"--jwt-expiration", "2h",
	}

	err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9000", address)
	assert.Equal(t, "sqlite://:memory:", databaseDSN)
	assert.Equal(t, "testkey", jwtSecretKey)
	assert.Equal(t, 2*time.Hour, jwtExp)
}

func TestRunWithSQLite(t *testing.T) {
	// Используем SQLite in-memory для теста
	driver := "sqlite"
	dsn := ":memory:"     // только :memory:, без sqlite://
	addr := "127.0.0.1:0" // случайный порт
	jwtKey := "testkey"
	jwtExpiration := time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Запускаем сервер в отдельной горутине, чтобы не блокировать тест
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, driver, dsn, addr, jwtKey, jwtExpiration)
	}()

	select {
	case <-ctx.Done():
		// таймаут
	case err := <-errCh:
		assert.NoError(t, err)
	}
}

type AuthSuite struct {
	suite.Suite
	postgresC  tc.Container
	serverURL  string
	httpClient *resty.Client
}

func (s *AuthSuite) SetupSuite() {
	ctx := context.Background()

	// запускаем postgres
	req := tc.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "bil_message_user",
			"POSTGRES_PASSWORD": "bil_message_password",
			"POSTGRES_DB":       "bil_message_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	var err error
	s.postgresC, err = tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)

	host, err := s.postgresC.Host(ctx)
	s.Require().NoError(err)
	port, err := s.postgresC.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	dsn := fmt.Sprintf("postgres://bil_message_user:bil_message_password@%s:%s/bil_message_db?sslmode=disable", host, port.Port())

	// подключаемся к БД
	conn, err := db.New("pgx", dsn)
	s.Require().NoError(err)

	// запускаем миграции напрямую
	s.Require().NoError(runMigrations(ctx, conn))

	// запускаем сервер
	s.serverURL = "http://127.0.0.1:18080"
	go func() {
		err := run(
			ctx,
			"pgx",
			dsn,
			"127.0.0.1:18080",
			"test-jwt-secret", // <- ключ JWT
			1*time.Hour,
		)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)
	s.httpClient = resty.New().SetBaseURL(s.serverURL + "/api/v1")
}

func (s *AuthSuite) TearDownSuite() {
	if s.postgresC != nil {
		_ = s.postgresC.Terminate(context.Background())
	}
}

// runMigrations выполняет все миграции напрямую через SQL
func runMigrations(ctx context.Context, conn *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			public_key TEXT,
			created_at TIMESTAMP DEFAULT now() NOT NULL,
			updated_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS devices (
			device_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
			public_key TEXT,
			created_at TIMESTAMP DEFAULT now() NOT NULL,
			updated_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS chat_types (
			chat_type_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS chats (
			chat_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100),
			type_uuid UUID NOT NULL REFERENCES chat_types(chat_type_uuid),
			created_by_uuid UUID NOT NULL REFERENCES users(user_uuid),
			created_at TIMESTAMP DEFAULT now() NOT NULL,
			updated_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS chat_members (
			chat_member_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			chat_uuid UUID NOT NULL REFERENCES chats(chat_uuid) ON DELETE CASCADE,
			user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
			role VARCHAR(20),
			joined_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			message_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			chat_uuid UUID NOT NULL REFERENCES chats(chat_uuid) ON DELETE CASCADE,
			sender_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
			encrypted_text TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS message_keys (
			message_key_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			message_uuid UUID NOT NULL REFERENCES messages(message_uuid) ON DELETE CASCADE,
			device_uuid UUID NOT NULL REFERENCES devices(device_uuid) ON DELETE CASCADE,
			encrypted_symmetric_key TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now() NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS message_reads (
			message_read_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			message_uuid UUID NOT NULL REFERENCES messages(message_uuid) ON DELETE CASCADE,
			user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
			read_at TIMESTAMP DEFAULT now()
		);`,
	}

	for _, q := range migrations {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// Тесты регистрации
// Тесты регистрации
func (s *AuthSuite) TestRegisterSuccess() {
	resp, err := s.httpClient.R().
		SetBody(map[string]string{
			"username": "newuser",
			"password": "Password123!",
		}).
		Post("/auth/register") // <- исправлено на фактический путь

	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())

	authHeader := resp.Header().Get("Authorization")
	s.NotEmpty(authHeader)
	s.Contains(authHeader, "Bearer ")

	// Проверяем, что тело не пустое
	s.Empty(resp.Body())
}

func (s *AuthSuite) TestRegisterDuplicate() {
	// Сначала создаём пользователя
	resp1, err := s.httpClient.R().
		SetBody(map[string]string{
			"username": "duplicateuser",
			"password": "Password123!",
		}).
		Post("/auth/register")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp1.StatusCode())

	// Повторная регистрация должна вернуть 409
	resp2, err := s.httpClient.R().
		SetBody(map[string]string{
			"username": "duplicateuser",
			"password": "Password123!",
		}).
		Post("/auth/register")
	s.Require().NoError(err)
	s.Equal(http.StatusConflict, resp2.StatusCode())
}

func (s *AuthSuite) TestRegisterInvalidBody() {
	resp, err := s.httpClient.R().
		SetBody(`invalid-json`).
		Post("/auth/register")
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode())
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
