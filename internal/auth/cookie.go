package auth

import (
	"encoding/hex"
	"net/http"
)

type CookieEncoder struct {
	crypto *Crypto
}

func NewCookieEncoder(secret string) (*CookieEncoder, error) {
	c, err := NewCrypto(secret)
	if err != nil {
		return nil, err
	}
	return &CookieEncoder{c}, nil
}

func (ce *CookieEncoder) EncodeUserID(userID string) (string, error) {
	enc, err := ce.crypto.Encrypt([]byte(userID))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(enc), nil
}

func (ce *CookieEncoder) DecodeUserID(c *http.Cookie) (string, error) {
	data, err := hex.DecodeString(c.Value)
	if err != nil {
		return "", err
	}

	plain, err := ce.crypto.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
