// Unit tests for Banana validation and ID helpers.
package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
)

func TestValidateID(t *testing.T) {
	t.Parallel()

	if err := domain.ValidateID(uuid.NewString()); err != nil {
		t.Fatalf("expected valid uuid, got %v", err)
	}

	if err := domain.ValidateID("not-a-uuid"); err == nil {
		t.Fatal("expected invalid id error")
	}
}

func TestValidateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "valid", content: "hello", wantErr: false},
		{name: "empty", content: "", wantErr: true},
		{name: "whitespace", content: "   ", wantErr: true},
		{name: "max length", content: strings.Repeat("a", domain.MaxContentLength), wantErr: false},
		{name: "too long", content: strings.Repeat("a", domain.MaxContentLength+1), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateContent(tt.content)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewID(t *testing.T) {
	t.Parallel()

	id := domain.NewID()
	if err := domain.ValidateID(id); err != nil {
		t.Fatalf("expected generated id to be valid uuid: %v", err)
	}
}
