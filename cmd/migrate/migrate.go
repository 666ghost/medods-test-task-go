package main

import (
	"fmt"
	"github.com/666ghost/medods-test-task-go/config"
	"github.com/666ghost/medods-test-task-go/connection"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
)

var action string

func init() {
	config.LoadFromFile()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Set action argument: drop, down or up\nFor example: ./migrate up")
		return
	}
	action = os.Args[1]

	cfg := config.New()
	client := connection.MGMain().Client()

	driver, err := mongodb.WithInstance(client, &mongodb.Config{DatabaseName: cfg.DbName, TransactionMode: action != "down"})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./db/migration/",
		"medods", driver)
	if err != nil {
		panic(err)
	}
	switch action {
	case "drop":
		err = m.Drop()
		fmt.Println("Dropping schema...")

	case "down":
		err = m.Steps(-2)
		fmt.Println("Downing migrations...")

	case "up":
		err = m.Steps(2)

		fmt.Println("Upping migrations...")
	}

	if err == os.ErrNotExist {
		fmt.Println("Nothing to migrate")
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully migrated")
}
