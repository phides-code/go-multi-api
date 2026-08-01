// Shared boolean error-expectation asserts for table-driven validation tests.
package testutil

import "testing"

// AssertWantErr fails when whether err is nil does not match wantErr.
func AssertWantErr(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Fatal("expected error")
	}
	if !wantErr && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
