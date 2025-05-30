package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker[T UserIDConstraint] struct {
	paseto      *paseto.V2
	symmerickey []byte
}

func NewPasetoMaker[T UserIDConstraint](symmerickey string) (Maker[T], error) {
	if len(symmerickey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid ket size : must be exactly %d charcters", chacha20poly1305.KeySize)
	}
	return &PasetoMaker[T]{paseto.NewV2(), []byte(symmerickey)}, nil
}

func (maker *PasetoMaker[T]) CreateToken(upn string, userID T, duration time.Duration) (string, *Payload[T], error) {
	payload, err := NewPayload(upn, userID, duration)
	if err != nil {
		return "", payload, err
	}

	//相比jwt之所以只有這行是因為你不需要決定加密演算法
	//固定使用chacha演算法
	token, err := maker.paseto.Encrypt(maker.symmerickey, payload, nil)
	return token, payload, err
}

func (maker *PasetoMaker[T]) VertifyToken(token string) (*Payload[T], error) {
	payload := &Payload[T]{}

	err := maker.paseto.Decrypt(token, maker.symmerickey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
