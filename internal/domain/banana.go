// Banana entity and validation rules for create/update payloads.
package domain

import (
	"strings"
	"unicode/utf8"
)

const (
	MinContentLength = 1
	MaxContentLength = 1000
)

type Banana struct {
	ID      string `json:"id" dynamodbav:"id"`
	Content string `json:"content" dynamodbav:"content"`
}

type CreateBananaInput struct {
	Content string
}

type UpdateBananaInput struct {
	ID      string
	Content string
}

func ValidateContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrInvalidContent
	}
	length := utf8.RuneCountInString(content)
	if length < MinContentLength || length > MaxContentLength {
		return ErrInvalidContent
	}
	return nil
}

func ValidateCreateInput(input CreateBananaInput) error {
	return ValidateContent(input.Content)
}

func ValidateUpdateInput(input UpdateBananaInput) error {
	if err := ValidateID(input.ID); err != nil {
		return err
	}
	return ValidateContent(input.Content)
}
