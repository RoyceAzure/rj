package test

import (
	"testing"
	"time"

	"github.com/RoyceAzure/rj/api/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func generateRandomSymmetricKey(t *testing.T) string {
	t.Helper()
	key, err := token.GenerateRandomSymmetricKey()
	require.NoError(t, err)
	require.NotEmpty(t, key)
	return key
}
func TestCreatePasteoTokenUUID(t *testing.T) {
	tokenMaker, err := token.NewPasetoMaker[uuid.UUID](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	upn := "test@test.com"
	userID := uuid.New()
	duration := time.Minute

	token, payload, err := tokenMaker.CreateToken(upn, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)

	payload, err = tokenMaker.VertifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.Equal(t, payload.UPN, upn)
	require.Equal(t, payload.UserId, userID)
	require.WithinDuration(t, payload.IssuedAt, time.Now(), time.Second)
	require.WithinDuration(t, payload.ExpiredAt, time.Now().Add(duration), time.Second)
}

func TestCreatePasteoTokenInt64(t *testing.T) {
	tokenMaker, err := token.NewPasetoMaker[int64](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	upn := "test@test.com"
	userID := int64(1234567890)
	duration := time.Minute

	token, payload, err := tokenMaker.CreateToken(upn, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)

	payload, err = tokenMaker.VertifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.Equal(t, payload.UPN, upn)
	require.Equal(t, payload.UserId, userID)
	require.WithinDuration(t, payload.IssuedAt, time.Now(), time.Second)
	require.WithinDuration(t, payload.ExpiredAt, time.Now().Add(duration), time.Second)
}
