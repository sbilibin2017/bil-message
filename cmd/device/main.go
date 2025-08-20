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
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/spf13/pflag"
)

// Конфигурация сборки
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// Конфигурационные переменные по умолчанию
var (
	address     string
	databaseDSN string
	driverName  string = "pgx"
)

func main() {
	printBuildInfo()

	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	if err := run(context.Background(), driverName, databaseDSN, address); err != nil {
		log.Fatal(err)
	}
}

func printBuildInfo() {
	fmt.Println("Build Information:")
	fmt.Printf("  Version : %s\n", buildVersion)
	fmt.Printf("  Commit  : %s\n", buildCommit)
	fmt.Printf("  Date    : %s\n", buildDate)
}

func parseFlags() error {
	pflag.StringVarP(&address, "address", "a", "localhost:8081", "Server address")
	pflag.StringVarP(&databaseDSN, "dsn", "d", "postgres://bil_message_user:bil_message_password@localhost:5432/bil_message_db?sslmode=disable", "Database DSN")
	pflag.Parse()
	return nil
}

func run(ctx context.Context, driverName, databaseDSN, address string) error {
	// Подключение к базе данных
	conn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(3),
	)
	if err != nil {
		return err
	}

	// Репозитории устройств
	deviceWriteRepo := repositories.NewDeviceWriteRepository(conn)

	// Репозиторий пользователей для проверки существования
	userReadRepo := repositories.NewUserReadRepository(conn)

	// Сервис устройств
	deviceService := services.NewDeviceService(userReadRepo, deviceWriteRepo)

	// Хендлеры устройств
	deviceHandler := handlers.DeviceRegisterHandler(deviceService)

	// Маршрутизатор chi
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/devices/register", deviceHandler)
	})

	// Настройка HTTP-сервера
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Грейсфулл-шатдаун
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
