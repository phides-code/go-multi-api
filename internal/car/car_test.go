// Unit tests for car create/update validation.
package car_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	validCreateInput := func() car.CreateInput {
		return car.CreateInput{Model: testutil.TestCarModel}
	}

	tests := []struct {
		name    string
		input   car.CreateInput
		wantErr bool
	}{
		{name: "valid", input: validCreateInput(), wantErr: false},
		{
			name: "empty model",
			input: func() car.CreateInput {
				in := validCreateInput()
				in.Model = ""
				return in
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := car.ValidateCreateInput(tt.input)
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
	validUpdateInput := func() car.UpdateInput {
		return car.UpdateInput{ID: validID, Model: testutil.TestCarModel}
	}

	tests := []struct {
		name    string
		input   car.UpdateInput
		wantErr bool
	}{
		{name: "valid", input: validUpdateInput(), wantErr: false},
		{
			name: "invalid id",
			input: func() car.UpdateInput {
				in := validUpdateInput()
				in.ID = "bad"
				return in
			}(),
			wantErr: true,
		},
		{
			name: "empty model",
			input: func() car.UpdateInput {
				in := validUpdateInput()
				in.Model = ""
				return in
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := car.ValidateUpdateInput(tt.input)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
