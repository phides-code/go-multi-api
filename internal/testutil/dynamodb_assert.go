// Shared DynamoDB repository test helpers for any resource.
package testutil

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AssertUpdateSets checks that UpdateItem SETs exactly the given attributes.
// Attribute names are sorted when building the expected UpdateExpression.
// Values may be string (AttributeValueMemberS) or int (AttributeValueMemberN).
func AssertUpdateSets(t *testing.T, params *awsdynamodb.UpdateItemInput, want map[string]any) {
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
		switch v := wantVal.(type) {
		case string:
			gotAV, ok := params.ExpressionAttributeValues[valKey].(*types.AttributeValueMemberS)
			if !ok {
				t.Fatalf("ExpressionAttributeValues[%q] is not AttributeValueMemberS", valKey)
			}
			if gotAV.Value != v {
				t.Fatalf("ExpressionAttributeValues[%q] = %q, want %q", valKey, gotAV.Value, v)
			}
		case int:
			gotAV, ok := params.ExpressionAttributeValues[valKey].(*types.AttributeValueMemberN)
			if !ok {
				t.Fatalf("ExpressionAttributeValues[%q] is not AttributeValueMemberN", valKey)
			}
			wantN := fmt.Sprintf("%d", v)
			if gotAV.Value != wantN {
				t.Fatalf("ExpressionAttributeValues[%q] = %q, want %q", valKey, gotAV.Value, wantN)
			}
		default:
			t.Fatalf("unsupported want type for %q: %T", attr, wantVal)
		}
	}
}
