package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/handlers"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := handlers.NewMockRegisterer(ctrl)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success",
			requestBody: `{"username":"user1","password":"pass1"}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user1", "pass1").
					Return(uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "11111111-1111-1111-1111-111111111111",
		},
		{
			name:        "user already exists",
			requestBody: `{"username":"user1","password":"pass1"}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user1", "pass1").
					Return(uuid.Nil, services.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			requestBody:    `{"username":"user1","password":`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:        "internal error",
			requestBody: `{"username":"user1","password":"pass1"}`,
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user1", "pass1").
					Return(uuid.Nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			handler := handlers.RegisterHandler(mockRegisterer)
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
