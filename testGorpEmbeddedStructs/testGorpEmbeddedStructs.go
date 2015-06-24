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
	"strconv"
	"time"
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
	fmt.Printf("AddTableWithName returned: %s\n", table.TableName)

	var r *gorp.RelationMap
	if len(table.Relations) > 0 {
		r = table.Relations[0]
		fmt.Printf("Relation DetailTable: %s\n", r.DetailTable.TableName)
	}

	// Add the comments table
	table = dbmap.AddTableWithName(post.Comment{}, "comments_embedded_test")
	table.SetKeys(true, "Id")
	fmt.Printf("AddTableWithName returned: %s\n", table.TableName)
	if r != nil {
		fmt.Printf("Relation DetailTable: %s\n", r.DetailTable.TableName)
	}

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
	var LastPkForGetTests uint64
	var p post.Post

	for i < 10 {
		p = post.NewPost()
		p.Title = fmt.Sprintf("Post number %d", i)
		p.PostDate = time.Now()
		p.WebPostId = strconv.FormatUint(post.Hash(p.Title), 10)

		x = 0
		for x < 10 {
			c := p.AddComment()
			c.Title = fmt.Sprintf("Comment %d on post %d", x, i)
			c.WebCommentId = strconv.FormatUint(post.Hash(c.Title), 10)
			x++
		}

		// Inserting a post also inserts all its detail records (=comments)
		err = dbmap.InsertWithChilds(&p)
		if DebugLevel > 3 {
			// Print out the crawled info
			fmt.Println("----------- INSERT POST START -----------------")
			fmt.Println(p.String("IP: "))
		}
		if err != nil {
			return errors.New("Insert failed: " + err.Error())
		}

		LastPkForGetTests = p.Id

		if DebugLevel > 3 {
			// Print out the end of the crawled info
			fmt.Println("----------- INSERT POST END -------------------")
		}

		for y, c := range p.Comments {

			c.Title = fmt.Sprintf("UpdatedComment %d ", y) + c.Title
			x++
		}

		p.Title = fmt.Sprintf("UpdatedPost %d ", i) + p.Title
		var rowsaffected int64
		rowsaffected, err = dbmap.UpdateWithChilds(&p)
		if DebugLevel > 3 {
			// Print out the crawled info
			fmt.Println("----------- UPDATE POST START -----------------")
			fmt.Printf("Rows affected: %d\n", rowsaffected)
			fmt.Println(p.String("UP: "))
		}
		if err != nil {
			return errors.New("update failed: " + err.Error())
		}
		if DebugLevel > 3 {
			// Print out the end of the crawled info
			fmt.Println("----------- UPDATE POST END -------------------")
		}

		i++

	}
	fmt.Println("Starting Get tests")

	res, err := dbmap.GetWithChilds(post.Post{}, LastPkForGetTests)

	if err != nil {
		return errors.New("get failed: " + err.Error())
	}
	if res == nil {
		return errors.New(fmt.Sprintf("Get post for id %d did not return any rows ", LastPkForGetTests))
	}

	resp := res.(*post.Post)

	if DebugLevel > 2 {
		// Print out the selected post
		fmt.Println("----------- GET POST START -----------------")
		fmt.Println(resp.String("GP: "))
	}

	if DebugLevel > 2 {
		// Print out the end of the selected post
		fmt.Println("----------- GET POST END -------------------")
	}

	err = AddUpdatableChilds(&p, resp, dbmap)
	if err != nil {
		return errors.New(fmt.Sprintf("AddUpdatableChilds for post '%s' failed: %s", resp.WebPostId, err.Error()))
	}

	var rowsaffected int64
	rowsaffected, err = dbmap.UpdateWithChilds(resp)
	if DebugLevel > 2 {
		// Print out the crawled info
		fmt.Println("----------- REUPDATE POST START -----------------")
		fmt.Printf("Rows affected: %d\n", rowsaffected)
		fmt.Println(resp.String("RUP: "))
	}
	if err != nil {
		return errors.New("reupdate failed: " + err.Error())
	}
	if DebugLevel > 2 {
		// Print out the end of the crawled info
		fmt.Println("----------- REUPDATE POST END -------------------")
	}

	return
}

func AddUpdatableChilds(htmlpost *post.Post, dbpost *post.Post, dbmap *gorp.DbMap) (err error) {
	// Check if there are aleady comments in dbpost
	// If not get them from the database

	if len(dbpost.Comments) == 0 {
		pk := dbpost.Id
		if pk == 0 {
			err = errors.New("primary key not set in dbpost")
			return
		}
		var res interface{}
		res, err = dbmap.GetWithChilds(post.Post{}, pk)
		if err != nil {
			err = errors.New("get failed: " + err.Error())
			return
		}
		if res == nil {
			err = errors.New(fmt.Sprintf("Get post for id %d did not return any rows ", pk))
			return
		}

		dbpost := res.(*post.Post)
		if DebugLevel > 2 {
			// Print out the update info
			fmt.Println("----------- DB POST -----------------")
			fmt.Println(dbpost.String("CHECK DB: "))
			fmt.Println("----------- DB POST END -------------------")
		}
	}
	if DebugLevel > 2 {
		// Print out the update info
		fmt.Println("----------- HTML POST -----------------")
		fmt.Println(htmlpost.String("CHECK HTML: "))
		fmt.Println("----------- HTML POST END -------------------")
	}

	updateNeeded := htmlpost.Hash() != dbpost.Hash()

	if updateNeeded {
		var UpdatedComments []*post.Comment
		var found bool
		for _, h := range htmlpost.Comments {
			found = false
			htmlHash := h.Hash()
			for _, d := range dbpost.Comments {
				if d.Hash() == htmlHash {
					// post with identical content has been found - do not store this comment
					found = true
					break
				}
				if h.WebCommentId == d.WebCommentId {
					// external unique comment id found - this comment is already stored
					// but the comment content has been changed - update needed
					h.Id = d.Id
					h.PostId = d.PostId
					break
				}
			}
			if !found {
				UpdatedComments = append(UpdatedComments, h)
				//htmlpost.Comments = append(htmlpost.Comments[:i], htmlpost.Comments[i+1:]...)
			}

		}
		dbpost.Comments = make([]*post.Comment, len(UpdatedComments), len(UpdatedComments))
		copy(dbpost.Comments, UpdatedComments)
	}
	if (DebugLevel > 2) && updateNeeded {
		// Print out the update info
		fmt.Println("----------- UPDATE NEEDED -----------------")

		for i := range htmlpost.Comments {
			fmt.Println(htmlpost.Comments[i].String("UPDATE NEEDED HTML: "))
			if i < len(dbpost.Comments) {
				fmt.Println(dbpost.Comments[i].String("UPDATE NEEDED DB: "))
			}
		}

		//fmt.Println(htmlpost.String("UPDATE NEEDED HTML: "))
		//fmt.Println(dbpost.String("UPDATE NEEDED DB: "))
		fmt.Println("----------- UPDATE NEEDED END -------------------")
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
