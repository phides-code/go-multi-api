// Package-local banana fixtures for handler tests (ID-linked entity and request bodies).
package banana_test

import (
	"strings"
	"testing"

	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type bananaValidationBodies struct {
	bananaWithEmptyValue      string
	bananaWithWhitespace      string
	bananaWithValueTooLong    string
	bananaWithRatingBelowMin  string
	bananaWithRatingAboveMax  string
}

func newBananaValidationBodies(t *testing.T) bananaValidationBodies {
	t.Helper()

	bananaWithEmptyValue := testutil.ValidBananaBody()
	bananaWithEmptyValue.Color = ""

	bananaWithWhitespace := testutil.ValidBananaBody()
	bananaWithWhitespace.Color = "   "

	bananaWithValueTooLong := testutil.ValidBananaBody()
	bananaWithValueTooLong.Color = strings.Repeat("a", domain.DefaultMaxStringLength+1)

	bananaWithRatingBelowMin := testutil.ValidBananaBody()
	bananaWithRatingBelowMin.Rating = domain.DefaultMinInt - 1

	bananaWithRatingAboveMax := testutil.ValidBananaBody()
	bananaWithRatingAboveMax.Rating = domain.DefaultMaxInt + 1

	return bananaValidationBodies{
		bananaWithEmptyValue:     bananaWithEmptyValue.JSON(t),
		bananaWithWhitespace:     bananaWithWhitespace.JSON(t),
		bananaWithValueTooLong:   bananaWithValueTooLong.JSON(t),
		bananaWithRatingBelowMin: bananaWithRatingBelowMin.JSON(t),
		bananaWithRatingAboveMax: bananaWithRatingAboveMax.JSON(t),
	}
}

// existingBananaFixture returns an ID-linked banana and matching PUT body for get/update/delete tests.
func existingBananaFixture(t *testing.T) (id string, b banana.Banana, updateBody string) {
	t.Helper()
	id, b = testutil.BananaWithID(testutil.ValidBananaBody(), 0)
	updateBody = testutil.ValidBananaBody().JSON(t)
	return
}
