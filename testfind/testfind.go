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
	"os"
	"unicode"
)

func ExampleScrape() (err error) {

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

	//s := `<p class="title"><a class="title may-blank " href="https://kabukky.github.io/journey/" tabindex="1">Journey - A blog engine written in Go, compatible with Ghost themes</a> <span class="domain">(<a href="/domain/kabukky.github.io/">kabukky.github.io</a>)</span></p><p class="tagline">submitted <time title="Sat Apr 25 14:56:54 2015 UTC" datetime="2015-04-25T14:56:54+00:00" class="live-timestamp">18 hours ago</time> by <a href="http://www.reddit.com/user/Kabukks" class="author may-blank id-t2_5j62h">Kabukks</a><span class="userattrs"></span></p><ul class="flat-list buttons"><li class="first"><a href="http://www.reddit.com/r/golang/comments/33tnj8/journey_a_blog_engine_written_in_go_compatible/" class="comments may-blank">9 comments</a></li><li class="share"><span class="share-button toggle" style=""><a class="option active login-required" href="#" tabindex="100">share</a><a class="option " href="#">cancel</a></span></li></ul><div class="expando" style="display: none"><span class="error">loading...</span></div>`
	filereader, err := os.Open("RedditCode.html")
	checkErr(err, "Read file failed")
	defer filereader.Close()

	ps := make([]post.Post, 0)

	ps, err = ParseHtml(filereader, ps)

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

					//post := post.NewPost()
					post.SetPostDate(reddit_postdate)

					// Print out the crawled info
					post.String()
					fmt.Println("-----------------------------------------------")

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

func main() {
	ExampleScrape()
}
