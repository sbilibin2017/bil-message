package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sbilibin2017/bil-message/internal/client"
	"github.com/sbilibin2017/bil-message/internal/configs/transport/http"
	"github.com/spf13/pflag"
)

// main — точка входа в CLI клиент
func main() {
	printBuildInfo()
	parseFlags()
	if err := run(context.Background()); err != nil {
		printHelp()
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
	address  string // Адрес сервера для подключения
	username string // Имя пользователя для регистрации
	password string // Пароль пользователя для регистрации
)

// parseFlags парсит флаги командной строки
func parseFlags() {
	pflag.StringVarP(&address, "address", "a", "http://localhost:8080", "Адрес сервера")
	pflag.StringVarP(&username, "username", "u", "user", "Имя пользователя для регистрации")
	pflag.StringVarP(&password, "password", "p", "password", "Пароль пользователя для регистрации")
	pflag.Parse()
}

// printHelp выводит информацию о доступных командах и флагах
func printHelp() {
	fmt.Println("Использование:")
	fmt.Println("  bil-message-client <команда> [флаги]")
	fmt.Println()
	fmt.Println("Доступные команды:")
	fmt.Println("  register    Регистрация нового пользователя")
	fmt.Println()
	fmt.Println("Флаги:")
	fmt.Println("  -a, --address       Адрес сервера")
	fmt.Println("  -u, --username      Имя пользователя для регистрации")
	fmt.Println("  -p, --password      Пароль пользователя для регистрации")
}

// run выполняет команду CLI
// Поддерживает команды:
//   - register: регистрация нового пользователя на сервере
func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("command is not provided")
	}

	command := os.Args[1]

	switch command {
	case "register":
		// Создание HTTP клиента с политикой повторных попыток
		restClient, err := http.New(address, http.WithRetryPolicy(
			http.RetryPolicy{
				Count:   3,               // Количество повторов
				Wait:    1 * time.Second, // Время ожидания между попытками
				MaxWait: 3 * time.Second, // Максимальное время ожидания
			},
		))
		if err != nil {
			return err
		}

		// Вызов регистрации пользователя на сервере
		userUUID, err := client.Register(ctx, restClient, username, password)
		if err != nil {
			return err
		}

		log.Println(*userUUID) // Вывод UUID зарегистрированного пользователя
		return nil

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
