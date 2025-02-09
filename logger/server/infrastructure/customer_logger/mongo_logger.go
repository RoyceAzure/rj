package customerlogger

import (
	"context"
	"encoding/json"
	"fmt"

	repository "github.com/RoyceAzure/rj/repo/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoLogger struct {
	mongoDao repository.IMongoDao
}

func NewMongoLogger(mongoDao repository.IMongoDao) *MongoLogger {
	return &MongoLogger{
		mongoDao: mongoDao,
	}
}

func (mw *MongoLogger) Write(p []byte) (n int, err error) {
	// Insert the record into the collection.
	if mw == nil {
		return 0, fmt.Errorf("mongo logger is not init")
	}

	var logEntry bson.M
	err = json.Unmarshal(p, &logEntry)
	if err != nil {
		return 0, err
	}
	err = mw.mongoDao.InsertBsonM(context.Background(), logEntry)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
