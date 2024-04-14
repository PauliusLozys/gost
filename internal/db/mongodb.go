package db

import (
	"context"
	"errors"
	"gost/internal/core"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	database          = "gost"
	messageCollection = "message"
	userCollection    = "user"
)

var ErrNotFound = errors.New("not found")

type MongoRepo struct {
	db *mongo.Client
}

func New() (*MongoRepo, func(), error) {
	options := options.Client()
	options.ApplyURI("mongodb://root:pass@localhost:27017")

	client, err := mongo.Connect(context.Background(), options)
	if err != nil {
		return nil, nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctxTimeout, readpref.Primary()); err != nil {
		return nil, nil, err
	}

	return &MongoRepo{
			db: client,
		},
		func() {
			if err = client.Disconnect(context.Background()); err != nil {
				slog.Error("disconnecting mangodb instance", "error", err)
			}
		}, nil
}

func (m *MongoRepo) MessageInsert(ctx context.Context, msg core.Message) error {
	collection := m.db.Database(database).Collection(messageCollection)
	_, err := collection.InsertOne(ctx, msg)
	return err
}

func (m *MongoRepo) MessageSelectDesc(ctx context.Context) ([]core.Message, error) {
	collection := m.db.Database(database).Collection(messageCollection)
	cur, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var messages []core.Message
	if err := cur.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MongoRepo) UserInsert(ctx context.Context, user core.User) error {
	collection := m.db.Database(database).Collection(userCollection)
	_, err := collection.InsertOne(ctx, user)
	return err
}

func (m *MongoRepo) UserGetByIP(ctx context.Context, ip string) (*core.User, error) {
	collection := m.db.Database(database).Collection(userCollection)
	result := collection.FindOne(ctx, bson.M{"ip": ip})
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var user core.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
