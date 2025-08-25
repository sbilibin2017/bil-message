package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/models"
)

func main() {
	// Подключаемся к NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	// Подписка на общий топик "messages"
	sub, err := nc.Subscribe("messages", func(m *nats.Msg) {
		var msg models.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Println("Invalid message:", err)
			return
		}

		// Обработка сообщения
		// Здесь можно добавлять: сохранение в БД, push-уведомления, фильтры/модерацию
		log.Printf("Room: %s | User: %s | Device: %s | Text: %s\n",
			msg.RoomUUID, msg.UserUUID, msg.DeviceUUID, msg.Text)
	})
	if err != nil {
		log.Fatal("Failed to subscribe to NATS topic:", err)
	}
	defer sub.Unsubscribe()

	log.Println("Message Worker listening to NATS topic 'messages'...")

	// Держим процесс живым
	select {}
}
