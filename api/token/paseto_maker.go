package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto      *paseto.V2
	symmerickey []byte
}

func NewPasetoMaker(symmerickey string) (Maker, error) {
	if len(symmerickey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid ket size : must be exactly %d charcters", chacha20poly1305.KeySize)
	}
	return &PasetoMaker{paseto.NewV2(), []byte(symmerickey)}, nil
}

func (maker *PasetoMaker) CreateToken(upn string, userID int64, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(upn, userID, duration)
	if err != nil {
		return "", payload, err
	}

	//相比jwt之所以只有這行是因為你不需要決定加密演算法
	//固定使用chacha演算法
	token, err := maker.paseto.Encrypt(maker.symmerickey, payload, nil)
	return token, payload, err
}

func (maker *PasetoMaker) VertifyToken(token string) (*Payload, error) {
	payload := &Payload{}

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
