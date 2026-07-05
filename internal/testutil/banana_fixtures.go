// Shared banana test fixtures for handler and DynamoDB tests.
package testutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
)

// TestBananaContent is the canonical valid content in handler and DynamoDB tests.
const TestBananaContent = "ripe"

// TestStoredBananaCreatedOn is a fixed timestamp for persisted-banana repository tests.
const TestStoredBananaCreatedOn uint64 = 12345

const (
	ListBananaContentFirst  = "first"
	ListBananaContentSecond = "second"
	ListBananaContentPage2  = "page2"
)

// BananaWithID returns a banana whose ID matches the returned id string.
func BananaWithID(content string, createdOn uint64) (id string, banana domain.Banana) {
	id = uuid.NewString()
	banana = domain.Banana{ID: id, Content: content, CreatedOn: createdOn}
	return
}

// BananaCreateBody returns JSON for a valid create/update request body.
func BananaCreateBody(content string) string {
	return fmt.Sprintf(`{"content":%q}`, content)
}

// ListBananaPage returns two list items and a third for cursor follow-up tests.
// When withTimestamps is true, CreatedOn is set to 1, 2, and 3 respectively.
func ListBananaPage(withTimestamps bool) (first, second, page2 domain.Banana) {
	first = domain.Banana{ID: uuid.NewString(), Content: ListBananaContentFirst}
	second = domain.Banana{ID: uuid.NewString(), Content: ListBananaContentSecond}
	page2 = domain.Banana{ID: uuid.NewString(), Content: ListBananaContentPage2}
	if withTimestamps {
		first.CreatedOn = 1
		second.CreatedOn = 2
		page2.CreatedOn = 3
	}
	return
}
