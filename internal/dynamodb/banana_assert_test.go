// Banana-specific DynamoDB repository test helpers.
package dynamodb_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/phides-code/go-multi-api/internal/domain"
)

// assertBananaPutItem checks Create's PutItem: no-overwrite condition and item payload.
func assertBananaPutItem(t *testing.T, params *awsdynamodb.PutItemInput, want domain.Banana) {
	t.Helper()

	if params.ConditionExpression == nil || *params.ConditionExpression != "attribute_not_exists(id)" {
		t.Fatalf("ConditionExpression = %v, want attribute_not_exists(id)", params.ConditionExpression)
	}

	var got domain.Banana
	if err := attributevalue.UnmarshalMap(params.Item, &got); err != nil {
		t.Fatalf("unmarshal item: %v", err)
	}
	if got != want {
		t.Fatalf("Item = %+v, want %+v", got, want)
	}
}

// assertBananaRepoResult checks got/want banana and error (errors.Is) for single-entity repo methods.
func assertBananaRepoResult(t *testing.T, op string, got domain.Banana, err error, want domain.Banana, wantErr error) {
	t.Helper()

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		return
	}

	if err != nil {
		t.Fatalf("%s: %v", op, err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
