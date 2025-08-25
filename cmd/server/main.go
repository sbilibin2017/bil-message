// main.go
// @title       bil-message API
// @version     1.0
// @description API для защищённого обмена сообщениями между пользователями
// @host        localhost:8080
// @BasePath    /api/v1
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
	"github.com/sbilibin2017/bil-message/internal/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/spf13/pflag"
)

func main() {
	printBuildInfo()
	parseFlags()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// Флаги сборки (ldflags)
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
	address      string
	databaseDSN  string
	jwtSecretKey string
	jwtExp       int
)

// parseFlags парсит флаги командной строки
func parseFlags() {
	pflag.StringVarP(&address, "address", "a", ":8080", "Адрес и порт для запуска сервера")
	pflag.StringVarP(&databaseDSN, "database-dsn", "d", "postgres://user:pass@localhost:5432/db?sslmode=disable", "DSN для подключения к базе данных")
	pflag.StringVarP(&jwtSecretKey, "jwt-secret", "", "super-secret-key", "Секретный ключ для генерации JWT")
	pflag.IntVarP(&jwtExp, "jwt-expiration", "", 86400, "Время жизни JWT токена в секундах")
	pflag.Parse()
}

// run выполняет запуск сервера
func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	db, err := db.New(
		"pgx",
		databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(3),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	userReadRepo := repositories.NewUserReadRepository(db)
	userWriteRepo := repositories.NewUserWriteRepository(db)

	deviceReadRepo := repositories.NewDeviceReadRepository(db)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(db)

	roomReadRepo := repositories.NewRoomReadRepository(db)
	roomWriteRepo := repositories.NewRoomWriteRepository(db)

	roomMemberReadRepo := repositories.NewRoomMemberReadRepository(db)
	roomMemberWriteRepo := repositories.NewRoomMemberWriteRepository(db)

	jwt, err := jwt.New(
		jwt.WithSecretKey(jwtSecretKey),
		jwt.WithExpiration(time.Duration(jwtExp)*time.Second),
	)
	if err != nil {
		return err
	}

	authService := services.NewAuthService(
		userReadRepo,
		userWriteRepo,
		deviceReadRepo,
		deviceWriteRepo,
		jwt,
	)

	chatService := services.NewChatService(
		roomWriteRepo,
		roomReadRepo,
		roomMemberWriteRepo,
		roomMemberReadRepo,
	)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.RegisterHandler(authService))
			r.Post("/device", handlers.AddDeviceHandler(authService))
			r.Post("/login", handlers.LoginHandler(authService))
		})

		r.Route("/chat", func(r chi.Router) {
			r.Post("/", handlers.CreateChatHandler(chatService, jwt))
			r.Delete("/{chat-uuid}", handlers.RemoveChatHandler(chatService, jwt))
			r.Post("/{chat-uuid}/{member-uuid}", handlers.AddChatMemberHandler(chatService, jwt))
			r.Delete("/{chat-uuid}/{member-uuid}", handlers.RemoveChatMemberHandler(chatService, jwt))
		})
	})

	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

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
