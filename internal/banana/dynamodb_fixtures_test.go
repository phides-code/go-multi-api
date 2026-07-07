// DynamoDB test fixture: persisted banana row plus marshaled item for Get/Delete mocks.
package banana_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func storedBananaFixture(t *testing.T) (id string, b banana.Banana, item map[string]types.AttributeValue) {
	t.Helper()
	id, b = testutil.BananaWithID(testutil.TestBananaContent, testutil.TestStoredBananaCreatedOn)
	var err error
	item, err = attributevalue.MarshalMap(b)
	if err != nil {
		t.Fatal(err)
	}
	return
}
