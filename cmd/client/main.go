package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/client"
	"github.com/sbilibin2017/bil-message/internal/transport/http"
	"github.com/spf13/cobra"
)

// Флаги сборки (ldflags)
var (
	buildCommit  string = "N/A" // Хэш коммита сборки
	buildDate    string = "N/A" // Дата сборки
	buildVersion string = "N/A" // Версия сборки
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run создаёт root команду и добавляет подкоманды, затем выполняет её
func run() error {
	cmd := newRootCommand()
	cmd.AddCommand(
		newRegisterCommand(),
		newDeviceCommand(),
		newLoginCommand(),
		newVersionCommand(),
	)
	return cmd.Execute()
}

// newRootCommand создаёт корневую команду CLI
func newRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bil-message-client",
		Short: "CLI клиент для bil-message",
		Long:  "CLI клиент для взаимодействия с сервером bil-message: регистрация, управление устройствами и вход в аккаунт.",
	}
}

// newRegisterCommand создаёт команду 'register' для регистрации нового пользователя
func newRegisterCommand() *cobra.Command {
	var address, username, password string

	cmd := &cobra.Command{
		Use:     "register",
		Short:   "Регистрация нового пользователя",
		Example: "bil-message-client register --username testuser --password secret --address http://localhost:8080",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			httpClient, err := http.New(address, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 3 * time.Second,
			}))
			if err != nil {
				return err
			}

			if err := client.Register(ctx, httpClient, username, password); err != nil {
				return fmt.Errorf("не удалось выполнить регистрацию: %w", err)
			}

			log.Println("Регистрация прошла успешно")
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "http://localhost:8080", "Адрес сервера")
	cmd.Flags().StringVarP(&username, "username", "u", "user", "Имя пользователя для регистрации")
	cmd.Flags().StringVarP(&password, "password", "p", "password", "Пароль пользователя для регистрации")
	return cmd
}

// newDeviceCommand создаёт команду 'device' для добавления нового устройства
func newDeviceCommand() *cobra.Command {
	var address, username, password, publicKey string

	cmd := &cobra.Command{
		Use:     "device",
		Short:   "Добавление нового устройства для пользователя",
		Example: "bil-message-client device --username testuser --password secret --public-key key123 --address http://localhost:8080",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			httpClient, err := http.New(address, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 3 * time.Second,
			}))
			if err != nil {
				return err
			}

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

			cmd.Println("UUID устройства сохранён в файле:", deviceFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "http://localhost:8080", "Адрес сервера")
	cmd.Flags().StringVarP(&username, "username", "u", "user", "Имя пользователя")
	cmd.Flags().StringVarP(&password, "password", "p", "password", "Пароль пользователя")
	cmd.Flags().StringVarP(&publicKey, "public-key", "k", "key", "Публичный ключ устройства")
	return cmd
}

// newLoginCommand создаёт команду 'login' для входа пользователя
func newLoginCommand() *cobra.Command {
	var address, username, password string

	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Вход пользователя",
		Example: "bil-message-client login --username testuser --password secret --address http://localhost:8080",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			httpClient, err := http.New(address, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 3 * time.Second,
			}))
			if err != nil {
				return err
			}

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

			cmd.Println(token)
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "http://localhost:8080", "Адрес сервера")
	cmd.Flags().StringVarP(&username, "username", "u", "user", "Имя пользователя")
	cmd.Flags().StringVarP(&password, "password", "p", "password", "Пароль пользователя")
	return cmd
}

// newVersionCommand создаёт команду 'version' для вывода информации о версии клиента
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Показать информацию о версии клиента",
		Example: "bil-message-client version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("bil-message-client")
			cmd.Printf("Версия: %s\n", buildVersion)
			cmd.Printf("Коммит: %s\n", buildCommit)
			cmd.Printf("Дата сборки: %s\n", buildDate)
		},
	}
}
