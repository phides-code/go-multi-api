// Banana entity and validation rules for create/update payloads.
package banana

import "github.com/phides-code/go-multi-api/internal/domain"

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Content   string `json:"content" dynamodbav:"content"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateInput struct {
	Content string
}

type UpdateInput struct {
	ID      string
	Content string
}

func validateContent(content string) error {
	return domain.ValidateRequiredString(content, domain.DefaultMinStringLength, domain.DefaultMaxStringLength)
}

func ValidateCreateInput(input CreateInput) error {
	return validateContent(input.Content)
}

func ValidateUpdateInput(input UpdateInput) error {
	if err := domain.ValidateID(input.ID); err != nil {
		return err
	}
	return validateContent(input.Content)
}
