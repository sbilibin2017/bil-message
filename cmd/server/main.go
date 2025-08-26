package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/bil-message/internal/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
)

var (
	addr           string
	version        string = "/api/v1"
	databaseDriver string = "pgx"
	databaseDSN    string
	jwtSecretKey   string
	jwtExp         int
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "HTTP server address")
	flag.StringVar(&databaseDSN, "database-dsn", "", "Database DSN (connection string)")
	flag.StringVar(&jwtSecretKey, "jwt-secret", "secret-key", "JWT secret key")
	flag.IntVar(&jwtExp, "jwt-exp", 1, "JWT expiration duration in hours")
}

func main() {
	flag.Parse()

	err := run(
		context.Background(),
		addr,
		version,
		databaseDriver,
		databaseDSN,
		jwtSecretKey,
		jwtExp,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// run запускает HTTP сервер с эндпоинтами авторизации
func run(
	ctx context.Context,
	addr string,
	version string,
	databaseDriver string,
	databaseDSN string,
	jwtSecretKey string,
	jwtExp int,
) error {
	// Подключение к базе данных
	db, err := db.New(
		databaseDriver,
		databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(5),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	// Репозитории
	userWriteRepo := repositories.NewUserWriteRepository(db)
	userReadRepo := repositories.NewUserReadRepository(db)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(db)
	deviceReadRepo := repositories.NewDeviceReadRepository(db)

	// JWT генератор
	jwt, err := jwt.New(
		jwt.WithSecretKey(jwtSecretKey),
		jwt.WithExpiration(time.Duration(jwtExp)*time.Hour),
	)
	if err != nil {
		return err
	}

	// Сервис авторизации
	authSvc := services.NewAuthService(
		userWriteRepo,
		userReadRepo,
		deviceWriteRepo,
		deviceReadRepo,
		jwt,
	)

	// Chi роутер
	r := chi.NewRouter()

	// Группа роутов: /{version}/auth
	r.Route(version+"/auth", func(r chi.Router) {
		r.Post("/register", handlers.NewRegisterHandler(authSvc))
		r.Post("/device", handlers.NewAddDeviceHandler(authSvc))
		r.Post("/login", handlers.NewLoginHandler(authSvc))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Канал для ошибок сервера
	errCh := make(chan error, 1)

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// Проверяем возможную ошибку shutdown
	if shutdownErr, ok := <-errCh; ok && shutdownErr != nil {
		return shutdownErr
	}

	return nil
}
