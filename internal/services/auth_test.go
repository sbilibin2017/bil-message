package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUserWriteService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := NewMockUserReader(ctrl)
	mockUW := NewMockUserWriter(ctrl)
	mockCW := NewMockClientWriter(ctrl)
	mockTG := NewMockTokenGenerator(ctrl)

	svc := NewAuthService(mockUR, mockUW, mockCW, mockTG)

	t.Run("success", func(t *testing.T) {
		username := "testuser"
		password := "secret"
		token := "jwt-token"

		// Expect user not to exist
		mockUR.EXPECT().Get(gomock.Any(), username).Return(nil, nil)
		// Expect Save user
		mockUW.EXPECT().Save(gomock.Any(), gomock.Any(), username, gomock.Any()).Return(nil)
		// Expect Save client
		mockCW.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any(), "").Return(nil)
		// Expect Generate token
		mockTG.EXPECT().Generate(gomock.Any(), gomock.Any()).Return(token, nil)

		tk, err := svc.Register(context.Background(), username, password)
		assert.NoError(t, err)
		assert.Equal(t, token, tk)
	})

	t.Run("user exists", func(t *testing.T) {
		username := "existinguser"
		password := "secret"

		mockUR.EXPECT().Get(gomock.Any(), username).Return(&models.UserDB{}, nil)

		tk, err := svc.Register(context.Background(), username, password)
		assert.Error(t, err)
		assert.EqualError(t, err, "user already exists")
		assert.Empty(t, tk)
	})

	t.Run("Get error", func(t *testing.T) {
		username := "user"
		password := "pass"

		mockUR.EXPECT().Get(gomock.Any(), username).Return(nil, errors.New("db error"))

		tk, err := svc.Register(context.Background(), username, password)
		assert.Error(t, err)
		assert.EqualError(t, err, "db error")
		assert.Empty(t, tk)
	})
}
