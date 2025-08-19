package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockUserReader(ctrl)
	mockWriter := services.NewMockUserWriter(ctrl)
	mockToken := services.NewMockTokenGenerator(ctrl)

	svc := services.NewAuthService(mockReader, mockWriter, mockToken)

	ctx := context.Background()

	tests := []struct {
		name          string
		username      string
		password      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "successful registration",
			username: "alice",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().Get(ctx, "alice").Return(nil, nil)
				mockWriter.EXPECT().Save(ctx, gomock.Any(), "alice", gomock.Any()).Return(nil)
				mockToken.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("token123", nil)
			},
			expectedError: nil,
		},
		{
			name:     "user already exists",
			username: "bob",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().Get(ctx, "bob").Return(&models.UserDB{}, nil)
			},
			expectedError: services.ErrUserAlreadyExists,
		},
		{
			name:     "reader returns error",
			username: "charlie",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().Get(ctx, "charlie").Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:     "writer returns error",
			username: "dave",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().Get(ctx, "dave").Return(nil, nil)
				mockWriter.EXPECT().Save(ctx, gomock.Any(), "dave", gomock.Any()).Return(errors.New("save error"))
			},
			expectedError: errors.New("save error"),
		},
		{
			name:     "token generator returns error",
			username: "eve",
			password: "password123",
			setupMocks: func() {
				mockReader.EXPECT().Get(ctx, "eve").Return(nil, nil)
				mockWriter.EXPECT().Save(ctx, gomock.Any(), "eve", gomock.Any()).Return(nil)
				mockToken.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("token error"))
			},
			expectedError: errors.New("token error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := svc.Register(ctx, tt.username, tt.password)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
