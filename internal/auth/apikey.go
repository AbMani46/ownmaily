package auth

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func GenerateAPIKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return "om_" + hex.EncodeToString(b)
}

func KeyPrefix(key string) string {
	if len(key) < 10 {
		return key
	}
	return key[:10]
}

func HashKey(key string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(key), 10)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckKey(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
