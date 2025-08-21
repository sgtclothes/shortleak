package utils

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	ValidateUrlDirect(url string) error
	ValidateUrlFormatDirect(url string) error
	ValidateStruct(req interface{}) error
	GenerateRandomString(n int) string
}

type DefaultValidator struct{}

func (DefaultValidator) ValidateUrlDirect(url string) error {
	return ValidateUrlDirect(url)
}

func (DefaultValidator) ValidateUrlFormatDirect(url string) error {
	return ValidateUrlFormatDirect(url)
}

func (DefaultValidator) ValidateStruct(req interface{}) error {
	if errs := ValidateStruct(req); errs != nil {
		return fmt.Errorf("validation failed")
	}
	return nil
}

func (DefaultValidator) GenerateRandomString(n int) string {
	return GenerateRandomString(n)
}

var validate = validator.New()

func init() {
	/** Register custom validation for password */
	validate.RegisterValidation("password", validatePassword)
	/** Register custom validation for URL */
	validate.RegisterValidation("url", validateUrl)
}

/** ValidateStruct validates the struct and returns a map of validation errors */
func ValidateStruct(data interface{}) map[string]string {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}
	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		errors[e.Field()] = msgForTag(e.Tag(), e.Param())
	}
	return errors
}

/** msgForTag returns a user-friendly message for each validation tag */
func msgForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "field must be at least " + param + " characters long"
	case "max":
		return "field must be at most " + param + " characters long"
	case "password":
		return "password must be at least 8 characters, contain 1 uppercase, 1 lowercase, 1 number, and 1 special character"
	case "url":
		return "invalid URL format"
	}
	return "invalid field"
}

/** validatePassword checks if the password meets the required criteria */
func validatePassword(fl validator.FieldLevel) bool {
	/** get the password from the field */
	password := fl.Field().String()
	/** check if password is at least 8 characters */
	if len(password) < 8 {
		return false
	}
	/** regex: at least 1 lowercase, 1 uppercase, 1 number, 1 special char */
	var (
		lower   = regexp.MustCompile(`[a-z]`)
		upper   = regexp.MustCompile(`[A-Z]`)
		number  = regexp.MustCompile(`[0-9]`)
		special = regexp.MustCompile(`[^a-zA-Z0-9]`)
	)
	/** check if all conditions are met */
	return lower.MatchString(password) &&
		upper.MatchString(password) &&
		number.MatchString(password) &&
		special.MatchString(password)
}

func validateUrl(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	/** Simple Validation URL */
	var urlRegex = regexp.MustCompile(`^(http|https)://[a-zA-Z0-9.-]+(:[0-9]+)?(/.*)?$`)
	return urlRegex.MatchString(url)
}

func ValidateUrlFormatDirect(url string) error {
	if err := validate.Var(url, "url"); err != nil {
		return err
	}
	return nil
}

func ValidateUrlDirect(url string) error {
	if err := validate.Var(url, "required"); err != nil {
		return err
	}
	return nil
}
