package models

import (
	"context"
	"github.com/666ghost/medods-test-task-go/connection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Token struct {
	ID      primitive.ObjectID `bson:"_id"`
	UserId  primitive.ObjectID `bson:"user_id"`
	Refresh string             `bson:"refresh"`
	Used    bool               `bson:"used"`
}

func (u *Token) Insert(ctx context.Context) (*mongo.InsertOneResult, error) {
	coll := connection.MGMain().Collection("tokens")

	insertResult, err := coll.InsertOne(ctx, u)
	return insertResult, err
}
func (u *Token) Update(ctx context.Context, data primitive.D) (*mongo.UpdateResult, error) {
	coll := connection.MGMain().Collection("tokens")

	filter := bson.M{"_id": u.ID}
	result, err := coll.UpdateMany(
		ctx,
		filter,
		bson.D{
			{"$set", data},
		},
	)
	return result, err
}
func RemoveToken(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {

	coll := connection.MGMain().Collection("tokens")

	result, err := coll.DeleteMany(ctx, filter)

	return result, err
}

func SelectUnusedTokenById(ctx context.Context, id string) (*Token, error) {
	t := new(Token)
	coll := connection.MGMain().Collection("tokens")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}
	log.Print("objectVal ", objectId)
	filter := bson.M{"_id": objectId, "used": false}
	err = coll.FindOne(ctx, filter).Decode(&t)

	return t, err
}
func InvalidateOldUserRefreshTokens(ctx context.Context, u *User) (*mongo.UpdateResult, error) {
	coll := connection.MGMain().Collection("tokens")

	filter := bson.M{"user_id": u.ID, "used": false}
	documents, _ := coll.CountDocuments(ctx, filter)
	log.Print("documents count", documents)
	result, err := coll.UpdateMany(
		ctx,
		filter,
		bson.D{
			{"$set", bson.D{{"used", true}}},
		},
	)

	return result, err
}
