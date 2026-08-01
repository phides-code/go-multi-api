// Package-local fixtures for banana handler tests (existing item + invalid request bodies).
package banana_test

import (
	"strings"
	"testing"

	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

// bananaValidationBodies holds invalid create/update JSON for handler client-error tests.
// Names describe the invalidation shape (empty, whitespace, too long, …), not which
// entity field was mutated — keeps find-replace and new fields simple.
type bananaValidationBodies struct {
	bananaWithEmptyValue    string
	bananaWithWhitespace    string
	bananaWithValueTooLong  string
	bananaWithValueBelowMin string
	bananaWithValueAboveMax string
}

func newBananaValidationBodies(t *testing.T) bananaValidationBodies {
	t.Helper()

	emptyValue := testutil.ValidBananaBody()
	emptyValue.Descriptor = ""

	whitespace := testutil.ValidBananaBody()
	whitespace.Descriptor = "   "

	valueTooLong := testutil.ValidBananaBody()
	valueTooLong.Descriptor = strings.Repeat("a", domain.DefaultMaxStringLength+1)

	valueBelowMin := testutil.ValidBananaBody()
	valueBelowMin.Rating = domain.DefaultMinInt - 1

	valueAboveMax := testutil.ValidBananaBody()
	valueAboveMax.Rating = domain.DefaultMaxInt + 1

	return bananaValidationBodies{
		bananaWithEmptyValue:    emptyValue.JSON(t),
		bananaWithWhitespace:    whitespace.JSON(t),
		bananaWithValueTooLong:  valueTooLong.JSON(t),
		bananaWithValueBelowMin: valueBelowMin.JSON(t),
		bananaWithValueAboveMax: valueAboveMax.JSON(t),
	}
}

// existingBananaFixture returns a UUID, matching banana entity, and valid PUT JSON body.
func existingBananaFixture(t *testing.T) (id string, b banana.Banana, updateBody string) {
	t.Helper()
	id, b = testutil.BananaWithID(testutil.ValidBananaBody(), 0)
	updateBody = testutil.ValidBananaBody().JSON(t)
	return
}
