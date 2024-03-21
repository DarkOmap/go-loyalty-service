package hasher

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/models"
)

func GetHashedUser(u models.User) (*models.User, error) {
	b := make([]byte, 75)

	_, err := rand.Read(b)

	if err != nil {
		return nil, fmt.Errorf("generate random value: %w", err)
	}

	u.Salt = hex.EncodeToString(b)
	hash := sha256.New()
	data := append(b, u.Password...)
	hash.Write(data)
	dst := hash.Sum(nil)

	u.Password = hex.EncodeToString(dst)
	return &u, nil
}

func GetHashPassword(password, salt string) (string, error) {
	decodeSalt, err := hex.DecodeString(salt)

	if err != nil {
		return "", fmt.Errorf("decode salt: %w", err)
	}

	hash := sha256.New()
	data := append(decodeSalt, password...)
	hash.Write(data)
	dst := hash.Sum(nil)

	return hex.EncodeToString(dst), nil
}
