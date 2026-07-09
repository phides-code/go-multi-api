// Shared banana test fixtures for handler and DynamoDB tests.
package testutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
)

// TestBananaContent is the canonical valid content in handler and DynamoDB tests.
const TestBananaContent = "ripe"

// TestStoredBananaCreatedOn is a fixed timestamp for persisted-banana repository tests.
const TestStoredBananaCreatedOn uint64 = 12345

const (
	ListBananaContentFirst  = "first"
	ListBananaContentSecond = "second"
	ListBananaContentThird  = "third"
)

// BananaWithID returns a banana whose ID matches the returned id string.
func BananaWithID(content string, createdOn uint64) (id string, b banana.Banana) {
	id = uuid.NewString()
	b = banana.Banana{ID: id, Content: content, CreatedOn: createdOn}
	return
}

// BananaCreateBody returns JSON for a valid create/update request body.
func BananaCreateBody(content string) string {
	return fmt.Sprintf(`{"content":%q}`, content)
}

// ListBananas returns three list items for repository list tests.
// When withTimestamps is true, CreatedOn is set to 1, 2, and 3 respectively.
func ListBananas(withTimestamps bool) (first, second, third banana.Banana) {
	first = banana.Banana{ID: uuid.NewString(), Content: ListBananaContentFirst}
	second = banana.Banana{ID: uuid.NewString(), Content: ListBananaContentSecond}
	third = banana.Banana{ID: uuid.NewString(), Content: ListBananaContentThird}
	if withTimestamps {
		first.CreatedOn = 1
		second.CreatedOn = 2
		third.CreatedOn = 3
	}
	return
}
