// Unit tests for UUID validation and generation.
package domain_test

import (
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

func TestNewID(t *testing.T) {
	t.Parallel()

	id := domain.NewID()
	if err := domain.ValidateID(id); err != nil {
		t.Fatalf("expected generated id to be valid uuid: %v", err)
	}
}
