package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	internalErrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"
	password := "password123"
	token := "jwt-token"

	// Пользователь не существует
	mockUR.EXPECT().Get(gomock.Any(), username).Return(nil, nil)

	// Save вызывается с любым UUID, username и хэшем
	mockUW.EXPECT().Save(gomock.Any(), gomock.Any(), username, gomock.Any()).Return(nil)

	// TokenGenerator.Generate возвращает токен
	mockTG.EXPECT().Generate(gomock.Any()).Return(token, nil)

	gotToken, err := svc.Register(context.Background(), username, password)

	assert.NoError(t, err)
	assert.Equal(t, token, gotToken)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"
	existingUser := &models.UserDB{UserUUID: "uuid-123", Username: username, PasswordHash: "hash"}

	mockUR.EXPECT().Get(gomock.Any(), username).Return(existingUser, nil)

	gotToken, err := svc.Register(context.Background(), username, "password123")

	assert.ErrorIs(t, err, internalErrors.ErrUserAlreadyExists)
	assert.Empty(t, gotToken)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"
	password := "password123"
	userUUID := "uuid-123"
	token := "jwt-token"

	// Сначала создаём валидный хэш
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.UserDB{UserUUID: userUUID, Username: username, PasswordHash: string(hash)}

	mockUR.EXPECT().Get(gomock.Any(), username).Return(user, nil)
	mockTG.EXPECT().Generate(userUUID).Return(token, nil)

	gotToken, err := svc.Login(context.Background(), username, password)

	assert.NoError(t, err)
	assert.Equal(t, token, gotToken)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"

	mockUR.EXPECT().Get(gomock.Any(), username).Return(nil, nil)

	gotToken, err := svc.Login(context.Background(), username, "password123")

	assert.ErrorIs(t, err, internalErrors.ErrUserNotFound)
	assert.Empty(t, gotToken)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"
	password := "password123"
	userUUID := "uuid-123"

	// Хэш от другого пароля
	hash, _ := bcrypt.GenerateFromPassword([]byte("wrongpass"), bcrypt.DefaultCost)
	user := &models.UserDB{UserUUID: userUUID, Username: username, PasswordHash: string(hash)}

	mockUR.EXPECT().Get(gomock.Any(), username).Return(user, nil)

	gotToken, err := svc.Login(context.Background(), username, password)

	assert.ErrorIs(t, err, internalErrors.ErrInvalidPassword)
	assert.Empty(t, gotToken)
}

func TestAuthService_Login_TokenGenerationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockTG)

	username := "testuser"
	password := "password123"
	userUUID := "uuid-123"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.UserDB{UserUUID: userUUID, Username: username, PasswordHash: string(hash)}

	mockUR.EXPECT().Get(gomock.Any(), username).Return(user, nil)
	mockTG.EXPECT().Generate(userUUID).Return("", errors.New("token error"))

	gotToken, err := svc.Login(context.Background(), username, password)

	assert.Error(t, err)
	assert.Empty(t, gotToken)
}
