package utils

import (
	"math/rand"
	"os"
)

// GetEnv returns the value of the environment variable named by the key.
// If the variable is not set, it returns the fallback value.
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// ToUpper returns a new string with the first character of s converted to its uppercase.
// If the length of s is zero, it returns s.
func ToUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	return string([]rune(s)[0]-32) + s[1:]
}

// GenerateRandomString generates a random string consisting of uppercase letters and digits of length n.
var GenerateRandomString = func(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
