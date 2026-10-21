package cli

import (
	"testing"
	"time"
)

// SetWaitDelay changes waitDelay for the rest of the test.
func SetWaitDelay(t testing.TB, d time.Duration) {
	t.Helper()
	old := waitDelay
	waitDelay = d
	t.Cleanup(func() {
		waitDelay = old
	})
}
