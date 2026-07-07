// Cursor-based list options shared by all paginated resources.
package domain

const DefaultListLimit int32 = 50

type ListOptions struct {
	Limit  int32
	Cursor string
}
