package auth

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,32}$`)
	hasUpper      = regexp.MustCompile(`[A-Z]`)
	hasLower      = regexp.MustCompile(`[a-z]`)
	hasDigit      = regexp.MustCompile(`[0-9]`)
)

func ValidateCredentials(username, email, password string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username deve ter 3-32 caracteres e usar apenas letras, numeros, ponto, underscore ou hifen")
	}
	if len(email) < 5 {
		return fmt.Errorf("email invalido")
	}
	if len(password) < 8 || !hasUpper.MatchString(password) || !hasLower.MatchString(password) || !hasDigit.MatchString(password) {
		return fmt.Errorf("senha deve ter ao menos 8 caracteres, com maiuscula, minuscula e numero")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	data, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(data), err
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
