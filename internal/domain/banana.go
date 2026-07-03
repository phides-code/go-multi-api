// Banana entity and validation rules for create/update payloads.
package domain

const (
	MinStringLength = 1
	MaxStringLength = 1000
)

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Content   string `json:"content" dynamodbav:"content"`
	Variety   string `json:"variety" dynamodbav:"variety"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateBananaInput struct {
	Content string
	Variety string
}

type UpdateBananaInput struct {
	ID      string
	Content string
	Variety string
}

func validateBananaInput(content, variety string) error {
	if err := ValidateRequiredString(content, MinStringLength, MaxStringLength); err != nil {
		return err
	}
	return ValidateRequiredString(variety, MinStringLength, MaxStringLength)
}

func ValidateCreateInput(input CreateBananaInput) error {
	return validateBananaInput(input.Content, input.Variety)
}

func ValidateUpdateInput(input UpdateBananaInput) error {
	if err := ValidateID(input.ID); err != nil {
		return err
	}
	return validateBananaInput(input.Content, input.Variety)
}
