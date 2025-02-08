package token

import "time"

type Maker interface {
	CreateToken(upn string, userID int64, duration time.Duration) (string, *Payload, error)
	VertifyToken(token string) (*Payload, error)
}
