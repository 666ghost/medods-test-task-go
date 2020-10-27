package models

import (
	"context"
	"github.com/666ghost/medods-test-task-go/connection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID   primitive.ObjectID `bson:"_id"`
	Guid string             `bson:"guid"`
}

func (u *User) Insert(ctx context.Context) (*mongo.InsertOneResult, error) {
	coll := connection.MGMain().Collection("users")

	insertResult, err := coll.InsertOne(ctx, u)
	return insertResult, err
}

func SelectUserByGuid(ctx context.Context, guid string) (*User, error) {
	u := new(User)
	coll := connection.MGMain().Collection("users")
	filter := bson.M{"guid": guid}
	err := coll.FindOne(ctx, filter).Decode(&u)
	return u, err
}

/*
func filterUsers(ctx context.Context, c *mongo.Collection, filter interface{}) ([]*User, error) {
	// A slice of tasks for storing the decoded documents
	var users []*User

	cur, err := c.Find(ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return users, err
		}

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	err = cur.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}
*/
