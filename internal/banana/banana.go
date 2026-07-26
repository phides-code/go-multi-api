// Banana entity and validation rules for create/update payloads.
package banana

import "github.com/phides-code/go-multi-api/internal/domain"

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Color     string `json:"color" dynamodbav:"color"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateInput struct {
	Color string
}

type UpdateInput struct {
	ID    string
	Color string
}

func validateColor(color string) error {
	return domain.ValidateRequiredString(color, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
}

func ValidateCreateInput(input CreateInput) error {
	return validateColor(input.Color)
}

func ValidateUpdateInput(input UpdateInput) error {
	if err := domain.ValidateID(input.ID); err != nil {
		return err
	}
	return validateColor(input.Color)
}
