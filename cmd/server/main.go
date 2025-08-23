package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sbilibin2017/bil-message/internal/configs/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/spf13/pflag"
)

// main — точка входа в сервер
func main() {
	printBuildInfo()
	parseFlags()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// Флаги сборки(ldflags)
var (
	buildCommit  string = "N/A"
	buildDate    string = "N/A"
	buildVersion string = "N/A"
)

// printBuildInfo выводит информацию о версии, коммите и дате сборки
func printBuildInfo() {
	log.Printf("Server build info:\n")
	log.Printf("  Version: %s\n", buildVersion)
	log.Printf("  Commit:  %s\n", buildCommit)
	log.Printf("  Date:    %s\n", buildDate)
}

// Флаги командной строки
var (
	address     string
	databaseDSN string
)

// parseFlags парсит флаги мкомандной строки
func parseFlags() {
	pflag.StringVarP(&address, "address", "a", ":8080", "Адрес и порт для запуска сервера")
	pflag.StringVarP(&databaseDSN, "database-dsn", "d", "postgres://user:pass@localhost:5432/db?sslmode=disable", "DSN для подключения к базе данных")
	pflag.Parse()
}

// run запускает HTTP-сервер с поддержкой graceful shutdown
func run(ctx context.Context) error {
	// Контекст с отменой по сигналу прерывания
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Подключение к базе данных
	dbConn, err := db.New(
		"pgx",
		databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(3),
	)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Репозитории
	userReadRepo := repositories.NewUserReadRepository(dbConn)
	userWriteRepo := repositories.NewUserWriteRepository(dbConn)

	// Сервис аутентификации
	authService := services.NewAuthService(userReadRepo, userWriteRepo)

	// Настройка Chi роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.RegisterHandler(authService))
		})
	})

	// HTTP-сервер
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	// Канал для ошибок сервера
	errChan := make(chan error, 1)

	// Запуск сервера в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// Ожидание сигнала или ошибки сервера
	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctxShutdown); err != nil {
			return err
		}
		return nil

	case err := <-errChan:
		return err
	}
}
