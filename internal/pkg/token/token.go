package token

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math"
)

// New ...
func New() string {
	return randomBase64String(60)
}

// NewHex ...
func NewHex() string {
	return randomBase16String(60)
}

func randomBase64String(l int) string {
	buff := make([]byte, int(math.Round(float64(l)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return str[:l] // Strip 1 extra character we get from odd length results.
}

func randomBase16String(l int) string {
	buff := make([]byte, int(math.Round(float64(l)/2)))
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l] // Strip 1 extra character we get from odd length results.
}
