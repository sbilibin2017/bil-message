package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type AuthClient struct {
	client  *resty.Client
	baseURL string
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		client:  resty.New(),
		baseURL: baseURL,
	}
}

// Register делает POST-запрос на /auth/register и возвращает UUID нового пользователя
func (c *AuthClient) Register(ctx context.Context, username, password string) (uuid.UUID, error) {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(c.baseURL + "/auth/register")
	if err != nil {
		return uuid.Nil, err
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("register failed: %s", resp.Status())
	}

	userUUID, err := uuid.Parse(string(resp.Body()))
	if err != nil {
		return uuid.Nil, err
	}

	return userUUID, nil
}

// AddDevice делает POST-запрос на /auth/device и возвращает UUID нового устройства
func (c *AuthClient) AddDevice(ctx context.Context, username, password, publicKey string) (uuid.UUID, error) {
	reqBody := map[string]string{
		"username":   username,
		"password":   password,
		"public_key": publicKey,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(c.baseURL + "/auth/device")
	if err != nil {
		return uuid.Nil, err
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("add device failed: %s", resp.Status())
	}

	deviceUUID, err := uuid.Parse(string(resp.Body()))
	if err != nil {
		return uuid.Nil, err
	}

	return deviceUUID, nil
}

// Login делает POST-запрос на /auth/login и возвращает JWT токен
func (c *AuthClient) Login(ctx context.Context, username, password string, deviceUUID uuid.UUID) (string, error) {
	reqBody := map[string]string{
		"username":    username,
		"password":    password,
		"device_uuid": deviceUUID.String(),
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(c.baseURL + "/auth/login")
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("login failed: %s", resp.Status())
	}

	token := resp.Header().Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("missing Authorization header in response")
	}

	return token, nil
}
