package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/client"
	"github.com/sbilibin2017/bil-message/internal/client/http"
	"github.com/spf13/pflag"
)

// main — точка входа в CLI клиент
func main() {
	printBuildInfo()
	parseFlags()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// Флаги сборки (ldflags)
var (
	buildCommit  string = "N/A" // Хэш коммита сборки
	buildDate    string = "N/A" // Дата сборки
	buildVersion string = "N/A" // Версия сборки
)

// printBuildInfo выводит информацию о версии, коммите и дате сборки
func printBuildInfo() {
	log.Printf("Client build info:\n")
	log.Printf("  Version: %s\n", buildVersion)
	log.Printf("  Commit:  %s\n", buildCommit)
	log.Printf("  Date:    %s\n", buildDate)
}

// Флаги командной строки
var (
	address   string
	username  string
	password  string
	publicKey string
)

// parseFlags парсит флаги командной строки
func parseFlags() {
	pflag.StringVarP(&address, "address", "a", "http://localhost:8080", "Адрес сервера")
	pflag.StringVarP(&username, "username", "u", "user", "Имя пользователя для регистрации")
	pflag.StringVarP(&password, "password", "p", "password", "Пароль пользователя для регистрации")
	pflag.StringVarP(&publicKey, "public-key", "k", "key", "Публичный ключ пользователя для устройства")
	pflag.Parse()
}

// run выполняет команду CLI
// Поддерживает команды:
//   - register: регистрация нового пользователя на сервере
func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("command is not provided")
	}

	command := os.Args[1]

	httpClient, err := http.New(address, http.WithRetryPolicy(
		http.RetryPolicy{
			Count:   3,
			Wait:    1 * time.Second,
			MaxWait: 3 * time.Second,
		},
	))
	if err != nil {
		return err
	}

	switch command {
	case "register":
		err := client.Register(ctx, httpClient, username, password)
		if err != nil {
			return fmt.Errorf("не удалось выполнить регистрацию: %w", err)
		}
		return nil

	case "device":
		deviceUUID, err := client.AddDevice(ctx, httpClient, username, password, publicKey)
		if err != nil {
			return fmt.Errorf("не удалось добавить устройство: %w", err)
		}

		configDir := os.ExpandEnv("$HOME/.config")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("не удалось создать директорию конфигурации: %w", err)
		}

		deviceFile := fmt.Sprintf("%s/bil_message_client_device_uuid", configDir)
		if err := os.WriteFile(deviceFile, []byte(deviceUUID.String()), 0o600); err != nil {
			return fmt.Errorf("не удалось записать uuid устройства в файл: %w", err)
		}

		log.Println("uuid устройства сохранён в файле:", deviceFile)
		return nil

	case "login":
		configDir := os.ExpandEnv("$HOME/.config")
		deviceFile := fmt.Sprintf("%s/bil_message_client_device_uuid", configDir)

		data, err := os.ReadFile(deviceFile)
		if err != nil {
			return fmt.Errorf("не удалось прочитать UUID устройства из файла: %w", err)
		}

		deviceUUID, err := uuid.Parse(string(data))
		if err != nil {
			return fmt.Errorf("некорректный uuid устройства в файле: %w", err)
		}

		token, err := client.Login(ctx, httpClient, username, password, deviceUUID)
		if err != nil {
			return fmt.Errorf("не удалось выполнить вход: %w", err)
		}

		log.Println(token)
		return nil

	default:
		return fmt.Errorf("неизвестная команда: %s", command)
	}
}
