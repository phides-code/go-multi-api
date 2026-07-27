// Shared car test fixtures for handler and DynamoDB tests.
package testutil

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/car"
)

// TestCarModel is the canonical valid model in handler and DynamoDB tests.
const TestCarModel = "civic"

// TestStoredCarCreatedOn is a fixed timestamp for persisted-car repository tests.
const TestStoredCarCreatedOn uint64 = 12345

const (
	ListCarModelFirst  = "civic"
	ListCarModelSecond = "accord"
	ListCarModelThird  = "pilot"
)

// CarBody is the create/update request payload for car HTTP tests.
// Declared independently of car.Car so tag regressions surface as test failures.
type CarBody struct {
	Model string `json:"model"`
}

// ValidCarBody returns a CarBody with canonical valid field values.
func ValidCarBody() CarBody {
	return CarBody{Model: TestCarModel}
}

// JSON marshals the body to a request payload string.
func (b CarBody) JSON(t *testing.T) string {
	t.Helper()
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal car body: %v", err)
	}
	return string(data)
}

// CarWithID returns a car whose ID matches the returned id string.
func CarWithID(body CarBody, createdOn uint64) (id string, c car.Car) {
	id = uuid.NewString()
	c = car.Car{
		ID:        id,
		Model:     body.Model,
		CreatedOn: createdOn,
	}
	return
}

// ListCars returns three list items for repository list tests.
func ListCars(withTimestamps bool) (first, second, third car.Car) {
	first = car.Car{ID: uuid.NewString(), Model: ListCarModelFirst}
	second = car.Car{ID: uuid.NewString(), Model: ListCarModelSecond}
	third = car.Car{ID: uuid.NewString(), Model: ListCarModelThird}
	if withTimestamps {
		first.CreatedOn = 1
		second.CreatedOn = 2
		third.CreatedOn = 3
	}
	return
}
