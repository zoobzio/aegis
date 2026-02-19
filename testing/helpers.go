//go:build testing

// Package testing provides test helpers for aegis.
package testing

import (
	"testing"
	"time"
)

// Eventually retries a condition function until it returns true or times out.
func Eventually(t *testing.T, condition func() bool, timeout, interval time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatal("condition not met within timeout")
}

// RequireNoError fails the test immediately if err is not nil.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// RequireError fails the test immediately if err is nil.
func RequireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
