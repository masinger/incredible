package cmd

import (
	"io"
	"strings"
	"testing"
)

func TestNegativeProviderTimeoutIsRejected(t *testing.T) {
	timeout := executionOptions.ProviderTimeout
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		executionOptions.ProviderTimeout = timeout
	})
	rootCmd.SetArgs([]string{"--provider-timeout=-1s", "true"})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)

	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--provider-timeout") {
		t.Errorf("err = %v, want an error about --provider-timeout", err)
	}
}
