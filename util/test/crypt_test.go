package test

import (
	"testing"

	"github.com/RoyceAzure/rj/util/crypt"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "@Aa123456789"
	hashedPassword, err := crypt.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	correcetPassword := "@Aa123456789"
	err = crypt.CheckPassword(correcetPassword, hashedPassword)
	require.NoError(t, err)

	wrongPassword := "@Aa1234567890"
	err = crypt.CheckPassword(wrongPassword, hashedPassword)
	require.Error(t, err)
}
