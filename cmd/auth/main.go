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

// main – точка входа в приложение.
// Здесь выводится информация о сборке, парсятся флаги командной строки и запускается HTTP-сервер.
func main() {
	printBuildInfo() // Вывод информации о версии сборки

	err := parseFlags() // Парсинг флагов командной строки
	if err != nil {
		log.Fatal(err) // Завершение программы при ошибке парсинга
	}

	// Запуск основной логики сервера
	if err := run(
		context.Background(),
		driverName,
		databaseDSN,
		address,
		jwtSecretKey,
		jwtExp,
	); err != nil {
		log.Fatal(err) // Завершение программы при ошибке запуска сервера
	}
}

// Флаги сборки (передаются через ldflags при сборке)
var (
	buildVersion = "N/A" // Версия сборки
	buildDate    = "N/A" // Дата сборки
	buildCommit  = "N/A" // Хэш коммита
)

// Конфигурационные переменные по умолчанию
var (
	address      string                // Адрес сервера
	databaseDSN  string                // DSN базы данных
	jwtSecretKey string                // Секретный ключ для JWT
	jwtExp       time.Duration         // Время жизни JWT
	driverName   string        = "pgx" // Драйвер базы данных (Postgres)
)

// printBuildInfo выводит информацию о сборке приложения.
// Полезно для отладки и проверки версии.
func printBuildInfo() {
	fmt.Println("Build Information:")
	fmt.Printf("  Version : %s\n", buildVersion)
	fmt.Printf("  Commit  : %s\n", buildCommit)
	fmt.Printf("  Date    : %s\n", buildDate)
}

// parseFlags парсит флаги командной строки.
// Используется библиотека spf13/pflag для удобного определения и чтения флагов.
func parseFlags() error {
	pflag.StringVarP(&address, "address", "a", "localhost:8080", "Server address")                                                                               // Адрес сервера
	pflag.StringVarP(&databaseDSN, "dsn", "d", "postgres://bil_message_user:bil_message_password@localhost:5432/bil_message_db?sslmode=disable", "Database DSN") // DSN БД
	pflag.StringVarP(&jwtSecretKey, "jwt-key", "k", "jwt-secret-key", "JWT secret key")                                                                          // JWT ключ
	pflag.DurationVarP(&jwtExp, "jwt-expiration", "e", 1*time.Hour, "JWT expiration time")                                                                       // Время жизни JWT

	pflag.Parse() // Применение и чтение флагов
	return nil
}

// run инициализирует соединение с базой данных, репозитории, сервисы, маршруты и запускает HTTP-сервер.
func run(
	ctx context.Context,
	driverName,
	databaseDSN,
	address,
	jwtSecretKey string,
	jwtExp time.Duration,
) error {
	// Создание подключения к базе данных с конфигурацией пула соединений
	conn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(3),
	)
	if err != nil {
		return err // Возврат ошибки, если не удалось подключиться
	}

	// Инициализация репозиториев для работы с пользователями
	userReadRepo := repositories.NewUserReadRepository(conn)
	userWriteRepo := repositories.NewUserWriteRepository(conn)

	// Инициализация JWT-сервиса
	jwt, err := jwt.New(jwtSecretKey, jwtExp)
	if err != nil {
		return err
	}

	// Создание сервиса аутентификации
	authService := services.NewAuthService(userReadRepo, userWriteRepo, jwt)

	// Инициализация HTTP-хендлеров
	authHandler := handlers.RegisterHandler(authService)

	// Создание middleware для проверки JWT
	authMiddleware := middlewares.AuthMiddleware(jwt)

	// Создание маршрутизатора chi
	router := chi.NewRouter()

	// Создание middleware для логирования запросов
	router.Use(middleware.Logger)

	// Создание middleware для восстановления после паники
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		// Публичные маршруты
		r.Post("/auth/register", authHandler) // Регистрация нового пользователя
		r.Post("/auth/login", nil)

		// Защищенные маршруты (требуют JWT)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/auth/logout", nil)
		})
	})

	// Настройка HTTP-сервера
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Обработка сигналов завершения (CTRL+C, SIGTERM)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Канал для ошибок сервера
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }() // Запуск сервера в отдельной горутине

	// Ожидание завершения через сигнал или ошибку сервера
	select {
	case <-ctx.Done():
		// Грейсфулл-выход с таймаутом 5 секунд
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil // Сервер корректно завершил работу
		}
		return err // Возврат ошибки сервера
	}
}
