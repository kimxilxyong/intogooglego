package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/rpcbotinterfaceobjects"
	"log"
	"os"
	"unicode"
)

func ExampleScrape() {

	// connect to db using standard Go database/sql API
	//db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", "golang:mukkk@/golang")
	checkErr(err, "sql.Open failed")

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
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
	dbmap.AddTableWithName(rpcbotinterfaceobjects.Post{}, "post_test").SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	//s := `<p class="title"><a class="title may-blank " href="https://kabukky.github.io/journey/" tabindex="1">Journey - A blog engine written in Go, compatible with Ghost themes</a> <span class="domain">(<a href="/domain/kabukky.github.io/">kabukky.github.io</a>)</span></p><p class="tagline">submitted <time title="Sat Apr 25 14:56:54 2015 UTC" datetime="2015-04-25T14:56:54+00:00" class="live-timestamp">18 hours ago</time> by <a href="http://www.reddit.com/user/Kabukks" class="author may-blank id-t2_5j62h">Kabukks</a><span class="userattrs"></span></p><ul class="flat-list buttons"><li class="first"><a href="http://www.reddit.com/r/golang/comments/33tnj8/journey_a_blog_engine_written_in_go_compatible/" class="comments may-blank">9 comments</a></li><li class="share"><span class="share-button toggle" style=""><a class="option active login-required" href="#" tabindex="100">share</a><a class="option " href="#">cancel</a></span></li></ul><div class="expando" style="display: none"><span class="error">loading...</span></div>`
	filereader, err := os.Open("RedditCode.html")
	checkErr(err, "Read file failed")
	defer filereader.Close()
	//doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	doc, err := goquery.NewDocumentFromReader(filereader)
	if err != nil {
		log.Fatal(err)
	}

	thing := doc.Find(".thing")
	for iThing := range thing.Nodes {

		// use `single` as a selection of 1 node
		singlething := thing.Eq(iThing)

		// get the reddit post identifier
		reddit_post_id, exists := singlething.Attr("data-fullname")
		if exists == true {

			reddit_post_title := singlething.Find(".title .may-blank").Text()
			reddit_post_user := singlething.Find(".author").Text()
			reddit_post_url, _ := singlething.Find(".title .may-blank").Attr("href")
			reddit_post_score := singlething.Find(".score.likes").Text()

			fmt.Println("reddit_post_score = " + reddit_post_score)

			for i, v := range singlething.Find(".score").Nodes {
				fmt.Printf("Node i %v\n", i)
				fmt.Printf("Node Data %v\n", v.Data)
				for k, w := range v.Attr {
					fmt.Printf("Attr %v\n", k)
					fmt.Printf("Attr Key %v\n", w.Key)
					fmt.Printf("Attr Val %v\n", w.Val)

				}
			}

			postdate, exists := singlething.Find("time").Attr("datetime")
			if exists == true {

				// Remove CRLF
				//reddit_post_title = strings.Replace(reddit_post_title, "\n", "", -1)
				reddit_post_title = stringMinifier(reddit_post_title)

				fmt.Println("Date = " + postdate)
				fmt.Println("User = " + reddit_post_user)
				fmt.Println("Title = " + reddit_post_title)
				fmt.Println("Score = " + reddit_post_score)

				// create new post
				/*post := rpcbotinterfaceobjects.NewPost(site: "reddit.com",
				         postid = reddit_post_id, postdate = postdate,
						score = reddit_post_score, title = reddit_post_title, url = reddit_post_url,
						user = reddit_post_user)
				*/
				post := rpcbotinterfaceobjects.NewPost(
					"reddit.com",
					reddit_post_id, postdate,
					reddit_post_score,
					reddit_post_title,
					reddit_post_url,
					reddit_post_user)

				// insert rows - auto increment PKs will be set properly after the insert
				err = dbmap.Insert(&post)
				checkErr(err, "Insert failed")
			}
		}

	}
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
