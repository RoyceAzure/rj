package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretSize = 32

type JWTMaker[T UserIDConstraint] struct {
	secretKey string
}

// 驗證field合法
func NewJWTMaker[T UserIDConstraint](secret string) (Maker[T], error) {
	if len(secret) < minSecretSize {
		return nil, fmt.Errorf("invalid ket size : must be at least %d charcters", minSecretSize)
	}
	return &JWTMaker[T]{secret}, nil
}

// 準備好自己的payload
// 使用jwt.NewWithClaims 產生claim, 要指定加密演算法
// 使用secret加密  要自己給secret
func (maker *JWTMaker[T]) CreateToken(username string, userID T, duration time.Duration) (string, *Payload[T], error) {
	payload, err := NewPayload(username, userID, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

func (maker *JWTMaker[T]) VertifyToken(token string) (*Payload[T], error) {
	//需要一個自訂keyFunc  提供加密演算法  也可以用來驗證提供的token所使用的演算法合不合法
	//根據這個設計  使用者可以根據token內容來決定要使用何種key  通常應該是根據header??
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	//ParseWithClaims會使用我們提供的keyfunc  並根據claim的GetExpirationTime()等方法自動驗證
	//v5會回傳包裝過的sentinel error  用errors.Is判斷實際錯誤種類
	jwtToken, err := jwt.ParseWithClaims(token, &Payload[T]{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload[T])
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
