// Repository interface and list page type; DynamoDB implementation is in dynamodb.go.
package banana

import (
	"context"

	"github.com/phides-code/go-multi-api/internal/domain"
)

type Page struct {
	Items      []Banana `json:"items"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, banana Banana) (Banana, error)
	GetByID(ctx context.Context, id string) (Banana, error)
	List(ctx context.Context, opts domain.ListOptions) (Page, error)
	Update(ctx context.Context, banana Banana) (Banana, error)
	Delete(ctx context.Context, id string) (Banana, error)
}
