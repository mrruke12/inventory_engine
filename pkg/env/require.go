package env

import (
	"fmt"
	"os"
	"strconv"
)

// Lookups for a key in the environment, returns an error if the key is not present.
func Require(key string) (string, error) {
	v, exists := os.LookupEnv(key)

	if !exists {
		return "", fmt.Errorf("key is not present in env: %s", key)
	}

	return v, nil
}

// Lookups for a key in the environment and casts it to int, returns an error if the key is not present.
func RequireInt(key string) (int, error) {
	v, err := Require(key)

	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(v)

	if err != nil {
		return 0, fmt.Errorf("key is present in env is not an int: %s", key)
	}

	return i, nil
}

// Lookups for multiple keys in the environment, returns an error if any key is not present.
func RequireKeys(keys []string) ([]string, error) {
	result := make([]string, len(keys))

	for i, key := range keys {
		v, err := Require(key)

		if err != nil {
			return nil, err
		}

		result[i] = v
	}

	return result, nil
}
