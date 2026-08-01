// UUID rules for path `{id}` and generation of new resource IDs.
package domain

import "github.com/google/uuid"

func ValidateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}
	return nil
}

func NewID() string {
	return uuid.NewString()
}
