package rjhttp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TwseResponse struct {
	Stat       string     `json:"stat"`
	Date       string     `json:"date"`
	Title      string     `json:"title"`
	Fields     []string   `json:"fields"`
	Data       [][]string `json:"data"`
	TotalCount int        `json:total`
}

func TestGet(t *testing.T) {
	client := NewHttpClient(WithTimeout(time.Second * 30))

	url, err := UrlWithParams(
		"https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY",
		map[string]string{
			"date":     "20100101",
			"stockNo":  "0050",
			"response": "json",
		},
	)
	require.Nil(t, err)

	res, err := client.Get(context.Background(), url)
	require.Nil(t, err)
	var resBody TwseResponse

	err = json.Unmarshal(res, &resBody)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, resBody.Fields)
	require.NotEmpty(t, resBody.Data)
}
