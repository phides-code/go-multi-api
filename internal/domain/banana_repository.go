// BananaRepository interface and pagination types; persistence is implemented elsewhere.
package domain

import "context"

const DefaultListLimit int32 = 50

type Page struct {
	Items      []Banana `json:"items"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

type ListOptions struct {
	Limit  int32
	Cursor string
}

type BananaRepository interface {
	Create(ctx context.Context, banana Banana) (Banana, error)
	GetByID(ctx context.Context, id string) (Banana, error)
	List(ctx context.Context, opts ListOptions) (Page, error)
	Update(ctx context.Context, banana Banana) (Banana, error)
	Delete(ctx context.Context, id string) (Banana, error)
}
