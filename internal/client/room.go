package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type RoomClient struct {
	client *resty.Client
}

func NewRoomClient(client *resty.Client) *RoomClient {
	return &RoomClient{
		client: client,
	}
}

// CreateRoom создаёт новую комнату и возвращает её UUID
func (c *RoomClient) CreateRoom(ctx context.Context, token string) (uuid.UUID, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		Post("/rooms")
	if err != nil {
		return uuid.Nil, err
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("create room failed: %s", resp.Status())
	}

	roomUUID, err := uuid.Parse(string(resp.Body()))
	if err != nil {
		return uuid.Nil, err
	}

	return roomUUID, nil
}

// DeleteRoom удаляет комнату по UUID
func (c *RoomClient) DeleteRoom(ctx context.Context, token string, roomUUID uuid.UUID) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		Delete(fmt.Sprintf("/rooms/%s", roomUUID.String()))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("delete room failed: %s", resp.Status())
	}

	return nil
}

// AddMember добавляет участника в комнату
func (c *RoomClient) AddMember(ctx context.Context, token string, roomUUID, memberUUID uuid.UUID) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		Post(fmt.Sprintf("/rooms/%s/%s", roomUUID.String(), memberUUID.String()))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("add member failed: %s", resp.Status())
	}

	return nil
}

// RemoveMember удаляет участника из комнаты
func (c *RoomClient) RemoveMember(ctx context.Context, token string, roomUUID, memberUUID uuid.UUID) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		Post(fmt.Sprintf("/rooms/%s/%s", roomUUID.String(), memberUUID.String()))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("remove member failed: %s", resp.Status())
	}

	return nil
}
