package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretSize = 32

type JWTMaker struct {
	secretKey string
}


// 驗證field合法
func NewJWTMaker(secret string) (Maker, error) {
	if len(secret) < minSecretSize {
		return nil, fmt.Errorf("invalid ket size : must be at least %d charcters", minSecretSize)
	}
	return &JWTMaker{secret}, nil
}

// 準備好自己的payload
// 使用jwt.NewWithClaims 產生claim, 要指定加密演算法
// 使用secret加密  要自己給secret
func (maker *JWTMaker) CreateToken(username string, userID int64, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, userID, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

func (maker *JWTMaker) VertifyToken(token string) (*Payload, error) {
	//需要一個自訂keyFunc  提供加密演算法  也可以用來驗證提供的token所使用的演算法合不合法
	//根據這個設計  使用者可以根據token內容來決定要使用何種key  通常應該是根據header??
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	//ParseWithClaims會使用我們提供的keyfunc與claim Valid()做驗證  且是回傳我們自訂的錯誤
	//且他返回的錯誤訊息會使用自己的格式包起來，所以我們必須拆解  才知道真正的錯誤是哪個
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
