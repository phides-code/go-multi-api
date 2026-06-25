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

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   domain.CreateBananaInput
		wantErr bool
	}{
		{name: "valid", input: domain.CreateBananaInput{Content: "hello"}, wantErr: false},
		{name: "empty content", input: domain.CreateBananaInput{Content: ""}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateCreateInput(tt.input)

			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUpdateInput(t *testing.T) {
	t.Parallel()

	validID := uuid.NewString()

	tests := []struct {
		name    string
		input   domain.UpdateBananaInput
		wantErr bool
	}{
		{name: "valid", input: domain.UpdateBananaInput{ID: validID, Content: "hello"}, wantErr: false},
		{name: "invalid id", input: domain.UpdateBananaInput{ID: "bad", Content: "hello"}, wantErr: true},
		{name: "empty content", input: domain.UpdateBananaInput{ID: validID, Content: ""}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateUpdateInput(tt.input)

			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
