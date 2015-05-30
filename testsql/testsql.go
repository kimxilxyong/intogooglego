package main

import (
	"database/sql"
	"errors"
	//"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"log"
)

// Print Debug info to stdout
var Debug bool = true

func TestSQL() (err error) {
	drivername := "postgres"
	dsn := "user=golang password=golang dbname=golang sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	//drivername := "mysql"
	//dsn := "golang:golang@/golang"
	//dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// connect to db using standard Go database/sql API
	//db, err := sql.Open("mysql", "golang:golang@/golang")
	//db, err := sql.Open("postgres", "user=golang password=golang dbname=golang sslmode=disable")

	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data:
	if err = db.Ping(); err != nil {
		return errors.New("db.Ping failed: " + err.Error())
	}

	postid := "t3_36qf3s"
	_, err = db.Query("SELECT * FROM posts WHERE postid=:postid", postid)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func main() {
	err := TestSQL()
	if err != nil {
		log.Fatalln("Failed TestSQL: ", err)
		panic(err)
	}
}
