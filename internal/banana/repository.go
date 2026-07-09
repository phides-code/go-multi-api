// Repository interface; DynamoDB implementation is in dynamodb.go.
package banana

import "context"

type Repository interface {
	Create(ctx context.Context, banana Banana) (Banana, error)
	GetByID(ctx context.Context, id string) (Banana, error)
	List(ctx context.Context) ([]Banana, error)
	Update(ctx context.Context, banana Banana) (Banana, error)
	Delete(ctx context.Context, id string) (Banana, error)
}
