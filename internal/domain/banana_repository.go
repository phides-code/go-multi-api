// BananaRepository interface and pagination types; persistence is implemented elsewhere.
package domain

import "context"

type BananaPage struct {
	Items      []Banana `json:"items"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

type BananaRepository interface {
	Create(ctx context.Context, banana Banana) (Banana, error)
	GetByID(ctx context.Context, id string) (Banana, error)
	List(ctx context.Context, opts ListOptions) (BananaPage, error)
	Update(ctx context.Context, banana Banana) (Banana, error)
	Delete(ctx context.Context, id string) (Banana, error)
}
