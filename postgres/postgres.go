package postgres

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func Init() {

	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatalf("database url is not specified")
	}

	var err error
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("can't connect to database: %v", err)
	}
}
