package util

import (
	"crypto/rand"
	"math/big"
)

const cdAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const cdRandomLen = 17

func generateCode(prefix string) (string, error) {
	b := make([]byte, cdRandomLen)
	alphabetLen := big.NewInt(int64(len(cdAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		b[i] = cdAlphabet[n.Int64()]
	}
	return prefix + string(b), nil
}

// GenerateIdealCode generates a unique ideal code in the format "il_[17 alphanumeric chars]".
func GenerateIdealCode() (string, error) {
	return generateCode("il_")
}

// GenerateGoalCode generates a unique goal code in the format "gl_[17 alphanumeric chars]".
func GenerateGoalCode() (string, error) {
	return generateCode("gl_")
}

// GenerateTaskCode generates a unique task code in the format "tk_[17 alphanumeric chars]".
func GenerateTaskCode() (string, error) {
	return generateCode("tk_")
}
