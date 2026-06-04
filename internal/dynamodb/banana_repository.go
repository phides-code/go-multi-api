// DynamoDB implementation of domain.BananaRepository for the bananas table.
package dynamodb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/domain"
)

const bananasTableName = "AppnameBananas"

const defaultListLimit int32 = 50

type BananaRepository struct {
	client *dynamodb.Client
}

func NewBananaRepository(client *dynamodb.Client) *BananaRepository {
	return &BananaRepository{client: client}
}

func (r *BananaRepository) Create(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	item, err := attributevalue.MarshalMap(banana)
	if err != nil {
		return domain.Banana{}, fmt.Errorf("marshal banana: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(bananasTableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var conditionalCheck *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheck) {
			return domain.Banana{}, fmt.Errorf("create banana: %w", err)
		}
		return domain.Banana{}, fmt.Errorf("put item: %w", err)
	}

	return banana, nil
}

func (r *BananaRepository) GetByID(ctx context.Context, id string) (domain.Banana, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(bananasTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return domain.Banana{}, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return domain.Banana{}, domain.ErrNotFound
	}

	var banana domain.Banana
	if err := attributevalue.UnmarshalMap(out.Item, &banana); err != nil {
		return domain.Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return banana, nil
}

func (r *BananaRepository) List(ctx context.Context, opts domain.ListOptions) (domain.Page, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	input := &dynamodb.ScanInput{
		TableName: aws.String(bananasTableName),
		Limit:     aws.Int32(limit),
	}

	if opts.Cursor != "" {
		startKey, err := decodeCursor(opts.Cursor)
		if err != nil {
			return domain.Page{}, fmt.Errorf("decode cursor: %w", err)
		}
		input.ExclusiveStartKey = startKey
	}

	out, err := r.client.Scan(ctx, input)
	if err != nil {
		return domain.Page{}, fmt.Errorf("scan items: %w", err)
	}

	items := make([]domain.Banana, 0, len(out.Items))
	for _, item := range out.Items {
		var banana domain.Banana
		if err := attributevalue.UnmarshalMap(item, &banana); err != nil {
			return domain.Page{}, fmt.Errorf("unmarshal banana: %w", err)
		}
		items = append(items, banana)
	}

	page := domain.Page{Items: items}
	if out.LastEvaluatedKey != nil {
		page.NextCursor, err = encodeCursor(out.LastEvaluatedKey)
		if err != nil {
			return domain.Page{}, fmt.Errorf("encode cursor: %w", err)
		}
	}

	return page, nil
}

func (r *BananaRepository) Update(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	out, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(bananasTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: banana.ID},
		},
		UpdateExpression:         aws.String("SET #content = :content"),
		ConditionExpression:      aws.String("attribute_exists(id)"),
		ExpressionAttributeNames: map[string]string{"#content": "content"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":content": &types.AttributeValueMemberS{Value: banana.Content},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		var conditionalCheck *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheck) {
			return domain.Banana{}, domain.ErrNotFound
		}
		return domain.Banana{}, fmt.Errorf("update item: %w", err)
	}

	var updated domain.Banana
	if err := attributevalue.UnmarshalMap(out.Attributes, &updated); err != nil {
		return domain.Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return updated, nil
}

func (r *BananaRepository) Delete(ctx context.Context, id string) (domain.Banana, error) {
	out, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(bananasTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ReturnValues: types.ReturnValueAllOld,
	})
	if err != nil {
		return domain.Banana{}, fmt.Errorf("delete item: %w", err)
	}
	if out.Attributes == nil {
		return domain.Banana{}, domain.ErrNotFound
	}

	var deleted domain.Banana
	if err := attributevalue.UnmarshalMap(out.Attributes, &deleted); err != nil {
		return domain.Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return deleted, nil
}

func encodeCursor(key map[string]types.AttributeValue) (string, error) {
	idVal, ok := key["id"].(*types.AttributeValueMemberS)
	if !ok {
		return "", fmt.Errorf("missing id in cursor")
	}

	raw, err := json.Marshal(map[string]string{"id": idVal.Value})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var parsed map[string]string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	id, ok := parsed["id"]
	if !ok {
		return nil, fmt.Errorf("missing id in cursor")
	}

	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}, nil
}
