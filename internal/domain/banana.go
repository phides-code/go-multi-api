// Banana entity and validation rules for create/update payloads.
package domain

const (
	BananaMinContentLength = 1
	BananaMaxContentLength = 1000
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

func validateBananaInput(content string) error {
	if err := ValidateRequiredString(content, BananaMinContentLength, BananaMaxContentLength); err != nil {
		return err
	}
	return nil
}

func ValidateCreateInput(input CreateBananaInput) error {
	return validateBananaInput(input.Content)
}

func ValidateUpdateInput(input UpdateBananaInput) error {
	if err := ValidateID(input.ID); err != nil {
		return err
	}
	return validateBananaInput(input.Content)
}
