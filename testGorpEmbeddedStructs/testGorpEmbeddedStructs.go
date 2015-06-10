package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	_ "github.com/lib/pq"

	"log"
	"os"
)

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 3

func Test() (err error) {
	//drivername := "postgres"
	//dsn := "user=golang password=golang dbname=golang sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	drivername := "mysql"
	dsn := "golang:golang@/golang?parseTime=true"
	dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data using ping
	if err = db.Ping(); err != nil {
		return errors.New("db.Ping failed: " + err.Error())
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	defer dbmap.Db.Close()
	dbmap.DebugLevel = DebugLevel
	// Will log all SQL statements + args as they are run
	// The first arg is a string prefix to prepend to all log messages
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "fetch:", log.Lmicroseconds))

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name

	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	table := dbmap.AddTableWithName(post.Post{}, "posts_embedded_test")
	table.SetKeys(true, "PID")

	// Add the comments table
	//table = dbmap.AddTableWithName(post.Comment{}, "comments_embedded_test")
	//table.SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	if err = dbmap.CreateTablesIfNotExists(); err != nil {
		return errors.New("Create tables failed: " + err.Error())
	}

	// Force create all indexes for this database
	if err = dbmap.CreateIndexes(); err != nil {
		return errors.New("Create indexes failed: " + err.Error())
	}

	i := 0
	x := 0

	for i < 10 {
		p := post.NewPost()
		p.Title = fmt.Sprintf("Post number %d", i)
		x = 0
		for x < 10 {
			c := post.NewComment()
			c.Title = fmt.Sprintf("Comment %d on post %d", x, i)
			p.Comments = append(p.Comments, &c)

			// Insert the new post into the database
			err = dbmap.Insert(&p)
			if DebugLevel > 2 {
				// Print out the crawled info
				fmt.Println("----------- INSERT POST START -----------------")
				fmt.Println(p.String())
			}
			if err != nil {
				return errors.New("insert failed: " + err.Error())
			}
			if DebugLevel > 2 {
				// Print out the end of the crawled info
				fmt.Println("----------- INSERT POST END -------------------")
			}

		}
		i++

	}

	return
}

func main() {
	err := Test()
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Test failed: ", err)
			panic(err)
		}
	}
}
