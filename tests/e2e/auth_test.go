package e2e

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

type AuthSuite struct {
	suite.Suite

	container   testcontainers.Container
	postgresURL string

	serverPath string
	clientPath string
	serverCmd  *exec.Cmd
	address    string
}

func (s *AuthSuite) SetupSuite() {
	ctx := context.Background()

	log.Println("[SetupSuite] Компиляция бинарников...")

	// Генерация случайного хэша для бинарников
	binHash := func() string {
		b := make([]byte, 4) // 8 hex символов
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		return hex.EncodeToString(b)
	}

	s.serverPath = filepath.Join(os.TempDir(), fmt.Sprintf("bil-server-%s", binHash()))
	s.clientPath = filepath.Join(os.TempDir(), fmt.Sprintf("bil-client-%s", binHash()))

	// Компиляция серверного бинарника
	serverBuild := exec.Command("go", "build", "-o", s.serverPath, "../../cmd/server")
	serverBuild.Stdout = os.Stdout
	serverBuild.Stderr = os.Stderr
	s.Require().NoError(serverBuild.Run())
	log.Println("[SetupSuite] Серверный бинарник скомпилирован:", s.serverPath)

	// Компиляция клиентского бинарника
	clientBuild := exec.Command("go", "build", "-o", s.clientPath, "../../cmd/client")
	clientBuild.Stdout = os.Stdout
	clientBuild.Stderr = os.Stderr
	s.Require().NoError(clientBuild.Run())
	log.Println("[SetupSuite] Клиентский бинарник скомпилирован:", s.clientPath)

	// Запуск PostgreSQL контейнера
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

	// Поиск свободного порта для сервера
	l, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	s.address = fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
	l.Close()
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

func (s *AuthSuite) TearDownSuite() {
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
func (s *AuthSuite) TestRegisterUser() {
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	password := "password123"

	log.Println("[TestRegisterUser] Запуск клиентского бинарника для регистрации...")
	cmd := exec.Command(s.clientPath,
		"register",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
	)
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err, "client command failed: %s", string(out))
	log.Println("[TestRegisterUser] Клиентский бинарник завершён. Output:", string(out))

	dbConn, err := sql.Open("pgx", s.postgresURL)
	s.Require().NoError(err)
	defer dbConn.Close()

	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1", username).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)
	log.Println("[TestRegisterUser] Пользователь успешно зарегистрирован")
}

func (s *AuthSuite) TestAddDevice() {
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	password := "password123"
	publicKey := "publickey123"

	// Регистрация нового пользователя
	cmdRegister := exec.Command(s.clientPath,
		"register",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
	)
	out, err := cmdRegister.CombinedOutput()
	s.Require().NoError(err, "не удалось зарегистрировать пользователя: %s", string(out))

	// Добавление устройства
	cmd := exec.Command(s.clientPath,
		"device",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
		"--public-key", publicKey,
	)
	out, err = cmd.CombinedOutput()
	s.Require().NoError(err, "не удалось добавить устройство: %s", string(out))
	log.Println("[TestAddDevice] Клиентский бинарник завершён. Output:", string(out))

	configDir := os.ExpandEnv("$HOME/.config")
	deviceFile := fmt.Sprintf("%s/bil_message_client_device_uuid", configDir)
	data, err := os.ReadFile(deviceFile)
	s.Require().NoError(err, "не удалось прочитать файл с UUID устройства")
	s.Require().NotEmpty(data, "UUID устройства пустой")
	log.Println("[TestAddDevice] UUID устройства успешно сохранён:", string(data))
}

func (s *AuthSuite) TestLogin() {
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	password := "password123"
	publicKey := "publickey123"

	// Регистрация и добавление устройства
	cmdRegister := exec.Command(s.clientPath,
		"register",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
	)
	out, err := cmdRegister.CombinedOutput()
	s.Require().NoError(err, "не удалось зарегистрировать пользователя: %s", string(out))

	cmdDevice := exec.Command(s.clientPath,
		"device",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
		"--public-key", publicKey,
	)
	out, err = cmdDevice.CombinedOutput()
	s.Require().NoError(err, "не удалось добавить устройство: %s", string(out))
	log.Println("[TestLogin] Устройство добавлено. Output:", string(out))

	// Login
	cmdLogin := exec.Command(s.clientPath,
		"login",
		"--address", "http://localhost"+s.address+"/api/v1",
		"--username", username,
		"--password", password,
	)
	out, err = cmdLogin.CombinedOutput()
	s.Require().NoError(err, "не удалось выполнить вход: %s", string(out))
	token := string(out)
	s.Require().NotEmpty(token, "JWT токен пустой")
	log.Println("[TestLogin] JWT токен получен:", token)
}

// Запуск Suite
func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
