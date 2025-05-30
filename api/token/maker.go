package token

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

type UserIDConstraint interface {
	~int64 | ~string | uuid.UUID
}

type Maker[T UserIDConstraint] interface {
	CreateToken(upn string, userID T, duration time.Duration) (string, *Payload[T], error)
	VertifyToken(token string) (*Payload[T], error)
}

func GenerateRandomSymmetricKey() (string, error) {
	key := make([]byte, 24)

	// 使用加密安全的隨機數生成器
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	return encodedKey, nil
}
