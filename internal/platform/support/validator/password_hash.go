package validator

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordInvalid = errors.New("password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character")

func HashPassword(password string) (string, error) {
	for _, char := range password {
		if char < 32 || char > 126 {
			return "", ErrPasswordInvalid
		}
	}
	var hasUppercase, hasLowercase, hasDigit, hasSpecialChar bool
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUppercase = true
		case char >= 'a' && char <= 'z':
			hasLowercase = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case (char >= 33 && char <= 47) || (char >= 58 && char <= 64) || (char >= 91 && char <= 96) || (char >= 123 && char <= 126):
			hasSpecialChar = true
		}
	}

	if len(password) < 8 || len(password) > 64 || !hasUppercase || !hasLowercase || !hasDigit || !hasSpecialChar {
		return "", ErrPasswordInvalid
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
