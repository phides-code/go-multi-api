// DynamoDB test fixture: persisted car row plus marshaled item for Get/Delete mocks.
package car_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func storedCarFixture(t *testing.T) (id string, c car.Car, item map[string]types.AttributeValue) {
	t.Helper()
	id, c = testutil.CarWithID(testutil.ValidCarBody(), testutil.TestStoredCarCreatedOn)
	var err error
	item, err = attributevalue.MarshalMap(c)
	if err != nil {
		t.Fatal(err)
	}
	return
}
