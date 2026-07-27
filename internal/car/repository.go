// Repository interface; DynamoDB implementation is in dynamodb.go.
package car

import "context"

type Repository interface {
	Create(ctx context.Context, c Car) (Car, error)
	GetByID(ctx context.Context, id string) (Car, error)
	List(ctx context.Context) ([]Car, error)
	Update(ctx context.Context, c Car) (Car, error)
	Delete(ctx context.Context, id string) (Car, error)
}
