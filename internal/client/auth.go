package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// Register отправляет запрос на регистрацию пользователя и возвращает user_uuid.
func Register(
	ctx context.Context,
	client *resty.Client,
	username string,
	password string,
) error {
	// Подготавливаем JSON тело запроса
	body := map[string]string{
		"username": username,
		"password": password,
	}

	// Отправка POST запроса на сервер
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/auth/register")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// AddDevice отправляет запрос на регистрацию нового устройства для пользователя и возвращает UUID устройства.
func AddDevice(
	ctx context.Context,
	client *resty.Client,
	username, password, publicKey string,
) (deviceUUID uuid.UUID, err error) {
	body := map[string]string{
		"username":   username,
		"password":   password,
		"public_key": publicKey,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/auth/device")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("server returned error: %s", resp.Status())
	}

	// Преобразуем строку в UUID
	deviceUUID, err = uuid.Parse(resp.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid device UUID returned: %w", err)
	}

	return deviceUUID, nil
}

// Login отправляет запрос на логин пользователя с указанным deviceUUID и возвращает JWT токен.
func Login(
	ctx context.Context,
	client *resty.Client,
	username, password string,
	deviceUUID uuid.UUID,
) (token string, err error) {
	body := map[string]string{
		"username":    username,
		"password":    password,
		"device_uuid": deviceUUID.String(),
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/auth/login")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("server returned error: %s", resp.Status())
	}

	// JWT возвращается в заголовке Authorization: "Bearer <token>"
	token = resp.Header().Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no Authorization header returned")
	}

	return token, nil
}
