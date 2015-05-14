package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/intogooglego/post"
	"io"
	"log"
	"net/http"
	"unicode"
)

func RedditPostScraper(sub string) (err error) {

	// connect to db using standard Go database/sql API
	//db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", "golang:mukkk@/golang")
	checkErr(err, "sql.Open failed")

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error()) // proper error handling instead of panic in your app
		return
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	defer dbmap.Db.Close()

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name
	//
	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	//dbmap.AddTableWithName(post.Post{}, "posts").SetKeys(true, "Id")
	table := dbmap.AddTableWithName(post.Post{}, "posts")
	table.SetKeys(true, "Id")
	table.ColMap("Site").SetMaxSize(32)
	table.ColMap("Site").SetNotNull(true)
	table.ColMap("PostId").SetMaxSize(32)
	table.ColMap("PostId").SetNotNull(true)
	// this creates an unique index on PostId
	table.ColMap("PostId").SetUnique(true)

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	resp, err := http.Get("http://www.reddit.com/r/" + sub + "/new")
	defer checkClose(resp.Body, &err)

	ps := make([]post.Post, 0)
	ps, err = ParseHtml(resp.Body, ps)

	if err == nil {
		// insert rows - auto increment PKs will be set properly after the insert
		for _, post := range ps {
			if post.Err == nil {

				// check if post already exists
				count, err := dbmap.SelectInt("select count(*) from posts where PostId = ?", post.PostId)
				checkErr(err, "select count(*) failed")

				if count == 0 {
					err = dbmap.Insert(&post)
					checkErr(err, "Insert failed")
					if err == nil {
						// Print out the crawled info
						fmt.Println(post.String())
						fmt.Println("-----------------------------------------------")
					}
				}
			} else {
				fmt.Println(post.Err)
			}
		}
	} else {
		fmt.Println(err)
	}

	return
}

func ParseHtml(io io.Reader, ps []post.Post) (psout []post.Post, err error) {

	// Create a qoquery document to parse from
	doc, err := goquery.NewDocumentFromReader(io)
	checkErr(err, "Failed to parse HTML")

	if err == nil {
		fmt.Println("---- Starting to parse ------------------------")

		// Find reddit posts = elements with class "thing"
		thing := doc.Find(".thing")
		for iThing := range thing.Nodes {

			post := post.NewPost()

			// use `single` as a selection of 1 node
			singlething := thing.Eq(iThing)

			// get the reddit post identifier
			reddit_post_id, exists := singlething.Attr("data-fullname")
			if exists == false {
				singlehtml, _ := singlething.Html()
				post.Err = fmt.Errorf("data-fullname not found in %s", singlehtml)
			} else {

				// find an element with class title and a child with class may-blank
				// Remove CRLF and unnecessary whitespaces
				post.Title = stringMinifier(singlething.Find(".title .may-blank").Text())

				post.PostId = reddit_post_id
				post.User = singlething.Find(".author").Text()
				post.Url, _ = singlething.Find(".comments.may-blank").Attr("href")
				post.SetScore(singlething.Find(".score.likes").Text())
				reddit_postdate, exists := singlething.Find("time").Attr("datetime")

				if exists == false {
					singlehtml, _ := singlething.Html()
					post.Err = fmt.Errorf("datetime not found in %s", singlehtml)
				} else {

					post.SetPostDate(reddit_postdate)

				}
			}
			ps = append(ps, post)

		}
	}

	return ps, err
}

func stringMinifier(in string) (out string) {

	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}

	return
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

// checkClose is used to check the return from Close in a defer
// statement.
func checkClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}

func main() {
	err := RedditPostScraper("golang")
	checkErr(err, "Failed to fetch from golang")
}
