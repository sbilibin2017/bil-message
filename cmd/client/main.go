package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sbilibin2017/bil-message/internal/client"
	"github.com/sbilibin2017/bil-message/internal/models"
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
	serverURL  string
	username   string
	password   string
	publicKey  string
	token      string
	roomUUID   string
	memberUUID string
)

const deviceUUIDFile = ".config/device_uuid"

func parseFlags() {
	flag.StringVar(&serverURL, "url", "http://localhost:8080/api/v1", "Base URL of the auth server")
	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&publicKey, "public-key", "", "Public key for device")
	flag.StringVar(&token, "token", "", "JWT token for auth")
	flag.StringVar(&roomUUID, "room-uuid", "", "Room UUID")
	flag.StringVar(&memberUUID, "member-uuid", "", "Member UUID")
}

func main() {
	parseFlags()

	if len(os.Args) < 2 {
		log.Fatalln("Commands: register, add-device, login, create-room, delete-room, add-room-member, remove-room-member")
	}

	command := os.Args[1]
	ctx := context.Background()
	authClient := client.NewAuthClient(serverURL)
	restyClient := resty.New()
	roomClient := client.NewRoomClient(restyClient)

	switch command {

	case "version":
		printBuildInfo()

	case "register":
		userUUID, err := authClient.Register(ctx, username, password)
		if err != nil {
			log.Fatalf("Register failed: %v", err)
		}
		log.Println("User UUID:", userUUID)

	case "add-device":
		deviceUUID, err := authClient.AddDevice(ctx, username, password, publicKey)
		if err != nil {
			log.Fatalf("Add device failed: %v", err)
		}
		if err := saveDeviceUUID(deviceUUID); err != nil {
			log.Fatalf("Failed to save device UUID: %v", err)
		}
		log.Println("Device UUID:", deviceUUID)

	case "login":
		deviceUUID, err := loadDeviceUUID()
		if err != nil {
			log.Fatalf("Failed to load device UUID: %v", err)
		}
		token, err := authClient.Login(ctx, username, password, deviceUUID)
		if err != nil {
			log.Fatalf("Login failed: %v", err)
		}
		log.Println("JWT token:", token)

	case "create-room":
		roomID, err := roomClient.CreateRoom(ctx, token)
		if err != nil {
			log.Fatalf("Create room failed: %v", err)
		}
		log.Println(roomID)

	case "delete-room":
		rUUID, err := uuid.Parse(roomUUID)
		if err != nil {
			log.Fatalf("Invalid room UUID: %v", err)
		}
		if err := roomClient.DeleteRoom(ctx, token, rUUID); err != nil {
			log.Fatalf("Delete room failed: %v", err)
		}

	case "add-room-member":
		rUUID, _ := uuid.Parse(roomUUID)
		mUUID, _ := uuid.Parse(memberUUID)
		if err := roomClient.AddMember(ctx, token, rUUID, mUUID); err != nil {
			log.Fatalf("Add member failed: %v", err)
		}

	case "remove-room-member":
		rUUID, _ := uuid.Parse(roomUUID)
		mUUID, _ := uuid.Parse(memberUUID)
		if err := roomClient.RemoveMember(ctx, token, rUUID, mUUID); err != nil {
			log.Fatalf("Remove member failed: %v", err)
		}

	case "room-connect":
		rUUID, err := uuid.Parse(roomUUID)
		if err != nil {
			log.Fatalf("Invalid room UUID: %v", err)
		}

		// Формируем URL для WebSocket
		wsURL := fmt.Sprintf("ws://%s/room/%s/ws", serverURL[len("http://"):], rUUID.String())

		header := http.Header{}
		header.Set("Authorization", "Bearer "+token)

		dialer := websocket.DefaultDialer
		conn, _, err := dialer.Dial(wsURL, header)
		if err != nil {
			log.Fatalf("Failed to connect to room WebSocket: %v", err)
		}
		defer conn.Close()
		log.Println("Connected to room WebSocket. Type messages and press Enter to send.")

		// Чтение входящих сообщений в отдельной горутине
		go func() {
			for {
				var msg models.RoomMessage
				if err := conn.ReadJSON(&msg); err != nil {
					log.Println("WebSocket read error:", err)
					return
				}
				log.Printf("[%s] %s\n", msg.UserUUID, msg.Message)
			}
		}()

		// Чтение сообщений с stdin и отправка plain text
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
				log.Println("Failed to send message:", err)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Println("Error reading input:", err)
		}

	default:
		log.Fatalf("Unknown command: %s\n", command)
	}
}

func saveDeviceUUID(id uuid.UUID) error {
	configDir, _ := os.UserHomeDir()
	path := filepath.Join(configDir, deviceUUIDFile)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	return os.WriteFile(path, []byte(id.String()), 0o600)
}

func loadDeviceUUID() (uuid.UUID, error) {
	configDir, _ := os.UserHomeDir()
	path := filepath.Join(configDir, deviceUUIDFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(string(data))
}
