package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	internalErrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestDeviceRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockDeviceRegisterer(ctrl)
	handler := DeviceRegisterHandler(mockReg)

	validUUID := uuid.New().String()
	publicKey := "ssh-rsa AAAAB3..."

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			body: `{"user_uuid":"` + validUUID + `","public_key":"` + publicKey + `"}`,
			mockSetup: func() {
				deviceUUID := "device-123"
				mockReg.EXPECT().
					Register(gomock.Any(), validUUID, publicKey).
					Return(&deviceUUID, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "device-123",
		},
		{
			name:           "invalid JSON",
			body:           `{"user_uuid":`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid user UUID",
			body:           `{"user_uuid":"not-a-uuid","public_key":"` + publicKey + `"}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			body: `{"user_uuid":"` + validUUID + `","public_key":"` + publicKey + `"}`,
			mockSetup: func() {
				mockReg.EXPECT().
					Register(gomock.Any(), validUUID, publicKey).
					Return(nil, internalErrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "internal server error",
			body: `{"user_uuid":"` + validUUID + `","public_key":"` + publicKey + `"}`,
			mockSetup: func() {
				mockReg.EXPECT().
					Register(gomock.Any(), validUUID, publicKey).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest(http.MethodPost, "/devices/register", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				assert.Equal(t, tt.expectedBody, buf.String())
			}
		})
	}
}
