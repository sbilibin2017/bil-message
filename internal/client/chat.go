package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// CreateChat отправляет запрос на создание новой комнаты и возвращает UUID комнаты
func CreateChat(ctx context.Context, client *resty.Client, token string) (uuid.UUID, error) {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Post("/chat")
	if err != nil {
		return uuid.Nil, err
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("server returned error: %s", resp.Status())
	}

	roomUUID, err := uuid.Parse(resp.String())
	if err != nil {
		return uuid.Nil, err
	}

	return roomUUID, nil
}

// RemoveChat удаляет комнату по UUID
func RemoveChat(ctx context.Context, client *resty.Client, token string, roomUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Delete("/chat/" + roomUUID.String())
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// AddChatMember добавляет текущего пользователя в указанную комнату
func AddChatMember(ctx context.Context, client *resty.Client, token string, roomUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Post("/chat/" + roomUUID.String() + "/member")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// RemoveChatMember удаляет текущего пользователя из указанной комнаты
func RemoveChatMember(ctx context.Context, client *resty.Client, token string, roomUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Delete("/chat/" + roomUUID.String() + "/member")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}
