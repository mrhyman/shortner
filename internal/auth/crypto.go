package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

type Crypto struct {
	aesgcm cipher.AEAD
}

func NewCrypto(secret string) (*Crypto, error) {
	key := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Crypto{aesgcm}, nil
}

func (c *Crypto) Encrypt(plain []byte) ([]byte, error) {
	nonce := make([]byte, c.aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return append(nonce, c.aesgcm.Seal(nil, nonce, plain, nil)...), nil
}

func (c *Crypto) Decrypt(ciphertext []byte) ([]byte, error) {
	ns := c.aesgcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:ns]
	body := ciphertext[ns:]

	return c.aesgcm.Open(nil, nonce, body, nil)
}
