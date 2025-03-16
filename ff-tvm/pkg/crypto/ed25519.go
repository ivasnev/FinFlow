package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// GenerateEd25519KeyPair генерирует новую пару ключей Ed25519
func GenerateEd25519KeyPair() (publicKey string, privateKey string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(pub),
		base64.StdEncoding.EncodeToString(priv),
		nil
}

// ParseEd25519PrivateKey парсит приватный ключ из base64 строки
func ParseEd25519PrivateKey(privateKeyStr string) (ed25519.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}

	return ed25519.PrivateKey(privateKeyBytes), nil
}

// ParseEd25519PublicKey парсит публичный ключ из base64 строки
func ParseEd25519PublicKey(publicKeyStr string) (ed25519.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	return ed25519.PublicKey(publicKeyBytes), nil
}

// SignEd25519 подписывает данные с помощью Ed25519
func SignEd25519(privateKey ed25519.PrivateKey, data []byte) (string, error) {
	signature := ed25519.Sign(privateKey, data)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyEd25519 проверяет подпись Ed25519
func VerifyEd25519(publicKey ed25519.PublicKey, data []byte, signatureStr string) error {
	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return err
	}

	if !ed25519.Verify(publicKey, data, signature) {
		return errors.New("invalid signature")
	}

	return nil
}
