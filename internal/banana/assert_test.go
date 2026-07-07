// Banana-specific test helpers for HTTP wire shape and DynamoDB repository assertions.
package banana_test

import (
	"encoding/json"
	"errors"
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func decodeBananaData(t *testing.T, envelope platform.APIResponse) banana.Banana {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %s", *envelope.Error)
	}
	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var b banana.Banana
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("unmarshal banana: %v", err)
	}
	return b
}

func decodeBananaPageData(t *testing.T, envelope platform.APIResponse) banana.Page {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %s", *envelope.Error)
	}
	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var page banana.Page
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatalf("unmarshal page: %v", err)
	}
	return page
}

func assertBananaDataKeys(t *testing.T, envelope platform.APIResponse) {
	t.Helper()

	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("unmarshal data keys: %v", err)
	}

	want := []string{"content", "createdOn", "id"}
	if len(keys) != len(want) {
		t.Fatalf("data has %d keys %v, want exactly %v", len(keys), maps.Keys(keys), want)
	}
	for _, k := range want {
		if _, ok := keys[k]; !ok {
			t.Fatalf("missing data key %q; got %v", k, maps.Keys(keys))
		}
	}
}

func assertBananaPutItem(t *testing.T, params *awsdynamodb.PutItemInput, want banana.Banana) {
	t.Helper()

	if params.ConditionExpression == nil || *params.ConditionExpression != "attribute_not_exists(id)" {
		t.Fatalf("ConditionExpression = %v, want attribute_not_exists(id)", params.ConditionExpression)
	}

	var got banana.Banana
	if err := attributevalue.UnmarshalMap(params.Item, &got); err != nil {
		t.Fatalf("unmarshal item: %v", err)
	}
	if got != want {
		t.Fatalf("Item = %+v, want %+v", got, want)
	}
}

func assertBananaRepoResult(t *testing.T, op string, got banana.Banana, err error, want banana.Banana, wantErr error) {
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
