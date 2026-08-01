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

// DynamoDB attribute names (must stay aligned with Banana dynamodbav / JSON tags).
const (
	AttrID         = "id"
	AttrDescriptor = "descriptor"
	AttrRating     = "rating"
	AttrCreatedOn  = "createdOn"
)

// DynamoDB condition expressions for the id key (create vs update).
const (
	ConditionIDNotExists = "attribute_not_exists(" + AttrID + ")"
	ConditionIDExists    = "attribute_exists(" + AttrID + ")"
)

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

func idKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		AttrID: &types.AttributeValueMemberS{Value: id},
	}
}

func isConditionalCheckFailed(err error) bool {
	var conditionalCheck *types.ConditionalCheckFailedException
	return errors.As(err, &conditionalCheck)
}

func unmarshalBanana(item map[string]types.AttributeValue) (Banana, error) {
	var banana Banana
	if err := attributevalue.UnmarshalMap(item, &banana); err != nil {
		return Banana{}, fmt.Errorf("unmarshal banana: %w", err)
	}
	return banana, nil
}

func (r *dynamoRepository) Create(ctx context.Context, banana Banana) (Banana, error) {
	item, err := attributevalue.MarshalMap(banana)
	if err != nil {
		return Banana{}, fmt.Errorf("marshal banana: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String(ConditionIDNotExists),
	})

	if err != nil {
		if isConditionalCheckFailed(err) {
			return Banana{}, domain.ErrAlreadyExists
		}
		return Banana{}, fmt.Errorf("put item: %w", err)
	}

	return banana, nil
}

func (r *dynamoRepository) GetByID(ctx context.Context, id string) (Banana, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       idKey(id),
	})
	if err != nil {
		return Banana{}, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return Banana{}, domain.ErrNotFound
	}

	return unmarshalBanana(out.Item)
}

func (r *dynamoRepository) List(ctx context.Context) ([]Banana, error) {
	var items []Banana
	var startKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName: aws.String(TableName),
		}
		if startKey != nil {
			input.ExclusiveStartKey = startKey
		}

		out, err := r.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("scan items: %w", err)
		}

		for _, item := range out.Items {
			banana, err := unmarshalBanana(item)
			if err != nil {
				return nil, err
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
		TableName: aws.String(TableName),
		Key:       idKey(banana.ID),
		UpdateExpression: aws.String(fmt.Sprintf(
			"SET #%s = :%s, "+
				"#%s = :%s",
			AttrDescriptor, AttrDescriptor,
			AttrRating, AttrRating,
		)),
		ConditionExpression: aws.String(ConditionIDExists),
		ExpressionAttributeNames: map[string]string{
			"#" + AttrDescriptor: AttrDescriptor,
			"#" + AttrRating:     AttrRating,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":" + AttrDescriptor: &types.AttributeValueMemberS{Value: banana.Descriptor},
			":" + AttrRating:     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", banana.Rating)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		if isConditionalCheckFailed(err) {
			return Banana{}, domain.ErrNotFound
		}
		return Banana{}, fmt.Errorf("update item: %w", err)
	}

	return unmarshalBanana(out.Attributes)
}

func (r *dynamoRepository) Delete(ctx context.Context, id string) (Banana, error) {
	out, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:    aws.String(TableName),
		Key:          idKey(id),
		ReturnValues: types.ReturnValueAllOld,
	})
	if err != nil {
		return Banana{}, fmt.Errorf("delete item: %w", err)
	}
	if out.Attributes == nil {
		return Banana{}, domain.ErrNotFound
	}

	return unmarshalBanana(out.Attributes)
}
