package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRegisterer(ctrl)

	tests := []struct {
		name           string
		reqBody        RegisterRequest
		mockReturn     uuid.UUID
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			reqBody:        RegisterRequest{Username: "user1", Password: "pass"},
			mockReturn:     uuid.New(),
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user exists",
			reqBody:        RegisterRequest{Username: "user1", Password: "pass"},
			mockReturn:     uuid.Nil,
			mockErr:        services.ErrUserExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "bad request empty username",
			reqBody:        RegisterRequest{Username: "", Password: "pass"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErr != nil || tt.mockReturn != uuid.Nil {
				mockSvc.EXPECT().Register(gomock.Any(), tt.reqBody.Username, tt.reqBody.Password).Return(tt.mockReturn, tt.mockErr)
			}

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			NewRegisterHandler(mockSvc)(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code == http.StatusOK {
				assert.Equal(t, tt.mockReturn.String(), w.Body.String())
			}
		})
	}
}

func TestAddDeviceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockDeviceAdder(ctrl)

	tests := []struct {
		name           string
		reqBody        AddDeviceRequest
		mockReturn     uuid.UUID
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			reqBody:        AddDeviceRequest{Username: "user1", Password: "pass", PublicKey: "key"},
			mockReturn:     uuid.New(),
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			reqBody:        AddDeviceRequest{Username: "user1", Password: "pass", PublicKey: "key"},
			mockReturn:     uuid.Nil,
			mockErr:        services.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad request empty username",
			reqBody:        AddDeviceRequest{Username: "", Password: "pass", PublicKey: "key"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErr != nil || tt.mockReturn != uuid.Nil {
				mockSvc.EXPECT().AddDevice(gomock.Any(), tt.reqBody.Username, tt.reqBody.Password, tt.reqBody.PublicKey).Return(tt.mockReturn, tt.mockErr)
			}

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/device", bytes.NewReader(body))
			w := httptest.NewRecorder()

			NewAddDeviceHandler(mockSvc)(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code == http.StatusOK {
				assert.Equal(t, tt.mockReturn.String(), w.Body.String())
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockLoginer(ctrl)

	tests := []struct {
		name           string
		reqBody        LoginRequest
		mockReturn     string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			reqBody:        LoginRequest{Username: "user", Password: "pass", DeviceUUID: uuid.New().String()},
			mockReturn:     "token123",
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user/device not found",
			reqBody:        LoginRequest{Username: "user", Password: "pass", DeviceUUID: uuid.New().String()},
			mockErr:        services.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid credentials",
			reqBody:        LoginRequest{Username: "user", Password: "wrong", DeviceUUID: uuid.New().String()},
			mockErr:        services.ErrInvalidCredential,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request invalid UUID",
			reqBody:        LoginRequest{Username: "user", Password: "pass", DeviceUUID: "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErr != nil || tt.mockReturn != "" {
				deviceUUID, _ := uuid.Parse(tt.reqBody.DeviceUUID)
				mockSvc.EXPECT().Login(gomock.Any(), tt.reqBody.Username, tt.reqBody.Password, deviceUUID).Return(tt.mockReturn, tt.mockErr)
			}

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			NewLoginHandler(mockSvc)(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code == http.StatusOK {
				assert.Equal(t, "Bearer "+tt.mockReturn, w.Header().Get("Authorization"))
			}
		})
	}
}
