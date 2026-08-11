package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

// 這個Payload也等同於Claim   這個套件的Valid完全由自己掌控??  不對  只有claim需要自己驗證  其餘簽名應由套件處理
type Payload[T UserIDConstraint] struct {
	ID        uuid.UUID `json:"id"`
	UPN       string    `json:"upn"`
	UserId    T         `json:"userid"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `josn:"expired_at"`
}

func NewPayload[T UserIDConstraint](upn string, userID T, duration time.Duration) (*Payload[T], error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload[T]{
		ID:        uuid,
		UPN:       upn,
		UserId:    userID,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// 需要實現jwt Claim的Valid街口  反正就是你的claim資料要自己寫驗證
// PasetoMaker.VertifyToken 會手動呼叫此方法做過期驗證  不可移除
func (payload *Payload[T]) Valid() error {
	if time.Now().UTC().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

// 以下實現 golang-jwt/jwt v5 的 Claims 介面  讓 Payload 可直接作為 jwt claims 使用
// v5 會依 GetExpirationTime() 自動比對並在過期時回傳 jwt.ErrTokenExpired
func (payload *Payload[T]) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.ExpiredAt), nil
}

func (payload *Payload[T]) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

func (payload *Payload[T]) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

func (payload *Payload[T]) GetIssuer() (string, error) {
	return "", nil
}

func (payload *Payload[T]) GetSubject() (string, error) {
	return payload.UPN, nil
}

func (payload *Payload[T]) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
