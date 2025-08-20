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
					Save(gomock.Any(), gomock.Any(), "newuser", gomock.Any()).
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
					Save(gomock.Any(), gomock.Any(), "newuser2", gomock.Any()).
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

	mockReader := services.NewMockUserReader(ctrl)
	mockDevice := services.NewMockDeviceReader(ctrl)
	mockToken := services.NewMockTokenGenerator(ctrl)

	svc := services.NewAuthService(mockReader, nil, mockDevice, mockToken)

	userUUID := uuid.New()
	deviceUUID := uuid.New()
	password := "password123"
	passwordHashBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	passwordHash := string(passwordHashBytes)

	tests := []struct {
		name          string
		username      string
		passwordInput string
		setupMocks    func()
		expectedError error
	}{
		{
			name:          "successful login",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID, PasswordHash: passwordHash}, nil)

				mockDevice.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID).
					Return(&models.DeviceDB{UserUUID: userUUID}, nil)

				mockToken.EXPECT().
					Generate(userUUID, deviceUUID).
					Return("token123", nil)
			},
			expectedError: nil,
		},
		{
			name:          "user not found",
			username:      "unknown",
			passwordInput: password,
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "unknown").
					Return(nil, nil)
			},
			expectedError: internalErrors.ErrUserNotFound,
		},
		{
			name:          "invalid password",
			username:      "user1",
			passwordInput: "wrongpassword",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID, PasswordHash: passwordHash}, nil)
			},
			expectedError: internalErrors.ErrInvalidPassword,
		},
		{
			name:          "device not found",
			username:      "user1",
			passwordInput: password,
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), "user1").
					Return(&models.UserDB{UserUUID: userUUID, PasswordHash: passwordHash}, nil)

				mockDevice.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID).
					Return(nil, nil)
			},
			expectedError: internalErrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := svc.Login(context.Background(), tt.username, tt.passwordInput, deviceUUID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Equal(t, "", token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "token123", token)
			}
		})
	}
}

func TestAuthService_Login_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockUserReader(ctrl)
	mockDeviceReader := services.NewMockDeviceReader(ctrl)
	mockTokenGen := services.NewMockTokenGenerator(ctrl)

	svc := services.NewAuthService(mockReader, nil, mockDeviceReader, mockTokenGen)

	username := "testuser"
	password := "password123"
	deviceUUID := uuid.New()
	userUUID := uuid.New()

	// Helper to generate bcrypt hash
	hashPassword := func(pw string) string {
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			t.Fatal(err) // fail the test if hashing fails
		}
		return string(hash)
	}

	tests := []struct {
		name          string
		setupMocks    func()
		expectedError error
	}{
		{
			name: "user not found",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(nil, nil)
			},
			expectedError: internalErrors.ErrUserNotFound,
		},
		{
			name: "reader returns error",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "invalid password",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(&models.UserDB{
						UserUUID:     userUUID,
						PasswordHash: "$2a$10$invalidhashforbcrypttest", // invalid hash
					}, nil)
			},
			expectedError: internalErrors.ErrInvalidPassword,
		},
		{
			name: "device not found",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(&models.UserDB{
						UserUUID:     userUUID,
						PasswordHash: hashPassword(password),
					}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID).
					Return(nil, nil)
			},
			expectedError: internalErrors.ErrDeviceNotFound,
		},
		{
			name: "device belongs to different user",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(&models.UserDB{
						UserUUID:     userUUID,
						PasswordHash: hashPassword(password),
					}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID).
					Return(&models.DeviceDB{
						UserUUID: uuid.New(), // different user
					}, nil)
			},
			expectedError: internalErrors.ErrDeviceNotFound,
		},
		{
			name: "token generation error",
			setupMocks: func() {
				mockReader.EXPECT().
					GetByUsername(gomock.Any(), username).
					Return(&models.UserDB{
						UserUUID:     userUUID,
						PasswordHash: hashPassword(password),
					}, nil)
				mockDeviceReader.EXPECT().
					GetByUUID(gomock.Any(), deviceUUID).
					Return(&models.DeviceDB{
						UserUUID: userUUID,
					}, nil)
				mockTokenGen.EXPECT().
					Generate(userUUID, deviceUUID).
					Return("", errors.New("token error"))
			},
			expectedError: errors.New("token error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := svc.Login(context.Background(), username, password, deviceUUID)
			assert.Equal(t, "", token)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Login_DeviceReaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockUserReader(ctrl)
	mockDeviceReader := services.NewMockDeviceReader(ctrl)
	mockTokenGen := services.NewMockTokenGenerator(ctrl)

	svc := services.NewAuthService(mockReader, nil, mockDeviceReader, mockTokenGen)

	username := "testuser"
	password := "password123"
	deviceUUID := uuid.New()
	userUUID := uuid.New()

	// Helper to generate bcrypt hash
	hashPassword := func(pw string) string {
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(hash)
	}

	// Mock: user exists
	mockReader.EXPECT().
		GetByUsername(gomock.Any(), username).
		Return(&models.UserDB{
			UserUUID:     userUUID,
			PasswordHash: hashPassword(password),
		}, nil)

	// Mock: device reader returns an error
	mockDeviceReader.EXPECT().
		GetByUUID(gomock.Any(), deviceUUID).
		Return(nil, errors.New("device db error"))

	// Call the Login method
	token, err := svc.Login(context.Background(), username, password, deviceUUID)

	// Assertions
	assert.Equal(t, "", token)
	assert.EqualError(t, err, "device db error")
}
