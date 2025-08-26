package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserWrite := NewMockUserWriter(ctrl)
	mockUserRead := NewMockUserReader(ctrl)

	svc := NewAuthService(mockUserWrite, mockUserRead, nil, nil, nil)

	tests := []struct {
		name        string
		username    string
		password    string
		mockGetUser *models.UserDB
		mockErr     error
		wantErr     error
	}{
		{
			name:        "success",
			username:    "johndoe",
			password:    "pass123",
			mockGetUser: nil,
			mockErr:     nil,
			wantErr:     nil,
		},
		{
			name:        "user exists",
			username:    "johndoe",
			password:    "pass123",
			mockGetUser: &models.UserDB{UserUUID: uuid.New(), Username: "johndoe"},
			mockErr:     nil,
			wantErr:     ErrUserExists,
		},
		{
			name:        "get user error",
			username:    "johndoe",
			password:    "pass123",
			mockGetUser: nil,
			mockErr:     errors.New("db error"),
			wantErr:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRead.EXPECT().Get(gomock.Any(), tt.username).Return(tt.mockGetUser, tt.mockErr)
			if tt.mockGetUser == nil && tt.mockErr == nil {
				mockUserWrite.EXPECT().Save(gomock.Any(), gomock.Any(), tt.username, gomock.Any()).Return(nil)
			}

			userUUID, err := svc.Register(context.Background(), tt.username, tt.password)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Equal(t, uuid.Nil, userUUID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userUUID)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRead := NewMockUserReader(ctrl)
	mockDeviceRead := NewMockDeviceReader(ctrl)
	mockTokenGen := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(nil, mockUserRead, nil, mockDeviceRead, mockTokenGen)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	user := &models.UserDB{UserUUID: uuid.New(), Username: "johndoe", Password: string(passwordHash)}
	device := &models.DeviceDB{DeviceUUID: uuid.New(), UserUUID: user.UserUUID}

	tests := []struct {
		name         string
		username     string
		password     string
		deviceUUID   uuid.UUID
		mockUser     *models.UserDB
		mockDevice   *models.DeviceDB
		mockUserErr  error
		mockDevErr   error
		mockToken    string
		mockTokenErr error
		wantErr      error
	}{
		{
			name:       "success",
			username:   "johndoe",
			password:   "pass123",
			deviceUUID: device.DeviceUUID,
			mockUser:   user,
			mockDevice: device,
			mockToken:  "token123",
			wantErr:    nil,
		},
		{
			name:       "user not found",
			username:   "johndoe",
			password:   "pass123",
			deviceUUID: device.DeviceUUID,
			mockUser:   nil,
			wantErr:    ErrUserNotFound,
		},
		{
			name:       "device not found",
			username:   "johndoe",
			password:   "pass123",
			deviceUUID: device.DeviceUUID,
			mockUser:   user,
			mockDevice: nil,
			wantErr:    ErrDeviceNotFound,
		},
		{
			name:       "invalid password",
			username:   "johndoe",
			password:   "wrong",
			deviceUUID: device.DeviceUUID,
			mockUser:   user,
			mockDevice: device,
			wantErr:    ErrInvalidCredential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRead.EXPECT().Get(gomock.Any(), tt.username).Return(tt.mockUser, nil)
			if tt.mockUser != nil {
				mockDeviceRead.EXPECT().Get(gomock.Any(), tt.deviceUUID).Return(tt.mockDevice, nil)
			}
			if tt.mockUser != nil && tt.mockDevice != nil && tt.password == "pass123" {
				mockTokenGen.EXPECT().Generate(tt.mockUser.UserUUID, tt.deviceUUID).Return(tt.mockToken, nil)
			}

			token, err := svc.Login(context.Background(), tt.username, tt.password, tt.deviceUUID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, "", token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockToken, token)
			}
		})
	}
}

func TestAuthService_AddDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRead := NewMockUserReader(ctrl)
	mockDeviceWrite := NewMockDeviceWriter(ctrl)

	svc := NewAuthService(nil, mockUserRead, mockDeviceWrite, nil, nil)

	// Создадим хэш пароля для тестового пользователя
	password := "pass123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userUUID := uuid.New()
	user := &models.UserDB{
		UserUUID: userUUID,
		Username: "johndoe",
		Password: string(passwordHash),
	}

	tests := []struct {
		name        string
		username    string
		password    string
		publicKey   string
		mockUser    *models.UserDB
		mockGetErr  error
		mockSaveErr error
		wantErr     error
	}{
		{
			name:      "success",
			username:  "johndoe",
			password:  password,
			publicKey: "pubkey123",
			mockUser:  user,
			wantErr:   nil,
		},
		{
			name:      "user not found",
			username:  "unknown",
			password:  password,
			publicKey: "pubkey123",
			mockUser:  nil,
			wantErr:   ErrUserNotFound,
		},
		{
			name:      "invalid password",
			username:  "johndoe",
			password:  "wrongpass",
			publicKey: "pubkey123",
			mockUser:  user,
			wantErr:   ErrInvalidCredential,
		},
		{
			name:        "save error",
			username:    "johndoe",
			password:    password,
			publicKey:   "pubkey123",
			mockUser:    user,
			mockSaveErr: errors.New("db save error"),
			wantErr:     errors.New("db save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRead.EXPECT().Get(gomock.Any(), tt.username).Return(tt.mockUser, tt.mockGetErr)
			if tt.mockUser != nil && tt.mockGetErr == nil && tt.wantErr == nil {
				mockDeviceWrite.EXPECT().Save(gomock.Any(), gomock.Any(), tt.mockUser.UserUUID, tt.publicKey).Return(nil)
			} else if tt.mockUser != nil && tt.mockGetErr == nil && tt.mockSaveErr != nil {
				mockDeviceWrite.EXPECT().Save(gomock.Any(), gomock.Any(), tt.mockUser.UserUUID, tt.publicKey).Return(tt.mockSaveErr)
			}

			deviceUUID, err := svc.AddDevice(context.Background(), tt.username, tt.password, tt.publicKey)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Equal(t, uuid.Nil, deviceUUID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, deviceUUID)
			}
		})
	}
}
