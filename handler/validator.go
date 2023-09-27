package handler

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Map to translate validation error to human readble message
var validationErrorMessages = map[string]string{
	"required":   "{field} is required.",
	"min":        "{field} must be at least {param} characters long.",
	"max":        "{field} must not exceed {param} characters.",
	"startswith": "{field} must start with '{param}'.",
	"password":   "{field} must meet password criteria. Minimum 6 characters, maximum 64 characters, containing at least 1 capital characters AND 1 number AND 1 special (non-alpha-numeric) characters.",
}

type UserRegistrationValidator struct {
	Validator *validator.Validate
}

func (v *UserRegistrationValidator) Validate(i interface{}) error {
	return v.Validator.Struct(i)
}

// ValidatePassword is a custom validator to validate password based on our rules
func ValidatePassword(f1 validator.FieldLevel) bool {
	password := f1.Field().String()

	// Length checker
	if len(password) < 6 || len(password) > 64 {
		return false
	}

	// Capital letter checker
	var hasCapital bool
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasCapital = true
			break
		}
	}

	if !hasCapital {
		return false
	}

	// Numeric checker
	var hasNumber bool
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return false
	}

	// Non-Alphanumeric checker
	specialCharacterRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !specialCharacterRegex.MatchString(password) {
		return false
	}

	return true
}

// TranslateErrorMessages returns list of human-readable error messages
func TranslateErrorMessages(errs []validator.FieldError) []string {
	var messages []string

	for _, err := range errs {
		field, param, tagName := err.Field(), err.Param(), err.Tag()
		message := validationErrorMessages[tagName]

		message = strings.ReplaceAll(message, "{field}", field)
		message = strings.ReplaceAll(message, "{param}", param)

		messages = append(messages, message)
	}

	return messages
}
