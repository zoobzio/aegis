//go:build testing

package aegis

import (
	"context"
	"testing"
)

func TestCallerFromContextNoPeer(t *testing.T) {
	ctx := context.Background()
	_, err := CallerFromContext(ctx)
	if err != ErrNoPeerInfo {
		t.Errorf("expected ErrNoPeerInfo, got %v", err)
	}
}

func TestMustCallerFromContextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	ctx := context.Background()
	MustCallerFromContext(ctx)
}
