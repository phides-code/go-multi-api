// DynamoDB implementation of Repository for the bananas table.
package banana

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phides-code/go-multi-api/internal/domain"
)

const tableName = "AppnameBananas"

type dynamoRepository struct {
	client dynamoAPI
}

type dynamoAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

func NewRepository(client dynamoAPI) Repository {
	return &dynamoRepository{client: client}
}

func (r *dynamoRepository) Create(ctx context.Context, banana Banana) (Banana, error) {
	item, err := attributevalue.MarshalMap(banana)
	if err != nil {
		return Banana{}, fmt.Errorf("marshal banana: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})

	if err != nil {
		var conditionalCheck *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheck) {
			return Banana{}, domain.ErrAlreadyExists
		}
		return Banana{}, fmt.Errorf("put item: %w", err)
	}

	return banana, nil
}

func (r *dynamoRepository) GetByID(ctx context.Context, id string) (Banana, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return Banana{}, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return Banana{}, domain.ErrNotFound
	}

	var banana Banana
	if err := attributevalue.UnmarshalMap(out.Item, &banana); err != nil {
		return Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return banana, nil
}

func (r *dynamoRepository) List(ctx context.Context) ([]Banana, error) {
	var items []Banana
	var startKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}
		if startKey != nil {
			input.ExclusiveStartKey = startKey
		}

		out, err := r.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("scan items: %w", err)
		}

		for _, item := range out.Items {
			var banana Banana
			if err := attributevalue.UnmarshalMap(item, &banana); err != nil {
				return nil, fmt.Errorf("unmarshal banana: %w", err)
			}
			items = append(items, banana)
		}

		if out.LastEvaluatedKey == nil {
			break
		}
		startKey = out.LastEvaluatedKey
	}

	return items, nil
}

func (r *dynamoRepository) Update(ctx context.Context, banana Banana) (Banana, error) {
	out, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
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
			return Banana{}, domain.ErrNotFound
		}
		return Banana{}, fmt.Errorf("update item: %w", err)
	}

	var updated Banana
	if err := attributevalue.UnmarshalMap(out.Attributes, &updated); err != nil {
		return Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return updated, nil
}

func (r *dynamoRepository) Delete(ctx context.Context, id string) (Banana, error) {
	out, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ReturnValues: types.ReturnValueAllOld,
	})
	if err != nil {
		return Banana{}, fmt.Errorf("delete item: %w", err)
	}
	if out.Attributes == nil {
		return Banana{}, domain.ErrNotFound
	}

	var deleted Banana
	if err := attributevalue.UnmarshalMap(out.Attributes, &deleted); err != nil {
		return Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}

	return deleted, nil
}

