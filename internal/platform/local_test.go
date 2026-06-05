// Unit tests for local execution detection.
package platform_test

import (
	"testing"

	"github.com/phides-code/go-multi-api/internal/platform"
)

func TestLocalMode(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"true", "true", true},
		{"one", "1", true},
		{"false", "false", false},
		{"empty", "", false},
		{"other", "yes", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_SAM_LOCAL", tc.value)
			if got := platform.LocalMode(); got != tc.want {
				t.Fatalf("LocalMode() = %v, want %v", got, tc.want)
			}
		})
	}
}
