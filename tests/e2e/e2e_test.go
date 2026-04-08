//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var testURL string

func TestMain(m *testing.M) {
	testURL = os.Getenv("PILLAR_TEST_URL")
	if testURL == "" {
		testURL = "http://localhost:8080"
	}

	// Wait for server to be healthy.
	if err := waitForHealthy(testURL, 30*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "server not healthy at %s: %v\n", testURL, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForHealthy(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out after %s", timeout)
}
