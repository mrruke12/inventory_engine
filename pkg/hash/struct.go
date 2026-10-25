package hash

import (
	"crypto/sha256"
	"encoding/json"
)

func HashStruct(v any) ([32]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(b), nil
}
