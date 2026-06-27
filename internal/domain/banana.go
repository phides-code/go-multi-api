// Banana entity and validation rules for create/update payloads.
package domain

const (
	MinContentLength = 1
	MaxContentLength = 1000
)

type Banana struct {
	ID        string `json:"id" dynamodbav:"id"`
	Content   string `json:"content" dynamodbav:"content"`
	CreatedOn uint64 `json:"createdOn" dynamodbav:"createdOn"`
}

type CreateBananaInput struct {
	Content string
}

type UpdateBananaInput struct {
	ID      string
	Content string
}

func ValidateCreateInput(input CreateBananaInput) error {
	return ValidateRequiredString(input.Content, MinContentLength, MaxContentLength)
}
func ValidateUpdateInput(input UpdateBananaInput) error {
	if err := ValidateID(input.ID); err != nil {
		return err
	}
	return ValidateRequiredString(input.Content, MinContentLength, MaxContentLength)
}
