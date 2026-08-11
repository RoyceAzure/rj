package test

import (
	"testing"
	"time"

	"github.com/RoyceAzure/rj/api/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateJWTTokenUUID(t *testing.T) {
	tokenMaker, err := token.NewJWTMaker[uuid.UUID](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	upn := "test@test.com"
	userID := uuid.New()
	duration := time.Minute

	tok, payload, err := tokenMaker.CreateToken(upn, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	require.NotNil(t, payload)

	payload, err = tokenMaker.VertifyToken(tok)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.Equal(t, payload.UPN, upn)
	require.Equal(t, payload.UserId, userID)
	require.WithinDuration(t, payload.IssuedAt, time.Now(), time.Second)
	require.WithinDuration(t, payload.ExpiredAt, time.Now().Add(duration), time.Second)
}

func TestCreateJWTTokenInt64(t *testing.T) {
	tokenMaker, err := token.NewJWTMaker[int64](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	upn := "test@test.com"
	userID := int64(1234567890)
	duration := time.Minute

	tok, payload, err := tokenMaker.CreateToken(upn, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	require.NotNil(t, payload)

	payload, err = tokenMaker.VertifyToken(tok)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.Equal(t, payload.UPN, upn)
	require.Equal(t, payload.UserId, userID)
	require.WithinDuration(t, payload.IssuedAt, time.Now(), time.Second)
	require.WithinDuration(t, payload.ExpiredAt, time.Now().Add(duration), time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	tokenMaker, err := token.NewJWTMaker[uuid.UUID](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	tok, payload, err := tokenMaker.CreateToken("test@test.com", uuid.New(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	require.NotNil(t, payload)

	payload, err = tokenMaker.VertifyToken(tok)
	require.Equal(t, token.ErrExpiredToken, err)
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := token.NewPayload("test@test.com", uuid.New(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	unsignedToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	tokenMaker, err := token.NewJWTMaker[uuid.UUID](generateRandomSymmetricKey(t))
	require.NoError(t, err)

	payload, err = tokenMaker.VertifyToken(unsignedToken)
	require.Equal(t, token.ErrInvalidToken, err)
	require.Nil(t, payload)
}
