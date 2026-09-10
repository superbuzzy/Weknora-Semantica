package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// StableID creates deterministic machine identifiers without depending on
// English transliteration. Business-facing Chinese labels are stored separately.
func StableID(prefix, domain, label string) string {
	raw := strings.TrimSpace(domain) + "\x00" + strings.TrimSpace(label)
	sum := sha256.Sum256([]byte(raw))
	return prefix + "_" + hex.EncodeToString(sum[:6])
}
