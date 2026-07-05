package handler_test

import (
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

// existingBananaFixture returns an ID-linked banana and matching PUT body for get/update/delete tests.
func existingBananaFixture() (id string, banana domain.Banana, updateBody string) {
	id, banana = testutil.BananaWithID(testutil.TestBananaContent, 0)
	updateBody = testutil.BananaCreateBody(testutil.TestBananaContent)
	return
}
