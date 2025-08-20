package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	myerrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceService_Add_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)

	ctx := context.Background()
	userUUID := uuid.New()
	publicKey := "test-public-key"

	// Ожидания: пользователь найден
	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(&models.UserDB{UserUUID: userUUID}, nil)

	// Save должен вызваться с любым UUID (deviceUUID)
	mockDeviceWriter.EXPECT().
		Save(ctx, gomock.Any(), userUUID, publicKey).
		Return(nil)

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)

	assert.NoError(t, err)
	assert.NotNil(t, deviceUUID)
}

func TestDeviceService_Add_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)

	ctx := context.Background()
	userUUID := uuid.New()
	publicKey := "test-public-key"

	// Ожидания: пользователь не найден
	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(nil, nil)

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)

	assert.Error(t, err)
	assert.Equal(t, myerrors.ErrUserNotFound, err)
	assert.Nil(t, deviceUUID)
}

func TestDeviceService_Add_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)

	ctx := context.Background()
	userUUID := uuid.New()
	publicKey := "test-public-key"

	// Пользователь найден
	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(&models.UserDB{UserUUID: userUUID}, nil)

	// Ошибка при сохранении устройства
	mockDeviceWriter.EXPECT().
		Save(ctx, gomock.Any(), userUUID, publicKey).
		Return(errors.New("db error"))

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Nil(t, deviceUUID)
}

func TestDeviceService_Add_GetUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)

	ctx := context.Background()
	userUUID := uuid.New()
	publicKey := "test-public-key"

	// Ошибка при получении пользователя
	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(nil, errors.New("db query failed"))

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db query failed")
	assert.Nil(t, deviceUUID)
}
