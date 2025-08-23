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
		wantBody   string
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
					Return(uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "11111111-1111-1111-1111-111111111111",
		},
		{
			name:       "invalid JSON",
			reqBody:    "{invalid-json",
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name: "empty username",
			reqBody: RegisterRequest{
				Username: "",
				Password: "secret",
			},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest, // <- must return 400 before calling Register
			wantBody:   "",
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
					Return(uuid.UUID{}, errors.New("service failure"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "",
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
			body := w.Body.String()

			require.Equal(t, tt.wantStatus, resp.StatusCode)
			require.Equal(t, tt.wantBody, body)
		})
	}
}
