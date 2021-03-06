package connection

import (
	"context"
	"fmt"
	"github.com/666ghost/medods-test-task-go/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	mgMainOnce sync.Once
	mgMain     *mongo.Database
)

func MGMain() *mongo.Database {
	cfg := config.New()
	mgMainOnce.Do(func() {
		mgMain = NewMongo(cfg)
	})
	return mgMain
}

func NewMongo(cfg *config.Config) *mongo.Database {

	// Ping the primary
	client, cancel := GetClient(cfg)
	defer cancel()

	db := client.Database(cfg.DbName)

	return db
}

func GetClient(cfg *config.Config) (*mongo.Client, context.CancelFunc) {
	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority",
		cfg.DbUser, cfg.DbPassword, strings.Join(cfg.DbAddresses, ","))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	OnExitSecondary(func(ctx context.Context) {
		if err := client.Disconnect(ctx); err != nil {
			log.Panic("pg.Close Failed")
		}
	})
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err) // todo logger
	}
	log.Print("connected to mongo")
	return client, cancel
}
