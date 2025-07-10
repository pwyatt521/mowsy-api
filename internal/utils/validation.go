package utils

import (
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}$`)
)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}

func IsValidZipCode(zipCode string) bool {
	return len(zipCode) == 5 && isNumeric(zipCode)
}

func isNumeric(str string) bool {
	for _, char := range str {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func SanitizeString(str string) string {
	return strings.TrimSpace(str)
}

func IsValidRating(rating int) bool {
	return rating >= 1 && rating <= 5
}