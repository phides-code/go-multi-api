// Unit tests for shared string and integer validation helpers.
package domain_test

import (
	"strings"
	"testing"

	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
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
			testutil.AssertWantErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateRequiredInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{name: "valid mid", value: 50, wantErr: false},
		{name: "min", value: domain.DefaultMinInt, wantErr: false},
		{name: "max", value: domain.DefaultMaxInt, wantErr: false},
		{name: "below min", value: domain.DefaultMinInt - 1, wantErr: true},
		{name: "above max", value: domain.DefaultMaxInt + 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateRequiredInt(tt.value, domain.DefaultMinInt, domain.DefaultMaxInt)
			testutil.AssertWantErr(t, err, tt.wantErr)
		})
	}
}
