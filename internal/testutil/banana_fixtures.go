// Shared banana test fixtures for handler and DynamoDB tests.
package testutil

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
)

// TestBananaVariety is the canonical valid variety in handler and DynamoDB tests.
const TestBananaVariety = "yellow"

// TestBananaRating is the canonical valid rating in handler and DynamoDB tests.
const TestBananaRating = 50

// TestStoredBananaCreatedOn is a fixed timestamp for persisted-banana repository tests.
const TestStoredBananaCreatedOn uint64 = 12345

const (
	ListBananaVarietyFirst  = "yellow"
	ListBananaVarietySecond = "green"
	ListBananaVarietyThird  = "brown"
)

const (
	ListBananaRatingFirst  = 10
	ListBananaRatingSecond = 20
	ListBananaRatingThird  = 30
)

// BananaBody is the create/update request payload for banana HTTP tests.
// Declared independently of banana.Banana so tag regressions surface as test failures.
type BananaBody struct {
	Variety  string `json:"variety"`
	Rating int    `json:"rating"`
}

// ValidBananaBody returns a BananaBody with canonical valid field values.
func ValidBananaBody() BananaBody {
	return BananaBody{Variety: TestBananaVariety, Rating: TestBananaRating}
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
		Variety:     body.Variety,
		Rating:    body.Rating,
		CreatedOn: createdOn,
	}
	return
}

// ListBananas returns three list items for repository list tests.
// When withTimestamps is true, CreatedOn is set to 1, 2, and 3 respectively.
func ListBananas(withTimestamps bool) (first, second, third banana.Banana) {
	first = banana.Banana{
		ID:     uuid.NewString(),
		Variety:  ListBananaVarietyFirst,
		Rating: ListBananaRatingFirst,
	}
	second = banana.Banana{
		ID:     uuid.NewString(),
		Variety:  ListBananaVarietySecond,
		Rating: ListBananaRatingSecond,
	}
	third = banana.Banana{
		ID:     uuid.NewString(),
		Variety:  ListBananaVarietyThird,
		Rating: ListBananaRatingThird,
	}
	if withTimestamps {
		first.CreatedOn = 1
		second.CreatedOn = 2
		third.CreatedOn = 3
	}
	return
}
