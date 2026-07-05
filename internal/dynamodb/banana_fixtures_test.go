package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func storedBananaFixture(t *testing.T) (id string, banana domain.Banana, item map[string]types.AttributeValue) {
	t.Helper()
	id, banana = testutil.BananaWithID(testutil.TestBananaContent, testutil.TestStoredBananaCreatedOn)
	var err error
	item, err = attributevalue.MarshalMap(banana)
	if err != nil {
		t.Fatal(err)
	}
	return
}
