package pg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Overrides is a set of mock results for queries
type Overrides map[string]OverrideResult

// GetOverride retrieves an override for a query with args
func (os Overrides) GetOverride(key OverrideKey) *OverrideResult {
	if o, ok := os[key.Hash()]; ok {
		return &o
	}
	logger.Panic().Msgf("failed to get override for %v", key)
	return nil
}

// OverrideKey is a query + args to be used as key for searching the result
type OverrideKey struct {
	Query string
	Args  []any
}

// Hash will generate hash for this set of query/args
func (ok OverrideKey) Hash() string {
	data, err := json.Marshal(ok)
	// This should never happen, so there is no unittest for it
	if err != nil {
		logger.Panic().Msgf("failed to marshal %v: %v", ok, err)
	}
	hasher := sha256.New()
	hasher.Write(data)

	// 3. Converteer de hash naar een hexadecimale string.
	return hex.EncodeToString(hasher.Sum(nil))
}

// OverrideResult is a mock result for a specific query
type OverrideResult struct {
	affected int64
	err      error
	rows     Result
}
