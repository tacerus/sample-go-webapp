package web

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
)

func randString(nByte int, urlEncode bool) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	if urlEncode {
		return base64.RawURLEncoding.EncodeToString(b), nil
	}

	return hex.EncodeToString(b), nil
}
