package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RoyceAzure/rj/api"
	"github.com/RoyceAzure/rj/api/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthorizeUser(t *testing.T) {
	tokenMaker, err := token.NewPasetoMaker[uuid.UUID](generateRandomSymmetricKey(t))
	require.NoError(t, err)
	require.NotNil(t, tokenMaker)

	userID := uuid.New()
	upn := "test@test.com"
	duration := time.Minute

	token, payload, err := tokenMaker.CreateToken(upn, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)

	authorizer := api.NewAuthorizor(tokenMaker)

	authorizedCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", fmt.Sprintf("bearer %s", token)))
	payload, accessToken, err := authorizer.AuthorizUser(authorizedCtx)
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.NotEmpty(t, accessToken)

	require.Equal(t, payload.UPN, upn)
	require.Equal(t, payload.UserId, userID)
	require.WithinDuration(t, payload.IssuedAt, time.Now(), time.Second)
	require.WithinDuration(t, payload.ExpiredAt, time.Now().Add(duration), time.Second)
}
