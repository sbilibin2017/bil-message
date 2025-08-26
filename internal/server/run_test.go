package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRunServerStartsAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "127.0.0.1:8086"
	version := "/api/v1"
	dsn := "file::memory:?cache=shared"
	jwtSecret := "test-secret"
	jwtExp := time.Hour

	// Запуск сервера в фоне
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, addr, version, "sqlite", dsn, jwtSecret, jwtExp)
	}()

	// Ждем немного, чтобы сервер успел стартовать
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервер отвечает на HEAD-запрос
	req, err := http.NewRequest("HEAD", "http://"+addr+version+"/auth/register", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Останавливаем сервер
	cancel()

	// Проверяем, что Run завершился без ошибок
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop in time")
	}
}
