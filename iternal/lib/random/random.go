package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func NewRandomString(length int) (string, error) {
	const op = "lib.random.NewRandomString"
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	result := make([]byte, length)
	limit := big.NewInt(int64(len(alphabet)))

	for i := range result {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		result[i] = alphabet[n.Int64()]
	}

	return string(result), nil
}
