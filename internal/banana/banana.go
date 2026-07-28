// Banana entity and validation rules for create/update payloads.
package banana

import "github.com/phides-code/go-multi-api/internal/domain"

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Variety     string `json:"variety" dynamodbav:"variety"`
	Rating    int    `json:"rating" dynamodbav:"rating"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateInput struct {
	Variety  string
	Rating int
}

type UpdateInput struct {
	ID     string
	Variety  string
	Rating int
}

func validateVariety(variety string) error {
	return domain.ValidateRequiredString(variety, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
}

func validateRating(rating int) error {
	return domain.ValidateRequiredInt(rating, domain.DefaultMinInt, domain.DefaultMaxInt)
}

func ValidateCreateInput(input CreateInput) error {
	if err := validateVariety(input.Variety); err != nil {
		return err
	}
	return validateRating(input.Rating)
}

func ValidateUpdateInput(input UpdateInput) error {
	if err := domain.ValidateID(input.ID); err != nil {
		return err
	}
	if err := validateVariety(input.Variety); err != nil {
		return err
	}
	return validateRating(input.Rating)
}
