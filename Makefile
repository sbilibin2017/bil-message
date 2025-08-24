# Генерация моков для интерфейсов Go
gen-mock:
	# Используется mockgen для создания mock-реализаций интерфейсов
	mockgen -source=$(file) \
		-destination=$(dir $(file))/$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

# Запуск всех тестов и генерация покрытия кода
test:
	# Тесты для всех пакетов с включением отчета покрытия
	go test ./... -cover

# Сборка клиентских бинарников для различных платформ
build-clients:
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -o builds/clients/bil-message-client-linux-amd64 ./cmd/client
	# MacOS amd64
	GOOS=darwin GOARCH=amd64 go build -o builds/clients/bil-message-client-macos-amd64 ./cmd/client
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -o builds/clients/bil-message-client-windows-amd64.exe ./cmd/client

# Сборка серверного бинарника для Linux
build-server:
	GOOS=linux GOARCH=amd64 go build -o builds/server/bil-message-server-linux-amd64 ./cmd/server	

# Генерация swagger-документации из хэндлеров
gen-swag:
	# Используется swag для анализа internal/handlers и генерации документации в api/http
	swag init -d internal/handlers -g ../../cmd/server/main.go -o api/http

# Применение миграций к базе данных PostgreSQL
migrate:
	# Используется goose для выполнения всех миграций в директории ./migrations
	goose -dir ./migrations postgres "host=localhost port=5432 user=bil_message_user password=bil_message_password dbname=bil_message_db sslmode=disable" up
