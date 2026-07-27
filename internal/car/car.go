// Car entity and validation rules for create/update payloads.
package car

import "github.com/phides-code/go-multi-api/internal/domain"

type Car struct {
	ID        string `json:"id" dynamodbav:"id"`
	Model     string `json:"model" dynamodbav:"model"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateInput struct {
	Model string
}

type UpdateInput struct {
	ID    string
	Model string
}

func validateModel(model string) error {
	return domain.ValidateRequiredString(model, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
}

func ValidateCreateInput(input CreateInput) error {
	return validateModel(input.Model)
}

func ValidateUpdateInput(input UpdateInput) error {
	if err := domain.ValidateID(input.ID); err != nil {
		return err
	}
	return validateModel(input.Model)
}
