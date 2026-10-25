package ports

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrruke12/inventory_engine/pkg/hash"
)

// Idempotency key carries information about operation
// and its result to replay it without reassembling.
type IdempotencyKey struct {
	Key         string
	Scope       string
	RequestHash [32]byte
	Result      json.RawMessage
	CreatedAt   time.Time
}

func NewIdempotencyKey(key, scope string, request any, result any) (*IdempotencyKey, error) {
	hash, err := hash.HashStruct(request)

	if err != nil {
		return nil, err
	}

	resultJson, err := json.Marshal(result)

	if err != nil {
		return nil, err
	}

	return &IdempotencyKey{
		Key:         key,
		Scope:       scope,
		RequestHash: hash,
		Result:      resultJson,
	}, nil
}

type CachedResult struct {
	Found       bool
	RequestHash [32]byte
	Result      json.RawMessage
}

func (r *CachedResult) Unmarshal(dest any) error {
	return json.Unmarshal(r.Result, dest)
}

type Idempotency interface {
	// Returns the cached result if the key exists, or creates idempotency key.
	GetOrCreate(ctx context.Context, key *IdempotencyKey) (*CachedResult, error)
}
