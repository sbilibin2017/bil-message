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
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockUserReader(ctrl)
	mockWriter := services.NewMockUserWriter(ctrl)

	svc := services.NewAuthService(mockReader, mockWriter)

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
