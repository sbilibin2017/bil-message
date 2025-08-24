package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRegisterer(ctrl)

	tests := []struct {
		name       string
		reqBody    interface{}
		mockSetup  func()
		wantStatus int
	}{
		{
			name: "successful registration",
			reqBody: RegisterRequest{
				Username: "johndoe",
				Password: "secret",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Register(gomock.Any(), "johndoe", "secret").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			reqBody:    "{invalid-json",
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty username",
			reqBody: RegisterRequest{
				Username: "",
				Password: "secret",
			},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "username already exists",
			reqBody: RegisterRequest{
				Username: "johndoe",
				Password: "secret",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Register(gomock.Any(), "johndoe", "secret").
					Return(services.ErrUsernameAlreadyExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "service error",
			reqBody: RegisterRequest{
				Username: "johndoe",
				Password: "secret",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Register(gomock.Any(), "johndoe", "secret").
					Return(errors.New("service failure"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch v := tt.reqBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			RegisterHandler(mockSvc).ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			require.Empty(t, w.Body.String())
		})
	}
}

func TestAddDeviceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockDeviceAdder(ctrl)
	deviceUUIDExample := uuid.New()

	tests := []struct {
		name       string
		reqBody    interface{}
		mockSetup  func()
		wantStatus int
		wantBody   string
	}{
		{
			name: "successful add device",
			reqBody: DeviceRequest{
				Username:  "johndoe",
				Password:  "secret",
				PublicKey: "pubkey",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					AddDevice(gomock.Any(), "johndoe", "secret", "pubkey").
					Return(deviceUUIDExample, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   deviceUUIDExample.String(),
		},
		{
			name:       "invalid JSON",
			reqBody:    "{invalid-json",
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty fields",
			reqBody: DeviceRequest{
				Username: "",
				Password: "secret",
			},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			reqBody: DeviceRequest{
				Username:  "johndoe",
				Password:  "secret",
				PublicKey: "pubkey",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					AddDevice(gomock.Any(), "johndoe", "secret", "pubkey").
					Return(uuid.Nil, services.ErrInvalidCredentials)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			reqBody: DeviceRequest{
				Username:  "johndoe",
				Password:  "secret",
				PublicKey: "pubkey",
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					AddDevice(gomock.Any(), "johndoe", "secret", "pubkey").
					Return(uuid.Nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch v := tt.reqBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			// Исправляем путь: добавляем ведущий слэш
			req := httptest.NewRequest(http.MethodPost, "/auth/device", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			AddDeviceHandler(mockSvc).ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusOK {
				require.Equal(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockLoginer(ctrl)
	validUUID := uuid.New()

	tests := []struct {
		name       string
		reqBody    interface{}
		mockSetup  func()
		wantStatus int
		wantAuth   string
	}{
		{
			name: "successful login",
			reqBody: LoginRequest{
				Username:   "johndoe",
				Password:   "secret",
				DeviceUUID: validUUID.String(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "johndoe", "secret", validUUID).
					Return("jwt-token-123", nil)
			},
			wantStatus: http.StatusOK,
			wantAuth:   "Bearer jwt-token-123",
		},
		{
			name:       "invalid JSON",
			reqBody:    "{invalid-json",
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty fields",
			reqBody: LoginRequest{
				Username:   "",
				Password:   "secret",
				DeviceUUID: validUUID.String(),
			},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid UUID",
			reqBody: LoginRequest{
				Username:   "johndoe",
				Password:   "secret",
				DeviceUUID: "not-a-uuid",
			},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			reqBody: LoginRequest{
				Username:   "johndoe",
				Password:   "secret",
				DeviceUUID: validUUID.String(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "johndoe", "secret", validUUID).
					Return("", services.ErrInvalidCredentials)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			reqBody: LoginRequest{
				Username:   "johndoe",
				Password:   "secret",
				DeviceUUID: validUUID.String(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "johndoe", "secret", validUUID).
					Return("", errors.New("service failure"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch v := tt.reqBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			LoginHandler(mockSvc).ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusOK {
				require.Equal(t, tt.wantAuth, w.Header().Get("Authorization"))
			}
		})
	}
}
