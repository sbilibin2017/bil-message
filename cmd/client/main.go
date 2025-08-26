package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/client"
)

var (
	serverURL string
	username  string
	password  string
	publicKey string
)

const deviceUUIDFile = ".config/device_uuid"

func init() {
	flag.StringVar(&serverURL, "url", "http://localhost:8080/api/v1", "Base URL of the auth server")
	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&publicKey, "public-key", "", "Public key for device")
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalln("Commands: register, add-device, login")
	}

	command := os.Args[1]

	authClient := client.NewAuthClient(serverURL)
	ctx := context.Background()

	switch command {
	case "register":
		userUUID, err := authClient.Register(ctx, username, password)
		if err != nil {
			log.Fatalf("Register failed: %v", err)
		}
		log.Println(userUUID)

	case "add-device":
		deviceUUID, err := authClient.AddDevice(ctx, username, password, publicKey)
		if err != nil {
			log.Fatalf("Add device failed: %v", err)
		}
		if err := saveDeviceUUID(deviceUUID); err != nil {
			log.Fatalf("Failed to save device UUID: %v", err)
		}
		log.Println(deviceUUID)

	case "login":
		deviceUUID, err := loadDeviceUUID()
		if err != nil {
			log.Fatalf("Failed to load device UUID: %v", err)
		}
		token, err := authClient.Login(ctx, username, password, deviceUUID)
		if err != nil {
			log.Fatalf("Login failed: %v", err)
		}
		log.Println(token)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func saveDeviceUUID(id uuid.UUID) error {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, deviceUUIDFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(id.String()), 0o600)
}

func loadDeviceUUID() (uuid.UUID, error) {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return uuid.Nil, err
	}
	path := filepath.Join(configDir, deviceUUIDFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(string(data))
}
