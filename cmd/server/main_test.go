package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintBuildInfo(t *testing.T) {
	// Backup original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set build info
	buildVersion = "v1.2.3"
	buildCommit = "abcdef123"
	buildDate = "2025-08-22"

	printBuildInfo()

	// Close writer and read captured output
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	output := buf.String()

	assert.Contains(t, output, "Build Information:")
	assert.Contains(t, output, "Version : v1.2.3")
	assert.Contains(t, output, "Commit  : abcdef123")
	assert.Contains(t, output, "Date    : 2025-08-22")
}

// TestParseFlags verifies that command line flags are parsed correctly.
func TestParseFlags(t *testing.T) {
	address = ""
	databaseDSN = ""
	jwtSecretKey = ""
	jwtExp = 0

	err := parseFlags()
	assert.NoError(t, err)
	assert.NotEmpty(t, address)
	assert.NotEmpty(t, databaseDSN)
	assert.NotEmpty(t, jwtSecretKey)
	assert.NotZero(t, jwtExp)
}

// TestRun initializes components and runs server in a goroutine, then cancels context to simulate shutdown.
func TestRun(t *testing.T) {
	// Use in-memory SQLite to avoid external DB dependency
	driver := "sqlite"
	dsn := ":memory:"
	address := "127.0.0.1:0" // random free port
	jwtKey := "test-jwt-key"
	jwtExp := 1 * time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Cancel context after short delay to trigger shutdown
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, driver, dsn, address, jwtKey, jwtExp)

	// Since we canceled context, run should exit cleanly
	assert.NoError(t, err)
}

// TestRunWithBadDSN ensures run returns error for invalid database DSN.
func TestRunWithBadDSN(t *testing.T) {
	ctx := context.Background()
	err := run(ctx, "postgres", "invalid_dsn", "127.0.0.1:0", "key", 1*time.Hour)
	assert.Error(t, err)
}

// TestRunWithInvalidDriver ensures run returns error for unsupported DB driver.
func TestRunWithInvalidDriver(t *testing.T) {
	ctx := context.Background()
	err := run(ctx, "unknown_driver", "test", "127.0.0.1:0", "key", 1*time.Hour)
	assert.Error(t, err)
}
