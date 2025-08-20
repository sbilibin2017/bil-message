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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/configs/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPrintBuildInfo(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	buildVersion = "v1.2.3"
	buildCommit = "abc123"
	buildDate = "2025-08-20"

	printBuildInfo()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	assert.Contains(t, output, "Version : v1.2.3")
	assert.Contains(t, output, "Commit  : abc123")
	assert.Contains(t, output, "Date    : 2025-08-20")
}

func TestParseFlags(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"cmd",
		"--address", "127.0.0.1:9001",
		"--dsn", "sqlite://:memory:",
	}

	err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9001", address)
	assert.Equal(t, "sqlite://:memory:", databaseDSN)
}

func TestRunWithSQLite(t *testing.T) {
	driver := "sqlite"
	dsn := ":memory:"
	addr := "127.0.0.1:0"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, driver, dsn, addr)
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		assert.NoError(t, err)
	}
}

type DeviceSuite struct {
	suite.Suite
	postgresC  tc.Container
	serverURL  string
	httpClient *resty.Client
}

func (s *DeviceSuite) SetupSuite() {
	ctx := context.Background()

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

	conn, err := db.New("pgx", dsn)
	s.Require().NoError(err)

	// Создаем расширение pgcrypto для gen_random_uuid()
	_, _ = conn.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`)

	s.Require().NoError(runDeviceMigrations(ctx, conn))

	s.serverURL = "http://127.0.0.1:18081"
	go func() {
		if err := run(ctx, "pgx", dsn, "127.0.0.1:18081"); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	s.httpClient = resty.New().SetBaseURL(s.serverURL + "/api/v1")
}

func (s *DeviceSuite) TearDownSuite() {
	if s.postgresC != nil {
		_ = s.postgresC.Terminate(context.Background())
	}
}

func runDeviceMigrations(ctx context.Context, conn *sqlx.DB) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`, // чтобы использовать UUID генерацию, если нужно
		`CREATE TABLE IF NOT EXISTS users (
			user_uuid UUID PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL DEFAULT '',
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
	}

	for _, q := range queries {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func (s *DeviceSuite) TestDeviceRegisterSuccess() {
	ctx := context.Background()

	host, _ := s.postgresC.Host(ctx)
	port, _ := s.postgresC.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://bil_message_user:bil_message_password@%s:%s/bil_message_db?sslmode=disable", host, port.Port())
	conn, err := db.New("pgx", dsn)
	s.Require().NoError(err)

	userUUID := uuid.New()
	_, err = conn.ExecContext(ctx, `INSERT INTO users (user_uuid, username) VALUES ($1, $2)`, userUUID, "testuser")
	s.Require().NoError(err)

	resp, err := s.httpClient.R().
		SetBody(map[string]string{
			"user_uuid":  userUUID.String(),
			"public_key": "test-public-key",
		}).
		Post("/devices/register")
	s.Require().NoError(err)

	s.Equal(http.StatusOK, resp.StatusCode())
	body := resp.String()
	s.NotEmpty(body)
	_, parseErr := uuid.Parse(body)
	s.NoError(parseErr)
}

func (s *DeviceSuite) TestDeviceRegisterInvalidBody() {
	resp, err := s.httpClient.R().
		SetBody(`invalid-json`).
		Post("/devices/register")
	s.Require().NoError(err)

	s.Equal(http.StatusBadRequest, resp.StatusCode())
}

func TestDeviceSuite(t *testing.T) {
	suite.Run(t, new(DeviceSuite))
}
