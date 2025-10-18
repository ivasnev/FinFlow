package key_pair

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// ExtractKeyPair преобразует модель пары ключей базы данных в обычную модель
func ExtractKeyPair(dbKeyPair *KeyPair) *models.KeyPair {
	if dbKeyPair == nil {
		return nil
	}

	return &models.KeyPair{
		ID:         dbKeyPair.ID,
		PublicKey:  dbKeyPair.PublicKey,
		PrivateKey: dbKeyPair.PrivateKey,
		IsActive:   dbKeyPair.IsActive,
		CreatedAt:  dbKeyPair.CreatedAt,
		UpdatedAt:  dbKeyPair.UpdatedAt,
	}
}

// loadKeyPair преобразует обычную модель пары ключей в модель базы данных
func loadKeyPair(keyPair *models.KeyPair) *KeyPair {
	if keyPair == nil {
		return nil
	}

	return &KeyPair{
		ID:         keyPair.ID,
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
		IsActive:   keyPair.IsActive,
		CreatedAt:  keyPair.CreatedAt,
		UpdatedAt:  keyPair.UpdatedAt,
	}
}
