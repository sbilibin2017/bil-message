package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// Register отправляет запрос на регистрацию пользователя и возвращает user_uuid.
func Register(
	ctx context.Context,
	client *resty.Client,
	username string,
	password string,
) (userUUID uuid.UUID, err error) {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/auth/register")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("server returned error: %s", resp.Status())
	}

	// Преобразуем строку из тела ответа в UUID
	userUUID, err = uuid.Parse(strings.TrimSpace(resp.String()))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user UUID returned: %w", err)
	}

	return userUUID, nil
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

	deviceUUID, err = uuid.Parse(strings.TrimSpace(resp.String()))
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

	// Убираем префикс "Bearer "
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	return token, nil
}
