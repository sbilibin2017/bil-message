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

	"github.com/gorilla/websocket"
)

type ChatSuite struct {
	suite.Suite

	container   testcontainers.Container
	postgresURL string

	serverPath string
	clientPath string
	serverCmd  *exec.Cmd
	address    string

	username1 string
	username2 string
	password  string
	token1    string
	token2    string
	userUUID1 string
	userUUID2 string
}

// SetupSuite запускает сервер, контейнер PostgreSQL и создаёт двух пользователей
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

	// Компиляция серверного бинарника
	log.Println("[SetupSuite] Компиляция сервера...")
	serverBuild := exec.Command("go", "build", "-o", s.serverPath, "../../cmd/server")
	serverBuild.Stdout = os.Stdout
	serverBuild.Stderr = os.Stderr
	s.Require().NoError(serverBuild.Run())
	log.Printf("[SetupSuite] Сервер скомпилирован: %s", s.serverPath)

	// Компиляция клиентского бинарника
	log.Println("[SetupSuite] Компиляция клиента...")
	clientBuild := exec.Command("go", "build", "-o", s.clientPath, "../../cmd/client")
	clientBuild.Stdout = os.Stdout
	clientBuild.Stderr = os.Stderr
	s.Require().NoError(clientBuild.Run())
	log.Printf("[SetupSuite] Клиент скомпилирован: %s", s.clientPath)

	// Запуск контейнера PostgreSQL
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
	mappedPort, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)
	hostIP, err := container.Host(ctx)
	s.Require().NoError(err)
	s.postgresURL = fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", hostIP, mappedPort.Port())
	log.Printf("[SetupSuite] PostgreSQL URL: %s", s.postgresURL)

	// Выполнение миграций Goose
	log.Println("[SetupSuite] Применение миграций...")
	dbConn, err := sql.Open("pgx", s.postgresURL)
	s.Require().NoError(err)
	defer dbConn.Close()
	projectRoot, err := filepath.Abs("../../")
	s.Require().NoError(err)
	migrationsDir := filepath.Join(projectRoot, "migrations")
	s.Require().NoError(goose.Up(dbConn, migrationsDir))
	log.Println("[SetupSuite] Миграции применены")

	// Поиск свободного порта для сервера
	l, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	s.address = fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
	l.Close()
	log.Printf("[SetupSuite] Выбран порт для сервера: %s", s.address)

	// Запуск сервера
	log.Println("[SetupSuite] Запуск сервера...")
	jwtSecret := "super-secret-key"
	s.serverCmd = exec.Command(s.serverPath,
		"--address", s.address,
		"--database-dsn", s.postgresURL,
		"--jwt-secret", jwtSecret,
	)
	s.serverCmd.Stdout = os.Stdout
	s.serverCmd.Stderr = os.Stderr
	s.Require().NoError(s.serverCmd.Start())
	log.Printf("[SetupSuite] Сервер запущен, PID: %d", s.serverCmd.Process.Pid)
	time.Sleep(2 * time.Second)

	clientAddr := fmt.Sprintf("http://localhost%s/api/v1", s.address)

	// Регистрация первого пользователя
	log.Println("[SetupSuite] Регистрация первого пользователя...")
	s.username1 = fmt.Sprintf("user1_%d", time.Now().UnixNano())
	s.password = "password123"
	out, err := exec.Command(s.clientPath, "register",
		"--address", clientAddr,
		"--username", s.username1,
		"--password", s.password,
	).CombinedOutput()
	s.Require().NoError(err)
	s.userUUID1 = strings.TrimSpace(string(out))
	log.Printf("[SetupSuite] Первый пользователь зарегистрирован, UUID: %q", s.userUUID1)

	// Добавление устройства и логин первого пользователя
	log.Println("[SetupSuite] Добавление устройства для первого пользователя...")
	out, err = exec.Command(s.clientPath, "device",
		"--address", clientAddr,
		"--username", s.username1,
		"--password", s.password,
		"--public-key", "publickey123",
	).CombinedOutput()
	s.Require().NoError(err)
	log.Println("[SetupSuite] Устройство добавлено")

	log.Println("[SetupSuite] Логин первого пользователя...")
	out, err = exec.Command(s.clientPath, "login",
		"--address", clientAddr,
		"--username", s.username1,
		"--password", s.password,
	).CombinedOutput()
	s.Require().NoError(err)
	s.token1 = strings.TrimSpace(string(out))
	s.token1 = strings.TrimPrefix(s.token1, "Bearer ")
	s.token1 = strings.TrimSpace(s.token1)
	log.Printf("[SetupSuite] Первый пользователь залогинен, токен: %q", s.token1)

	// Регистрация второго пользователя
	log.Println("[SetupSuite] Регистрация второго пользователя...")
	s.username2 = fmt.Sprintf("user2_%d", time.Now().UnixNano())
	out, err = exec.Command(s.clientPath, "register",
		"--address", clientAddr,
		"--username", s.username2,
		"--password", s.password,
	).CombinedOutput()
	s.Require().NoError(err)
	s.userUUID2 = strings.TrimSpace(string(out))
	log.Printf("[SetupSuite] Второй пользователь зарегистрирован, UUID: %q", s.userUUID2)

	// Добавление устройства и логин второго пользователя
	log.Println("[SetupSuite] Добавление устройства для второго пользователя...")
	out, err = exec.Command(s.clientPath, "device",
		"--address", clientAddr,
		"--username", s.username2,
		"--password", s.password,
		"--public-key", "publickey123",
	).CombinedOutput()
	s.Require().NoError(err)
	log.Println("[SetupSuite] Устройство второго пользователя добавлено")

	log.Println("[SetupSuite] Логин второго пользователя...")
	out, err = exec.Command(s.clientPath, "login",
		"--address", clientAddr,
		"--username", s.username2,
		"--password", s.password,
	).CombinedOutput()
	s.Require().NoError(err)
	s.token2 = strings.TrimSpace(string(out))
	s.token2 = strings.TrimPrefix(s.token2, "Bearer ")
	s.token2 = strings.TrimSpace(s.token2)
	log.Printf("[SetupSuite] Второй пользователь залогинен, токен: %q", s.token2)
}

func (s *ChatSuite) TearDownSuite() {
	ctx := context.Background()
	log.Println("[TearDownSuite] Завершение сервера и контейнера...")
	if s.serverCmd != nil && s.serverCmd.Process != nil {
		_ = s.serverCmd.Process.Kill()
	}
	if s.container != nil {
		_ = s.container.Terminate(ctx)
	}
	log.Println("[TearDownSuite] Завершение выполнено")
}

// Полный жизненный цикл чата (CRUD)
func (s *ChatSuite) TestChatLifecycle() {
	clientAddr := fmt.Sprintf("http://localhost%s/api/v1", s.address)

	// Создание новой комнаты первым пользователем
	log.Println("[TestChatLifecycle] Создание комнаты первым пользователем...")
	out, err := exec.Command(s.clientPath, "create",
		"--address", clientAddr,
		"--token", s.token1,
	).CombinedOutput()
	log.Printf("[TestChatLifecycle] create output: %q", string(out))
	s.Require().NoError(err)
	roomID := strings.TrimSpace(string(out))
	s.Require().NotEmpty(roomID)
	log.Printf("[TestChatLifecycle] Комната создана, UUID: %s", roomID)

	// Добавление второго пользователя в комнату
	log.Println("[TestChatLifecycle] Добавление второго пользователя в комнату...")
	cmd := exec.Command(s.clientPath, "add-member",
		"--address", clientAddr,
		"--token", s.token1,
		"--room-uuid", roomID,
		"--member-uuid", s.userUUID2,
	)
	out, err = cmd.CombinedOutput()
	log.Printf("[TestChatLifecycle] add-member output: %q", string(out))
	s.Require().NoError(err)
	log.Println("[TestChatLifecycle] Второй пользователь добавлен в комнату")

	// Удаление второго пользователя из комнаты
	log.Println("[TestChatLifecycle] Удаление второго пользователя из комнаты...")
	cmd = exec.Command(s.clientPath, "remove-member",
		"--address", clientAddr,
		"--token", s.token1,
		"--room-uuid", roomID,
		"--member-uuid", s.userUUID2,
	)
	out, err = cmd.CombinedOutput()
	log.Printf("[TestChatLifecycle] remove-member output: %q", string(out))
	s.Require().NoError(err)
	log.Println("[TestChatLifecycle] Второй пользователь удалён из комнаты")

	// Удаление комнаты
	log.Println("[TestChatLifecycle] Удаление комнаты...")
	cmd = exec.Command(s.clientPath, "remove",
		"--address", clientAddr,
		"--token", s.token1,
		"--room-uuid", roomID,
	)
	out, err = cmd.CombinedOutput()
	log.Printf("[TestChatLifecycle] remove output: %q", string(out))
	s.Require().NoError(err)
	log.Println("[TestChatLifecycle] Комната удалена")
}

// Тест обмена сообщениями между двумя пользователями через WebSocket
func (s *ChatSuite) TestMessaging() {
	clientAddr := fmt.Sprintf("http://localhost%s/api/v1", s.address)

	// Создание комнаты
	log.Println("[TestMessaging] Создание комнаты...")
	out, err := exec.Command(s.clientPath, "create",
		"--address", clientAddr,
		"--token", s.token1,
	).CombinedOutput()
	s.Require().NoError(err, string(out))
	roomID := strings.TrimSpace(string(out))
	s.Require().NotEmpty(roomID)
	log.Printf("[TestMessaging] Комната создана: %s", roomID)

	// Добавление второго пользователя
	log.Println("[TestMessaging] Добавление второго пользователя...")
	_, err = exec.Command(s.clientPath, "add-member",
		"--address", clientAddr,
		"--token", s.token1,
		"--room-uuid", roomID,
		"--member-uuid", s.userUUID2,
	).CombinedOutput()
	s.Require().NoError(err)

	// WebSocket URL
	wsURL := fmt.Sprintf("ws://localhost%s/api/v1/chat/%s/ws", s.address, roomID)

	// Подключение первого пользователя
	log.Println("[TestMessaging] Подключение первого пользователя...")
	header1 := map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", s.token1)},
	}
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, header1)
	s.Require().NoError(err)
	defer conn1.Close()

	// Подключение второго пользователя
	log.Println("[TestMessaging] Подключение второго пользователя...")
	header2 := map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", s.token2)},
	}
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, header2)
	s.Require().NoError(err)
	defer conn2.Close()

	// Канал для чтения сообщений от второго пользователя
	msgCh := make(chan string, 1)
	go func() {
		_, msg, err := conn2.ReadMessage()
		if err != nil {
			log.Printf("[TestMessaging] Ошибка чтения WS: %v", err)
			return
		}
		msgCh <- string(msg)
	}()

	// Первый отправляет сообщение
	expected := "Hello from user1!"
	err = conn1.WriteMessage(websocket.TextMessage, []byte(expected))
	s.Require().NoError(err)

	// Проверка получения сообщения
	select {
	case received := <-msgCh:
		log.Printf("[TestMessaging] Второй получил: %q", received)
		s.Require().Equal(expected, received)
	case <-time.After(5 * time.Second):
		s.T().Fatal("таймаут: второй пользователь не получил сообщение")
	}
}

func TestChatSuite(t *testing.T) {
	suite.Run(t, new(ChatSuite))
}
