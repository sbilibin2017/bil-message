package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/db"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// Build info, set via -ldflags
var (
	buildVersion = "N/A"
	buildCommit  = "N/A"
	buildDate    = "N/A"
)

func printBuildInfo() {
	fmt.Printf("Build version: %s\nCommit: %s\nDate: %s\n", buildVersion, buildCommit, buildDate)
}

var (
	addr           string
	version        string = "/api/v1"
	databaseDriver string = "pgx"
	databaseDSN    string
	jwtSecretKey   string
	jwtExp         time.Duration
	natsURL        string
)

func parseFlags() {
	flag.StringVar(&addr, "a", ":8080", "HTTP server address")
	flag.StringVar(&databaseDSN, "d", "postgres://user:password@localhost:5432/db?sslmode=disable", "Database DSN (connection string)")
	flag.StringVar(&jwtSecretKey, "k", "secret-key", "JWT secret key")
	flag.DurationVar(&jwtExp, "e", time.Duration(1)*time.Hour, "JWT expiration duration in hours")
	flag.StringVar(&natsURL, "n", nats.DefaultURL, "NATS server URL")
	flag.Parse()
}

func main() {
	printBuildInfo()
	parseFlags()

	err := run(
		context.Background(),
		addr,
		version,
		databaseDriver,
		databaseDSN,
		jwtSecretKey,
		jwtExp,
		natsURL,
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
	jwtExp time.Duration,
	natsURL string,
) error {
	// Подключение к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	defer nc.Close()

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

	roomWriteRepo := repositories.NewRoomWriteRepository(db)
	roomReadRepo := repositories.NewRoomReadRepository(db)

	roomMemberWriteRepo := repositories.NewRoomMemberWriteRepository(db)
	roomMemberReadRepo := repositories.NewRoomMemberReadRepository(db)

	// JWT генератор
	jwt, err := jwt.New(
		jwt.WithSecretKey(jwtSecretKey),
		jwt.WithExpiration(jwtExp),
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

	// Сервис комнат
	roomSvc := services.NewRoomService(
		roomWriteRepo,
		roomReadRepo,
		roomMemberWriteRepo,
		roomMemberReadRepo,
		userReadRepo,
	)

	// Инициализация обработчиков
	registerHandler := handlers.NewRegisterHandler(authSvc)
	deviceAddHandler := handlers.NewDeviceAddHandler(authSvc)
	loginHandler := handlers.NewLoginHandler(authSvc)

	roomCreateHandler := handlers.NewRoomCreateHandler(roomSvc, jwt)
	roomDeleteHandler := handlers.NewRoomDeleteHandler(roomSvc, jwt)
	roomMemberAddHandler := handlers.NewRoomMemberAddHandler(roomSvc, jwt)
	roomMemberRemoveHandler := handlers.NewRoomMemberRemoveHandler(roomSvc, jwt)
	roomWSHandler := handlers.NewRoomWebsocketHandler(jwt, nc)

	// Chi роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// Группа роутов: /{version}
	r.Route(version, func(r chi.Router) {
		// Auth
		r.Post("/auth/register", registerHandler)
		r.Post("/auth/device/add", deviceAddHandler)
		r.Post("/auth/login", loginHandler)

		// Rooms
		r.Post("/rooms", roomCreateHandler)
		r.Delete("/rooms/{room-uuid}", roomDeleteHandler)
		r.Post("/rooms/{room-uuid}/{member-uuid}", roomMemberAddHandler)
		r.Post("/rooms/{room-uuid}/{member-uuid}", roomMemberRemoveHandler)
		r.Get("/room/{room-uuid}/ws", roomWSHandler)
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
