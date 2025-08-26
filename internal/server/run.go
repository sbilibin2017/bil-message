package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/bil-message/internal/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// Run запускает HTTP сервер с эндпоинтами авторизации
func Run(
	ctx context.Context,
	addr string,
	version string,
	databaseDriver string,
	databaseDSN string,
	jwtSecretKey string,
	jwtExp time.Duration,
) error {
	// Подключение к базе данных
	dbConn, err := db.New(
		databaseDriver,
		databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(5),
	)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Репозитории
	userWriteRepo := repositories.NewUserWriteRepository(dbConn)
	userReadRepo := repositories.NewUserReadRepository(dbConn)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(dbConn)
	deviceReadRepo := repositories.NewDeviceReadRepository(dbConn)

	// JWT генератор
	tokenGen, err := jwt.New(jwt.WithSecretKey(jwtSecretKey), jwt.WithExpiration(jwtExp))
	if err != nil {
		return err
	}

	// Сервис авторизации
	authSvc := services.NewAuthService(
		userWriteRepo,
		userReadRepo,
		deviceWriteRepo,
		deviceReadRepo,
		tokenGen,
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
