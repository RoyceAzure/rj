package test

import (
	"testing"
	"time"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/require"
)

func TestElasticDao(t *testing.T) {

	type Tweet struct {
		User     string                `json:"user"`
		Message  string                `json:"message"`
		Retweets int                   `json:"retweets"`
		Image    string                `json:"image,omitempty"`
		Created  time.Time             `json:"created,omitempty"`
		Tags     []string              `json:"tags,omitempty"`
		Location string                `json:"location,omitempty"`
		Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
	}

	err := elsearch.InitELSearch("http://localhost:9200", "password")
	require.NoError(t, err)

	dao, err := elsearch.GetInstance()
	require.NoError(t, err)

	jsonString := `{
		"name": "範例物件",
		"id": 12345,
		"isActive": true,
		"tags": ["go", "json", "byte"]
	  }`

	jsonBytes := []byte(jsonString)
	// 3. 印出結果 (可選)
	// tweet := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	// jsonBytes, err := json.Marshal(tweet)
	require.NoError(t, err)
	err = dao.Create("test", jsonBytes)
	require.NoError(t, err)

}
