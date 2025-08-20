package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	internalErrors "github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestDeviceRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := handlers.NewMockDeviceRegisterer(ctrl)

	validUserUUID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	validDeviceUUID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success",
			requestBody: `{"user_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","public_key":"ssh-rsa AAAAB3Nza..."}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), validUserUUID, "ssh-rsa AAAAB3Nza...").
					Return(&validDeviceUUID, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		},
		{
			name:        "user not found",
			requestBody: `{"user_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","public_key":"ssh-rsa AAAAB3Nza..."}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), validUserUUID, "ssh-rsa AAAAB3Nza...").
					Return(nil, internalErrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			requestBody:    `{"user_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","public_key":`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "invalid uuid",
			requestBody:    `{"user_uuid":"not-a-uuid","public_key":"ssh-rsa AAAAB3Nza..."}`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:        "internal error",
			requestBody: `{"user_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","public_key":"ssh-rsa AAAAB3Nza..."}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), validUserUUID, "ssh-rsa AAAAB3Nza...").
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/devices/register", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			handler := handlers.DeviceRegisterHandler(mockRegisterer)
			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
			}
			assert.Equal(t, tt.expectedBody, buf.String())
		})
	}
}
