// Unit tests for X-CF-Token header validation.
package platform_test

import (
	"testing"

	"github.com/phides-code/go-multi-api/internal/platform"
)

func TestValidCFTToken(t *testing.T) {
	t.Parallel()

	headers := map[string]string{"X-CF-Token": "secret"}

	if !platform.ValidCFTToken("secret", headers) {
		t.Fatal("expected token to match")
	}
	if platform.ValidCFTToken("wrong", headers) {
		t.Fatal("expected token mismatch")
	}
	if platform.ValidCFTToken("", headers) {
		t.Fatal("expected empty expected token to fail")
	}
	if platform.ValidCFTToken("secret", map[string]string{}) {
		t.Fatal("expected missing header to fail")
	}
}

func TestHeaderValueCaseInsensitive(t *testing.T) {
	t.Parallel()

	headers := map[string]string{"x-cf-token": "value"}
	if got := platform.HeaderValue(headers, "X-CF-Token"); got != "value" {
		t.Fatalf("header value = %q, want %q", got, "value")
	}
}
