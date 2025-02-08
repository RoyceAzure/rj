package random

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

const (
	alphabet string = "abcdefghijklmnopqrstuvwxyz"
	integer  string = "1234567890"
)

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomFloat(max float64) float64 {
	return rand.Float64() * max
}

func RandomString(n int) string {
	var sb strings.Builder

	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomStrInt(n int) string {
	return fmt.Sprintf("%d", rand.Intn(n))
}

func RandomFloatString(n int, point int) string {
	var sb strings.Builder
	k := len(integer)
	for i := 0; i < n; i++ {
		c := integer[rand.Intn(k)]
		if i == 0 && c == '0' {
			i--
			continue
		}
		sb.WriteByte(c)
	}
	segma := byte('.')
	sb.WriteByte(segma)
	for i := 0; i < point; i++ {
		c := integer[rand.Intn(k)]
		if i == 0 && c == '0' {
			i--
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomDecimal(n int, point int) (decimal.Decimal, error) {
	numeric_string := RandomFloatString(n, point)
	var de decimal.Decimal
	de, err := decimal.NewFromString(numeric_string)
	if err != nil {
		de = decimal.NewFromInt(RandomInt64(1, 10000))
	}
	return de, err
}

func RandomNumeric() pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(rand.Int63()),
		Exp:   int32(rand.Intn(10) - 5), // 隨機小數位數
		Valid: true,
	}
}
