package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Register отправляет запрос на регистрацию пользователя и возвращает user_uuid.
func Register(
	ctx context.Context,
	client *resty.Client,
	username string,
	password string,
) (*string, error) {
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
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("server returned error: %s", resp.Status())
	}

	// Ответ сервера — просто UUID в теле
	userUUID := resp.String()
	return &userUUID, nil
}
