// Package-local banana fixtures for handler tests (ID-linked entity and request bodies).
package banana_test

import (
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

// existingBananaFixture returns an ID-linked banana and matching PUT body for get/update/delete tests.
func existingBananaFixture() (id string, b banana.Banana, updateBody string) {
	id, b = testutil.BananaWithID(testutil.TestBananaContent, 0)
	updateBody = testutil.BananaCreateBody(testutil.TestBananaContent)
	return
}
