// Unit tests for Banana create/update validation wiring.
package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
)

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   domain.CreateBananaInput
		wantErr bool
	}{
		{name: "valid", input: domain.CreateBananaInput{Content: "hello", Variety: "cavendish"}, wantErr: false},
		{name: "empty content", input: domain.CreateBananaInput{Content: ""}, wantErr: true},
		{name: "empty variety", input: domain.CreateBananaInput{Content: "hello", Variety: ""}, wantErr: true},
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
		{name: "valid", input: domain.UpdateBananaInput{ID: validID, Content: "hello", Variety: "cavendish"}, wantErr: false},
		{name: "invalid id", input: domain.UpdateBananaInput{ID: "bad", Content: "hello"}, wantErr: true},
		{name: "empty content", input: domain.UpdateBananaInput{ID: validID, Content: ""}, wantErr: true},
		{name: "empty variety", input: domain.UpdateBananaInput{ID: validID, Content: "hello", Variety: ""}, wantErr: true},
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
