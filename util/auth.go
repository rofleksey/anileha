package util

import "golang.org/x/crypto/bcrypt"

func applySalt(password, salt string) []byte {
	passwordBytes := []byte(password)
	saltBytes := []byte(salt)
	concatBytes := make([]byte, 0, len(passwordBytes)+len(saltBytes))
	concatBytes = append(concatBytes, passwordBytes...)
	concatBytes = append(concatBytes, saltBytes...)
	return concatBytes
}

func HashPassword(password, salt string) (string, error) {
	passwordBytes := applySalt(password, salt)
	bytes, err := bcrypt.GenerateFromPassword(passwordBytes, 14)
	return string(bytes), err
}

func CheckPasswordHash(password, salt, hash string) bool {
	passwordBytes := applySalt(password, salt)
	err := bcrypt.CompareHashAndPassword([]byte(hash), passwordBytes)
	return err == nil
}
