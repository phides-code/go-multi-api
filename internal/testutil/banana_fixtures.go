// Shared banana test fixtures for handler and DynamoDB tests.
package testutil

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
)

// TestBananaColor is the canonical valid color in handler and DynamoDB tests.
const TestBananaColor = "yellow"

// TestStoredBananaCreatedOn is a fixed timestamp for persisted-banana repository tests.
const TestStoredBananaCreatedOn uint64 = 12345

const (
	ListBananaColorFirst  = "yellow"
	ListBananaColorSecond = "green"
	ListBananaColorThird  = "brown"
)

// BananaBody is the create/update request payload for banana HTTP tests.
// Declared independently of banana.Banana so tag regressions surface as test failures.
type BananaBody struct {
	Color string `json:"color"`
}

// ValidBananaBody returns a BananaBody with canonical valid field values.
func ValidBananaBody() BananaBody {
	return BananaBody{Color: TestBananaColor}
}

// JSON marshals the body to a request payload string.
func (b BananaBody) JSON(t *testing.T) string {
	t.Helper()
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal banana body: %v", err)
	}
	return string(data)
}

// BananaWithID returns a banana whose ID matches the returned id string.
// Pass a BananaBody (or ValidBananaBody()) so client fields stay named.
func BananaWithID(body BananaBody, createdOn uint64) (id string, b banana.Banana) {
	id = uuid.NewString()
	b = banana.Banana{
		ID:        id,
		Color:     body.Color,
		CreatedOn: createdOn,
	}
	return
}

// ListBananas returns three list items for repository list tests.
// When withTimestamps is true, CreatedOn is set to 1, 2, and 3 respectively.
func ListBananas(withTimestamps bool) (first, second, third banana.Banana) {
	first = banana.Banana{
		ID:    uuid.NewString(),
		Color: ListBananaColorFirst,
	}
	second = banana.Banana{
		ID:    uuid.NewString(),
		Color: ListBananaColorSecond,
	}
	third = banana.Banana{
		ID:    uuid.NewString(),
		Color: ListBananaColorThird,
	}
	if withTimestamps {
		first.CreatedOn = 1
		second.CreatedOn = 2
		third.CreatedOn = 3
	}
	return
}
