// Unit tests for shared string validation helpers.
package domain_test

import (
	"strings"
	"testing"

	"github.com/phides-code/go-multi-api/internal/domain"
)

func TestValidateRequiredString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "hello", wantErr: false},
		{name: "empty", value: "", wantErr: true},
		{name: "whitespace", value: "   ", wantErr: true},
		{name: "max length", value: strings.Repeat("a", domain.DefaultMaxStringLength), wantErr: false},
		{name: "too long", value: strings.Repeat("a", domain.DefaultMaxStringLength+1), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateRequiredString(tt.value, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
