package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/bil-message/internal/configs/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/middlewares"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/spf13/pflag"
)

// Информация о сборке (передаётся через ldflags)
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// Конфигурационные переменные
var (
	address      string
	databaseDSN  string
	driverName   string = "sqlite"
	jwtSecretKey string
	jwtExp       time.Duration
)

func main() {
	// Печать информации о сборке
	printBuildInfo()

	// Парсинг флагов командной строки
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	// Запуск сервера
	if err := run(context.Background(), driverName, databaseDSN, address, jwtSecretKey, jwtExp); err != nil {
		log.Fatal(err)
	}
}

// printBuildInfo выводит информацию о версии, коммите и дате сборки
func printBuildInfo() {
	fmt.Println("Build Information:")
	fmt.Printf("  Version : %s\n", buildVersion)
	fmt.Printf("  Commit  : %s\n", buildCommit)
	fmt.Printf("  Date    : %s\n", buildDate)
}

// parseFlags читает параметры запуска сервера из командной строки
func parseFlags() error {
	pflag.StringVarP(&address, "address", "a", "localhost:8080", "Server address")
	pflag.StringVarP(&databaseDSN, "dsn", "d", "postgres://bil_message_user:bil_message_password@localhost:5432/bil_message_db?sslmode=disable", "Database DSN")
	pflag.StringVarP(&jwtSecretKey, "jwt-key", "k", "jwt-secret-key", "JWT secret key")
	pflag.DurationVarP(&jwtExp, "jwt-expiration", "e", 1*time.Hour, "JWT expiration time")
	pflag.Parse()
	return nil
}

// run выполняет инициализацию всех компонентов и запускает HTTP-сервер
func run(ctx context.Context, driverName, databaseDSN, address, jwtSecretKey string, jwtExp time.Duration) error {
	// Подключение к базе данных
	conn, err := db.New(driverName, databaseDSN, db.WithMaxOpenConns(10), db.WithMaxIdleConns(3))
	if err != nil {
		return err
	}

	// Репозитории для работы с данными
	userReadRepo := repositories.NewUserReadRepository(conn)
	userWriteRepo := repositories.NewUserWriteRepository(conn)
	deviceReadRepo := repositories.NewDeviceReadRepository(conn)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(conn)

	chatWriteRepo := repositories.NewChatWriteRepository(conn)
	chatMemberWriteRepo := repositories.NewChatMemberWriteRepository(conn)

	// Сервисы — бизнес-логика приложения
	deviceService := services.NewDeviceService(userReadRepo, deviceWriteRepo)

	// JWT сервис для авторизации
	jwt, err := jwt.New(
		jwt.WithSecretKey(jwtSecretKey),
		jwt.WithExpiration(jwtExp),
	)
	if err != nil {
		return err
	}

	authService := services.NewAuthService(userReadRepo, userWriteRepo, deviceReadRepo, jwt)
	chatService := services.NewChatService(chatWriteRepo, chatMemberWriteRepo)

	// Middleware авторизации
	authMiddleware := middlewares.AuthMiddleware(jwt)

	// Настройка роутера
	router := chi.NewRouter()
	router.Use(middleware.Logger)    // Логирование запросов
	router.Use(middleware.Recoverer) // Авто-восстановление после паник

	// Группировка маршрутов под /api/v1
	router.Route("/api/v1", func(r chi.Router) {
		// Роуты для авторизации
		r.Post("/auth/register", handlers.RegisterHandler(authService))
		r.Post("/auth/login", handlers.LoginHandler(authService))

		// Защищённые роуты (требуется JWT)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/auth/logout", nil) // TODO: реализовать logout
		})

		// Роуты для устройств
		r.Post("/devices/register", handlers.DeviceRegisterHandler(deviceService))

		// Роуты для чатов
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Создать чат
			r.Post("/chats", handlers.NewCreateChatHandler(jwt, chatService))

			// Добавить пользователя в чат
			r.Post("/chats/{chat-uuid}/members/{user-uuid}", handlers.NewAddMemberHandler(jwt, chatService))

			// WebSocket для чата
			r.Get("/chats/ws/{chat-uuid}", handlers.NewChatWSHandler(jwt))
		})
	})

	// HTTP сервер
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Грейсфулл завершение работы сервера
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done(): // при получении сигнала завершения
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh: // если сервер упал с ошибкой
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
