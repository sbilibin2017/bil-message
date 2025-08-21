package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	internalErrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockUserReader(ctrl)
	mockWriter := services.NewMockUserWriter(ctrl)

	svc := services.NewAuthService(mockReader, mockWriter, nil, nil)

	tests := []struct {
		name          string
		username      string
		password      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "successful registration",
			username: "newuser",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "newuser").
					Return(nil, nil)
				mockWriter.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "user already exists",
			username: "existinguser",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "existinguser").
					Return(&models.UserDB{}, nil)
			},
			expectedError: internalErrors.ErrUserAlreadyExists,
		},
		{
			name:     "reader returns error",
			username: "erroruser",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "erroruser").
					Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:     "writer returns error",
			username: "newuser2",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "newuser2").
					Return(nil, nil)
				mockWriter.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(errors.New("save error"))
			},
			expectedError: errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			userUUID, err := svc.Register(context.Background(), tt.username, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Equal(t, "", userUUID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, userUUID)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := services.NewMockUserReader(ctrl)
	mockDeviceReader := services.NewMockDeviceReader(ctrl)
	mockTokenGen := services.NewMockTokenGenerator(ctrl)

	svc := services.NewAuthService(mockUserReader, nil, mockDeviceReader, mockTokenGen)

	userUUID := uuid.New()
	deviceUUID := uuid.New()
	password := "password123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		username      string
		passwordInput string
		setupMocks    func()
		expectedToken string
		expectedError error
	}{
		{
			name:          "successful login",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID.String()).
					Return(&models.DeviceDB{DeviceUUID: deviceUUID.String(), UserUUID: userUUID.String()}, nil)
				mockTokenGen.EXPECT().
					Generate(&models.TokenPayload{UserUUID: userUUID.String(), DeviceUUID: deviceUUID.String()}).
					Return("token123", nil)
			},
			expectedToken: "token123",
			expectedError: nil,
		},
		{
			name:          "user not found",
			username:      "unknown",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "unknown").
					Return(nil, nil)
			},
			expectedToken: "",
			expectedError: internalErrors.ErrUserNotFound,
		},
		{
			name:          "invalid password",
			username:      "user1",
			passwordInput: "wrongpassword",
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
			},
			expectedToken: "",
			expectedError: internalErrors.ErrInvalidPassword,
		},
		{
			name:          "device not found",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID.String()).
					Return(nil, nil)
			},
			expectedToken: "",
			expectedError: internalErrors.ErrDeviceNotFound,
		},
		{
			name:          "device belongs to different user",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID.String()).
					Return(&models.DeviceDB{DeviceUUID: deviceUUID.String(), UserUUID: uuid.New().String()}, nil)
			},
			expectedToken: "",
			expectedError: internalErrors.ErrDeviceNotFound,
		},
		{
			name:          "token generation error",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID.String()).
					Return(&models.DeviceDB{DeviceUUID: deviceUUID.String(), UserUUID: userUUID.String()}, nil)
				mockTokenGen.EXPECT().
					Generate(&models.TokenPayload{UserUUID: userUUID.String(), DeviceUUID: deviceUUID.String()}).
					Return("", errors.New("token error"))
			},
			expectedToken: "",
			expectedError: errors.New("token error"),
		},
		{
			name:          "device reader returns error",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockUserReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID.String(), PasswordHash: string(passwordHash)}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID.String()).
					Return(nil, errors.New("device db error"))
			},
			expectedToken: "",
			expectedError: errors.New("device db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := svc.Login(context.Background(), tt.username, tt.passwordInput, deviceUUID.String())
			assert.Equal(t, tt.expectedToken, token)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
