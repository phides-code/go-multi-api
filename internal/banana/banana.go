// Banana entity and validation rules for create/update payloads.
package banana

import "github.com/phides-code/go-multi-api/internal/domain"

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Color     string `json:"color" dynamodbav:"color"`
	Rating    int    `json:"rating" dynamodbav:"rating"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateInput struct {
	Color  string
	Rating int
}

type UpdateInput struct {
	ID     string
	Color  string
	Rating int
}

func validateColor(color string) error {
	return domain.ValidateRequiredString(color, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
}

func validateRating(rating int) error {
	return domain.ValidateRequiredInt(rating, domain.DefaultMinInt, domain.DefaultMaxInt)
}

func ValidateCreateInput(input CreateInput) error {
	if err := validateColor(input.Color); err != nil {
		return err
	}
	return validateRating(input.Rating)
}

func ValidateUpdateInput(input UpdateInput) error {
	if err := domain.ValidateID(input.ID); err != nil {
		return err
	}
	if err := validateColor(input.Color); err != nil {
		return err
	}
	return validateRating(input.Rating)
}
