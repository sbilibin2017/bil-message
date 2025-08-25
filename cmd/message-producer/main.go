package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/facades"
	"github.com/sbilibin2017/bil-message/internal/handlers"
)

func main() {
	authService := "http://localhost:9000"

	// Создаём экземпляр AuthFacade
	authFacade := facades.NewAuthFacade(authService)

	// Подключение к NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	r := chi.NewRouter()
	r.Get("/ws/{room-uuid}", handlers.ChatHandler(authFacade, nc))

	log.Println("Message Producer WS server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
