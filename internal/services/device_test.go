package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	myerrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)
	ctx := context.Background()
	userUUID := "user-123"
	publicKey := "test-public-key"

	// Пользователь существует
	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(&models.UserDB{UserUUID: userUUID}, nil)

	// Сохранение устройства
	mockDeviceWriter.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)
	assert.NoError(t, err)
	assert.NotNil(t, deviceUUID)
}

func TestDeviceService_Register_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)
	ctx := context.Background()
	userUUID := "user-123"
	publicKey := "test-public-key"

	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(nil, nil)

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)
	assert.Error(t, err)
	assert.Equal(t, myerrors.ErrUserNotFound, err)
	assert.Nil(t, deviceUUID)
}

func TestDeviceService_Register_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)
	ctx := context.Background()
	userUUID := "user-123"
	publicKey := "test-public-key"

	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(&models.UserDB{UserUUID: userUUID}, nil)

	mockDeviceWriter.EXPECT().
		Save(ctx, gomock.Any()).
		Return(errors.New("db error"))

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Nil(t, deviceUUID)
}

func TestDeviceService_Register_GetUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockDeviceUserReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)

	service := NewDeviceService(mockUserReader, mockDeviceWriter)
	ctx := context.Background()
	userUUID := "user-123"
	publicKey := "test-public-key"

	mockUserReader.EXPECT().
		GetByUUID(ctx, userUUID).
		Return(nil, errors.New("db query failed"))

	deviceUUID, err := service.Register(ctx, userUUID, publicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db query failed")
	assert.Nil(t, deviceUUID)
}
