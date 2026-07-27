// Package-local car fixtures for handler tests.
package car_test

import (
	"strings"
	"testing"

	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type carValidationBodies struct {
	carWithEmptyValue   string
	carWithWhitespace   string
	carWithValueTooLong string
}

func newCarValidationBodies(t *testing.T) carValidationBodies {
	t.Helper()

	empty := testutil.ValidCarBody()
	empty.Model = ""

	whitespace := testutil.ValidCarBody()
	whitespace.Model = "   "

	tooLong := testutil.ValidCarBody()
	tooLong.Model = strings.Repeat("a", domain.DefaultMaxStringLength+1)

	return carValidationBodies{
		carWithEmptyValue:   empty.JSON(t),
		carWithWhitespace:   whitespace.JSON(t),
		carWithValueTooLong: tooLong.JSON(t),
	}
}

func existingCarFixture(t *testing.T) (id string, c car.Car, updateBody string) {
	t.Helper()
	id, c = testutil.CarWithID(testutil.ValidCarBody(), 0)
	updateBody = testutil.ValidCarBody().JSON(t)
	return
}
