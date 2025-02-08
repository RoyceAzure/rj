package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	LOG_DATABASE         = "logs"
	LOG_INFO_COLLECTION  = "infos"
	LOG_WARN_COLLECTION  = "warns"
	LOG_TRACE_COLLECTION = "traces"
	LOG_ERR_COLLECTION   = "errors"
	LevelTraceValue      = "trace"
	LevelDebugValue      = "debug"
	LevelInfoValue       = "info"
	LevelWarnValue       = "warn"
	LevelErrorValue      = "error"
	LevelFatalValue      = "fatal"
	LevelPanicValue      = "panic"
)

type IMongoDao interface {
	Insert(ctx context.Context, entry LogEntry) error
	GetAll(ctx context.Context) ([]*LogEntry, error)
	InsertBsonM(ctx context.Context, entry primitive.M) error
}

type MongoDao struct {
	client *mongo.Client
}

type LogEntry struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	ServiceName string    `bson:"service_name" json:"service_name"`
	Level       string    `bson:"level" json:"level"`
	Message     string    `bson:"message" json:"message"`
	Error       string    `bson:"error" json:"error"`
	CreatedAt   time.Time `bson:"time" json:"time"`
}

func InitCollection(client *mongo.Client) error {
	if client == nil {
		return fmt.Errorf("mongo client is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	err := client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	database := client.Database(LOG_DATABASE)

	collections := []string{LOG_INFO_COLLECTION, LOG_WARN_COLLECTION, LOG_ERR_COLLECTION, LOG_TRACE_COLLECTION}

	for _, collection := range collections {
		// options := options.CreateCollectionOptions{}
		err := database.CreateCollection(ctx, collection, nil)
		if err != nil {
			if commandErr, ok := err.(mongo.CommandError); ok {
				if commandErr.Code == 48 {
					continue
				}
			}
			return err
		}
	}

	return nil
}

func NewMongoDao(client *mongo.Client) IMongoDao {
	return &MongoDao{
		client: client,
	}
}

func (dao *MongoDao) Insert(ctx context.Context, entry LogEntry) error {

	var collection *mongo.Collection

	database := dao.client.Database(LOG_DATABASE)

	switch entry.Level {
	case LevelInfoValue:
		collection = database.Collection(LOG_INFO_COLLECTION)
	case LevelWarnValue:
		collection = database.Collection(LOG_WARN_COLLECTION)
	case LevelTraceValue:
		collection = database.Collection(LOG_TRACE_COLLECTION)
	case LevelErrorValue:
		collection = database.Collection(LOG_ERR_COLLECTION)
	default:
		collection = database.Collection(LOG_INFO_COLLECTION)
	}

	_, err := collection.InsertOne(ctx, entry)

	return err
}

func (dao *MongoDao) InsertBsonM(ctx context.Context, entry primitive.M) error {

	var collection *mongo.Collection

	database := dao.client.Database(LOG_DATABASE)

	level, ok := entry["level"].(string)
	if !ok {
		return fmt.Errorf("insert bson to db get err, can't fecth level")
	}

	switch level {
	case LevelInfoValue:
		collection = database.Collection(LOG_INFO_COLLECTION)
	case LevelWarnValue:
		collection = database.Collection(LOG_WARN_COLLECTION)
	case LevelTraceValue:
		collection = database.Collection(LOG_TRACE_COLLECTION)
	case LevelErrorValue:
		collection = database.Collection(LOG_ERR_COLLECTION)
	default:
		collection = database.Collection(LOG_INFO_COLLECTION)
	}

	_, err := collection.InsertOne(ctx, entry)

	return err
}

func (dao *MongoDao) GetAll(ctx context.Context) ([]*LogEntry, error) {
	collection := dao.client.Database(LOG_DATABASE).Collection(LOG_INFO_COLLECTION)
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry

		err := cursor.Decode(&item)
		if err != nil {
			return nil, err
		} else {
			logs = append(logs, &item)
		}
	}

	return logs, nil
}

func ConnectToMongo(ctx context.Context, address string) (*mongo.Client, error) {
	//自己設定clientOptions  連線資訊
	//透過mongo套件使用自己設定的clientOptions連線
	clientOptions := options.Client().ApplyURI(address)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	return c, nil
}
