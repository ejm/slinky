package internal

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// "Zrockford16" alphabet inspired by [Human Sharable Codes](https://kalifi.org/2019/09/human-shareable-codes.html)
// This is an `oldnew` representation for `strings.Replacer`, so
// the traditional hex digit goes first, then our modified one
var ALPHABET = []string{
	"0", "b", "1", "n", "2", "d", "3", "r",
	"4", "f", "5", "g", "6", "k", "7", "m",
	"8", "c", "9", "p", "a", "q", "b", "x",
	"c", "t", "d", "w", "e", "z", "f", "h",
}

var REPLACER = strings.NewReplacer(ALPHABET...)

// Encode a byte array with our custom alphabet
func encode(bytes []byte) string {
	return REPLACER.Replace(hex.EncodeToString(bytes))
}

// Generate a string of `size` characters
func GenerateShortUrl(size int) (string, error) {
	// One byte is encoded as two hex digits, so we only need to allocate
	// (ceil) half the length of our desired string
	bytes := make([]byte, (size+1)/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	replaced := encode(bytes)
	// For odd values of `size`, we need to exclude the extra character
	return replaced[:size], nil
}
