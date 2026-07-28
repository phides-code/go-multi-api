// Unit tests for banana create/update validation.
package banana_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	validCreateInput := func() banana.CreateInput {
		return banana.CreateInput{
			Variety:  testutil.TestBananaVariety,
			Rating: testutil.TestBananaRating,
		}
	}

	tests := []struct {
		name    string
		input   banana.CreateInput
		wantErr bool
	}{
		{name: "valid", input: validCreateInput(), wantErr: false},
		{
			name: "empty variety",
			input: func() banana.CreateInput {
				in := validCreateInput()
				in.Variety = ""
				return in
			}(),
			wantErr: true,
		},
		{
			name: "rating below min",
			input: func() banana.CreateInput {
				in := validCreateInput()
				in.Rating = domain.DefaultMinInt - 1
				return in
			}(),
			wantErr: true,
		},
		{
			name: "rating above max",
			input: func() banana.CreateInput {
				in := validCreateInput()
				in.Rating = domain.DefaultMaxInt + 1
				return in
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := banana.ValidateCreateInput(tt.input)

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

	validUpdateInput := func() banana.UpdateInput {
		return banana.UpdateInput{
			ID:     validID,
			Variety:  testutil.TestBananaVariety,
			Rating: testutil.TestBananaRating,
		}
	}

	tests := []struct {
		name    string
		input   banana.UpdateInput
		wantErr bool
	}{
		{name: "valid", input: validUpdateInput(), wantErr: false},
		{
			name: "invalid id",
			input: func() banana.UpdateInput {
				in := validUpdateInput()
				in.ID = "bad"
				return in
			}(),
			wantErr: true,
		},
		{
			name: "empty variety",
			input: func() banana.UpdateInput {
				in := validUpdateInput()
				in.Variety = ""
				return in
			}(),
			wantErr: true,
		},
		{
			name: "rating below min",
			input: func() banana.UpdateInput {
				in := validUpdateInput()
				in.Rating = domain.DefaultMinInt - 1
				return in
			}(),
			wantErr: true,
		},
		{
			name: "rating above max",
			input: func() banana.UpdateInput {
				in := validUpdateInput()
				in.Rating = domain.DefaultMaxInt + 1
				return in
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := banana.ValidateUpdateInput(tt.input)

			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
