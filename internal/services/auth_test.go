package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)

	svc := NewAuthService(mockGetter, mockSaver)

	tests := []struct {
		name        string
		username    string
		password    string
		setupMocks  func()
		expectError bool
		expectedErr error
	}{
		{
			name:     "successful registration",
			username: "johndoe",
			password: "secret",
			setupMocks: func() {
				// No existing user
				mockGetter.EXPECT().
					Get(gomock.Any(), "johndoe").
					Return(nil, nil)
				// Save will be called
				mockSaver.EXPECT().
					Save(gomock.Any(), gomock.Any(), "johndoe", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:     "username already exists",
			username: "johndoe",
			password: "secret",
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "johndoe").
					Return(&models.UserDB{}, nil)
			},
			expectError: true,
			expectedErr: ErrUsernameAlreadyExists,
		},
		{
			name:     "getter returns error",
			username: "johndoe",
			password: "secret",
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "johndoe").
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectedErr: errors.New("db error"),
		},
		{
			name:     "saver returns error",
			username: "johndoe",
			password: "secret",
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "johndoe").
					Return(nil, nil)
				mockSaver.EXPECT().
					Save(gomock.Any(), gomock.Any(), "johndoe", gomock.Any()).
					Return(errors.New("db save error"))
			},
			expectError: true,
			expectedErr: errors.New("db save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			userUUID, err := svc.Register(context.Background(), tt.username, tt.password)
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				assert.Equal(t, uuid.Nil, userUUID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userUUID)
			}
		})
	}
}
