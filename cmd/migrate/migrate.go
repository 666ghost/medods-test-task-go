package main

import (
	"fmt"
	"github.com/666ghost/medods-test-task-go/config"
	"github.com/666ghost/medods-test-task-go/connection"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"time"
)

var action string

func init() {
	config.LoadFromFile()
}

func main() {
	now := time.Now()
	file, err := os.OpenFile("logs/migrate_logfile_"+now.Format("20060102")+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.Print("Migrating db...")

	if len(os.Args) < 2 {
		fmt.Println("Set action argument: drop, down or up\nFor example: ./migrate up")
		return
	}
	action = os.Args[1]

	cfg := config.New()
	client := connection.MGMain().Client()

	driver, err := mongodb.WithInstance(client, &mongodb.Config{DatabaseName: cfg.DbName})
	if err != nil {
		log.Fatal("Failed setting driver up ", err)
		return
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./db/migration/",
		"medods", driver)
	if err != nil {
		log.Fatal("Failed creating new database instance ", err)
		return
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
		log.Fatal("Failed migrating db", err)

	}
	fmt.Println("Successfully migrated")
}
