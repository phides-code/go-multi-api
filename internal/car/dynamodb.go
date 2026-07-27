// DynamoDB implementation of Repository for the cars table.
package car

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

const tableName = "AppnameCars"

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

func (r *dynamoRepository) Create(ctx context.Context, c Car) (Car, error) {
	item, err := attributevalue.MarshalMap(c)
	if err != nil {
		return Car{}, fmt.Errorf("marshal car: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})

	if err != nil {
		var conditionalCheck *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheck) {
			return Car{}, domain.ErrAlreadyExists
		}
		return Car{}, fmt.Errorf("put item: %w", err)
	}

	return c, nil
}

func (r *dynamoRepository) GetByID(ctx context.Context, id string) (Car, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return Car{}, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return Car{}, domain.ErrNotFound
	}

	var c Car
	if err := attributevalue.UnmarshalMap(out.Item, &c); err != nil {
		return Car{}, fmt.Errorf("unmarshal car: %w", err)
	}

	return c, nil
}

func (r *dynamoRepository) List(ctx context.Context) ([]Car, error) {
	var items []Car
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
			var c Car
			if err := attributevalue.UnmarshalMap(item, &c); err != nil {
				return nil, fmt.Errorf("unmarshal car: %w", err)
			}
			items = append(items, c)
		}

		if out.LastEvaluatedKey == nil {
			break
		}
		startKey = out.LastEvaluatedKey
	}

	return items, nil
}

func (r *dynamoRepository) Update(ctx context.Context, c Car) (Car, error) {
	out, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: c.ID},
		},
		UpdateExpression:         aws.String("SET #model = :model"),
		ConditionExpression:      aws.String("attribute_exists(id)"),
		ExpressionAttributeNames: map[string]string{"#model": "model"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":model": &types.AttributeValueMemberS{Value: c.Model},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		var conditionalCheck *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheck) {
			return Car{}, domain.ErrNotFound
		}
		return Car{}, fmt.Errorf("update item: %w", err)
	}

	var updated Car
	if err := attributevalue.UnmarshalMap(out.Attributes, &updated); err != nil {
		return Car{}, fmt.Errorf("unmarshal car: %w", err)
	}

	return updated, nil
}

func (r *dynamoRepository) Delete(ctx context.Context, id string) (Car, error) {
	out, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ReturnValues: types.ReturnValueAllOld,
	})
	if err != nil {
		return Car{}, fmt.Errorf("delete item: %w", err)
	}
	if out.Attributes == nil {
		return Car{}, domain.ErrNotFound
	}

	var deleted Car
	if err := attributevalue.UnmarshalMap(out.Attributes, &deleted); err != nil {
		return Car{}, fmt.Errorf("unmarshal car: %w", err)
	}

	return deleted, nil
}
