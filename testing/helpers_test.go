//go:build testing

package testing

import (
	"errors"
	"testing"
	"time"
)

func TestEventually(t *testing.T) {
	t.Run("condition met immediately", func(t *testing.T) {
		Eventually(t, func() bool { return true }, 100*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("condition met after delay", func(t *testing.T) {
		start := time.Now()
		Eventually(t, func() bool {
			return time.Since(start) > 50*time.Millisecond
		}, 200*time.Millisecond, 10*time.Millisecond)
	})
}

func TestRequireNoError(t *testing.T) {
	t.Run("nil error passes", func(t *testing.T) {
		RequireNoError(t, nil)
	})
}

func TestRequireError(t *testing.T) {
	t.Run("non-nil error passes", func(t *testing.T) {
		RequireError(t, errors.New("test error"))
	})
}
