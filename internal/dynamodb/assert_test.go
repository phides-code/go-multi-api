// Shared and banana-specific DynamoDB repository test helpers.
package dynamodb_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/domain"
)

// assertUpdateSets checks that UpdateItem SETs exactly the given string attributes.
// Attribute names are sorted when building the expected UpdateExpression.
func assertUpdateSets(t *testing.T, params *awsdynamodb.UpdateItemInput, want map[string]string) {
	t.Helper()

	if params.UpdateExpression == nil {
		t.Fatal("UpdateExpression is nil")
	}

	attrs := make([]string, 0, len(want))
	for attr := range want {
		attrs = append(attrs, attr)
	}
	slices.Sort(attrs)

	parts := make([]string, len(attrs))
	for i, attr := range attrs {
		parts[i] = fmt.Sprintf("#%s = :%s", attr, attr)
	}
	wantExpr := "SET " + strings.Join(parts, ", ")
	if got := *params.UpdateExpression; got != wantExpr {
		t.Fatalf("UpdateExpression = %q, want %q", got, wantExpr)
	}

	if len(params.ExpressionAttributeNames) != len(want) {
		t.Fatalf("ExpressionAttributeNames: got %d entries, want %d", len(params.ExpressionAttributeNames), len(want))
	}
	if len(params.ExpressionAttributeValues) != len(want) {
		t.Fatalf("ExpressionAttributeValues: got %d entries, want %d", len(params.ExpressionAttributeValues), len(want))
	}

	for attr, wantVal := range want {
		nameKey := "#" + attr
		if params.ExpressionAttributeNames[nameKey] != attr {
			t.Fatalf("ExpressionAttributeNames[%q] = %q, want %q", nameKey, params.ExpressionAttributeNames[nameKey], attr)
		}

		valKey := ":" + attr
		gotAV, ok := params.ExpressionAttributeValues[valKey].(*types.AttributeValueMemberS)
		if !ok {
			t.Fatalf("ExpressionAttributeValues[%q] is not AttributeValueMemberS", valKey)
		}
		if gotAV.Value != wantVal {
			t.Fatalf("ExpressionAttributeValues[%q] = %q, want %q", valKey, gotAV.Value, wantVal)
		}
	}
}

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
