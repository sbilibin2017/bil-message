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
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ChatSuite struct {
	suite.Suite

	container   testcontainers.Container
	postgresURL string

	serverPath string
	clientPath string
	serverCmd  *exec.Cmd
	address    string

	username string
	password string
	token    string
}

func (s *ChatSuite) SetupSuite() {
	ctx := context.Background()
	log.Println("[SetupSuite] Компиляция бинарников...")

	binHash := func() string {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		return hex.EncodeToString(b)
	}

	s.serverPath = filepath.Join(os.TempDir(), fmt.Sprintf("bil-server-%s", binHash()))
	s.clientPath = filepath.Join(os.TempDir(), fmt.Sprintf("bil-client-%s", binHash()))

	// Компиляция бинарников
	s.Require().NoError(exec.Command("go", "build", "-o", s.serverPath, "../../cmd/server").Run())
	s.Require().NoError(exec.Command("go", "build", "-o", s.clientPath, "../../cmd/client").Run())

	// Запуск PostgreSQL контейнера
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

	mappedPort, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)
	hostIP, err := container.Host(ctx)
	s.Require().NoError(err)

	s.postgresURL = fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", hostIP, mappedPort.Port())

	dbConn, err := sql.Open("pgx", s.postgresURL)
	s.Require().NoError(err)
	defer dbConn.Close()

	projectRoot, err := filepath.Abs("../../")
	s.Require().NoError(err)
	s.Require().NoError(goose.Up(dbConn, filepath.Join(projectRoot, "migrations")))

	// Свободный порт для сервера
	l, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	s.address = fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
	l.Close()

	// Запуск сервера с фиксированным секретом для JWT
	jwtSecret := "super-secret-key"
	s.serverCmd = exec.Command(
		s.serverPath,
		"--address", s.address,
		"--database-dsn", s.postgresURL,
		"--jwt-secret", jwtSecret,
	)
	s.serverCmd.Stdout = os.Stdout
	s.serverCmd.Stderr = os.Stderr
	s.Require().NoError(s.serverCmd.Start())

	// Ждём пока сервер поднимется
	clientAddr := fmt.Sprintf("http://localhost%s/api/v1", s.address)
	for i := 0; i < 20; i++ {
		conn, _ := net.DialTimeout("tcp", "localhost"+s.address, 500*time.Millisecond)
		if conn != nil {
			conn.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Создание пользователя и получение токена
	s.username = fmt.Sprintf("user_%d", time.Now().UnixNano())
	s.password = "password123"

	// Регистрация
	cmdRegister := exec.Command(s.clientPath, "register",
		"--address", clientAddr,
		"--username", s.username,
		"--password", s.password,
	)
	out, err := cmdRegister.CombinedOutput()
	s.Require().NoError(err, "registration failed: %s", string(out))

	// Добавление устройства
	cmdDevice := exec.Command(s.clientPath, "device",
		"--address", clientAddr,
		"--username", s.username,
		"--password", s.password,
		"--public-key", "publickey123",
	)
	out, err = cmdDevice.CombinedOutput()
	s.Require().NoError(err, "add device failed: %s", string(out))

	// Проверка файла UUID устройства
	deviceFile := filepath.Join(os.ExpandEnv("$HOME/.config"), "bil_message_client_device_uuid")
	data, err := os.ReadFile(deviceFile)
	s.Require().NoError(err, "failed to read device UUID file")
	s.Require().NotEmpty(data, "device UUID is empty")

	// Login — сохраняем **только JWT** без "Bearer "
	cmdLogin := exec.Command(s.clientPath, "login",
		"--address", clientAddr,
		"--username", s.username,
		"--password", s.password,
	)
	out, err = cmdLogin.CombinedOutput()
	s.Require().NoError(err, "login failed: %s", string(out))
	s.token = strings.TrimSpace(string(out))
	s.token = strings.TrimPrefix(s.token, "Bearer ")
	s.token = strings.TrimSpace(s.token)
	s.Require().NotEmpty(s.token, "JWT token is empty")
}

func (s *ChatSuite) TearDownSuite() {
	ctx := context.Background()
	if s.serverCmd != nil && s.serverCmd.Process != nil {
		_ = s.serverCmd.Process.Kill()
	}
	if s.container != nil {
		_ = s.container.Terminate(ctx)
	}
}

// Полный жизненный цикл чата в одном тесте
func (s *ChatSuite) TestChatLifecycle() {
	clientAddr := fmt.Sprintf("http://localhost%s/api/v1", s.address)

	// Создание чата
	cmdCreate := exec.Command(s.clientPath, "create",
		"--address", clientAddr,
		"--token", s.token,
	)
	out, err := cmdCreate.CombinedOutput()
	s.Require().NoError(err, "create chat failed: %s", string(out))
	roomID := strings.TrimSpace(string(out))
	s.Require().NotEmpty(roomID, "room ID is empty")

	// Добавление текущего пользователя
	cmdAdd := exec.Command(s.clientPath, "add-member",
		"--address", clientAddr,
		"--token", s.token,
		"--chat-uuid", roomID,
	)
	out, err = cmdAdd.CombinedOutput()
	s.Require().NoError(err, "add member failed: %s", string(out))

	// Удаление текущего пользователя
	cmdRemoveMember := exec.Command(s.clientPath, "remove-member",
		"--address", clientAddr,
		"--token", s.token,
		"--chat-uuid", roomID,
	)
	out, err = cmdRemoveMember.CombinedOutput()
	s.Require().NoError(err, "remove member failed: %s", string(out))

	// Удаление чата
	cmdRemove := exec.Command(s.clientPath, "remove",
		"--address", clientAddr,
		"--token", s.token,
		"--chat-uuid", roomID,
	)
	out, err = cmdRemove.CombinedOutput()
	s.Require().NoError(err, "remove chat failed: %s", string(out))
}

func TestChatSuite(t *testing.T) {
	suite.Run(t, new(ChatSuite))
}
