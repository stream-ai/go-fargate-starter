package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/service/server"
)

func TestIntegration(t *testing.T) {
	// Start the server in a separate goroutine
	go func() {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
		addr := fmt.Sprintf(":8080")
		err := server.Run(context.Background(), logger, addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running server: %s\n", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// path, expectedStatusCode, expectedBody
	// Make a request to the server
	resp, err := http.Get("http://localhost:8080/healthz")
	assert.NoError(t, err)
	defer resp.Body.Close()
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
	}
	assert.Equal(t, "OK", string(body))
	// Assert the response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func RepoRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to determine repository root directory: %v", err)
	}
	return strings.TrimSpace(string(output))
}
