package domain

const DefaultListLimit int32 = 50

type ListOptions struct {
	Limit  int32
	Cursor string
}
