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

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)
	handler := RegisterHandler(mockReg)

	username := "user123"
	password := "P@ssw0rd"
	userUUID := uuid.New()

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: `successful registration`,
			body: `{"username":"` + username + `","password":"` + password + `"}`,
			mockSetup: func() {
				mockReg.EXPECT().Register(gomock.Any(), username, password).Return(userUUID.String(), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   userUUID.String(),
		},
		{
			name:           "invalid JSON",
			body:           `{"username":`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: `user already exists`,
			body: `{"username":"` + username + `","password":"` + password + `"}`,
			mockSetup: func() {
				mockReg.EXPECT().Register(gomock.Any(), username, password).Return("", internalErrors.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: `internal server error`,
			body: `{"username":"` + username + `","password":"` + password + `"}`,
			mockSetup: func() {
				mockReg.EXPECT().Register(gomock.Any(), username, password).Return("", errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tt.body))
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

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogin := NewMockLoginer(ctrl)
	handler := LoginHandler(mockLogin)

	username := "user123"
	password := "P@ssw0rd"
	deviceUUID := uuid.New().String()
	token := "token123"

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		expectedStatus int
		expectedHeader map[string]string
	}{
		{
			name: `successful login`,
			body: `{"username":"` + username + `","password":"` + password + `","device_uuid":"` + deviceUUID + `"}`,
			mockSetup: func() {
				mockLogin.EXPECT().Login(gomock.Any(), username, password, deviceUUID).Return(token, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{"Authorization": "Bearer " + token},
		},
		{
			name:           "invalid JSON",
			body:           `{"username":`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: `unauthorized user`,
			body: `{"username":"` + username + `","password":"` + password + `","device_uuid":"` + deviceUUID + `"}`,
			mockSetup: func() {
				mockLogin.EXPECT().Login(gomock.Any(), username, password, deviceUUID).Return("", internalErrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: `internal server error`,
			body: `{"username":"` + username + `","password":"` + password + `","device_uuid":"` + deviceUUID + `"}`,
			mockSetup: func() {
				mockLogin.EXPECT().Login(gomock.Any(), username, password, deviceUUID).Return("", errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedHeader != nil {
				for k, v := range tt.expectedHeader {
					assert.Equal(t, v, resp.Header.Get(k))
				}
			}
		})
	}
}
