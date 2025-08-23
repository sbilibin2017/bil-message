package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RegisterSuite struct {
	suite.Suite

	container   testcontainers.Container
	postgresURL string

	serverPath string
	clientPath string
	serverCmd  *exec.Cmd
	address    string
}

// getFreePort находит свободный TCP порт на локальной машине
func getFreePort() (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port), nil
}

func (s *RegisterSuite) SetupSuite() {
	ctx := context.Background()

	log.Println("[SetupSuite] Компиляция серверного бинарника...")
	s.serverPath = filepath.Join(os.TempDir(), "bil-server-test")
	s.clientPath = filepath.Join(os.TempDir(), "bil-client-test")

	// Путь от tests/e2e к корню проекта
	serverBuild := exec.Command("go", "build", "-o", s.serverPath, "../../cmd/server")
	serverBuild.Stdout = os.Stdout
	serverBuild.Stderr = os.Stderr
	s.Require().NoError(serverBuild.Run())
	log.Println("[SetupSuite] Серверный бинарник скомпилирован:", s.serverPath)

	clientBuild := exec.Command("go", "build", "-o", s.clientPath, "../../cmd/client")
	clientBuild.Stdout = os.Stdout
	clientBuild.Stderr = os.Stderr
	s.Require().NoError(clientBuild.Run())
	log.Println("[SetupSuite] Клиентский бинарник скомпилирован:", s.clientPath)

	log.Println("[SetupSuite] Запуск контейнера PostgreSQL...")
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = container
	log.Println("[SetupSuite] Контейнер PostgreSQL поднят")

	mappedPort, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)
	hostIP, err := container.Host(ctx)
	s.Require().NoError(err)

	s.postgresURL = fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", hostIP, mappedPort.Port())
	log.Println("[SetupSuite] Postgres URL:", s.postgresURL)

	// Выполнение миграций Goose
	log.Println("[SetupSuite] Запуск миграций Goose...")
	dbConn, err := sql.Open("pgx", s.postgresURL)
	s.Require().NoError(err)
	defer dbConn.Close()

	projectRoot, err := filepath.Abs("../../")
	s.Require().NoError(err)
	migrationsDir := filepath.Join(projectRoot, "migrations")
	s.Require().NoError(goose.Up(dbConn, migrationsDir), "failed to run migrations")
	log.Println("[SetupSuite] Миграции выполнены")

	// Свободный порт для сервера
	s.address, err = getFreePort()
	s.Require().NoError(err)
	log.Println("[SetupSuite] Сервер будет запущен на порту:", s.address)

	// Запуск сервера
	log.Println("[SetupSuite] Запуск сервера...")
	s.serverCmd = exec.Command(s.serverPath,
		"--address", s.address,
		"--database-dsn", s.postgresURL,
	)
	s.serverCmd.Stdout = os.Stdout
	s.serverCmd.Stderr = os.Stderr
	s.Require().NoError(s.serverCmd.Start())
	log.Println("[SetupSuite] Сервер запущен")

	time.Sleep(2 * time.Second) // ждём пока сервер поднимется
}

func (s *RegisterSuite) TearDownSuite() {
	ctx := context.Background()
	log.Println("[TearDownSuite] Завершение работы сервера и контейнера...")
	if s.serverCmd != nil && s.serverCmd.Process != nil {
		_ = s.serverCmd.Process.Kill()
	}
	if s.container != nil {
		_ = s.container.Terminate(ctx)
	}
	log.Println("[TearDownSuite] Завершение завершено")
}

// TestRegisterUser — e2e тест регистрации пользователя через клиента
func (s *RegisterSuite) TestRegisterUser() {
	username := "testuser"
	password := "password123"

	log.Println("[TestRegisterUser] Запуск клиентского бинарника для регистрации...")
	// Добавляем /api/v1 к адресу сервера
	cmd := exec.Command(s.clientPath,
		"register",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
	)
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err, "client command failed: %s", string(out))
	log.Println("[TestRegisterUser] Клиентский бинарник завершён. Output:", string(out))

	// Проверяем, что пользователь зарегистрирован в базе
	log.Println("[TestRegisterUser] Проверка пользователя в базе...")
	dbConn, err := sql.Open("pgx", s.postgresURL)
	s.Require().NoError(err)
	defer dbConn.Close()

	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1", username).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)
	log.Println("[TestRegisterUser] Пользователь успешно зарегистрирован")
}

// Запуск Suite
func TestRegisterSuite(t *testing.T) {
	suite.Run(t, new(RegisterSuite))
}
