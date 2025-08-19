package main

import (
	"context"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/middlewares"
)

// Переменные для конфигурации сервера и базы данных
var (
	address      string // Адрес и порт сервера
	databaseDSN  string // DSN для подключения к базе данных
	jwtSecretKey string // Секретный ключ для подписи JWT
	jwtExp       string // Время жизни JWT в секундах
)

// init инициализирует флаги командной строки
func init() {
	flag.StringVar(&address, "a", "localhost:8080", "Server address")
	flag.StringVar(&databaseDSN, "d", "postgres://bil_message_user:bil_message_password@localhost:5432/bil_message_db?sslmode=disable", "Database DSN")
	flag.StringVar(&jwtSecretKey, "k", "jwt-secret-key", "JWT secret key")
	flag.StringVar(&jwtExp, "e", "3600", "JWT expiration time in seconds")
}

// main парсит флаги и запускает сервер
func main() {
	flag.Parse()
	ctx := context.Background()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

// run запускает HTTP-сервер с маршрутизацией и middleware
func run(ctx context.Context) error {
	router := chi.NewRouter() // Создаём роутер chi

	// Инициализируем JWT обработчик с секретным ключом
	jwt := jwt.New(jwtSecretKey)

	// Создаём middleware аутентификации на основе JWT
	authMiddleware := middlewares.AuthMiddleware(jwt)

	// Группируем маршруты API
	router.Route("/api/v1", func(r chi.Router) {
		// Публичные маршруты
		r.Post("/register", nil) // Регистрирует нового пользователя
		r.Post("/login", nil)    // Аутентификация пользователя

		// Защищённые маршруты (требуют JWT)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)  // Применяем middleware аутентификации
			r.Post("/logout", nil) // Выход пользователя
		})
	})

	// Настраиваем HTTP-сервер
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Создаём контекст, который отменяется при получении сигнала завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop() // Отменяем контекст при завершении функции

	// Запуск сервера в отдельной горутине
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Ждём либо сигнала завершения, либо ошибки сервера
	select {
	case <-ctx.Done(): // Получен сигнал завершения
		// Выполняем graceful shutdown с таймаутом 5 секунд
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh: // Сервер завершился с ошибкой
		if err == http.ErrServerClosed {
			return nil // Сервер корректно остановлен
		}
		return err // Возвращаем ошибку
	}
}
