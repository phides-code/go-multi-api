// UUID validation for path parameters and UUID v4 generation for new resources.
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
