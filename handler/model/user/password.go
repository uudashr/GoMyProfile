package user

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

func genSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func hashPassword(password string, salt []byte) []byte {
	iterations := 10000
	keyLength := 32

	return pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)
}

func verifyPassword(password string, salt []byte, hashedPassword []byte) bool {
	return bytes.Equal(hashPassword(password, salt), hashedPassword)
}
