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

	//doc.Find(".thing").Each(func(i int, s *goquery.Selection) {
	//	aclass := s.Find(".may-blank").Text()

	//	fmt.Printf("aclass %d: %s\n", i, aclass)
	//fmt.Printf("thing: " + s.Html())
	//})

	//sel := doc.Find(".title .may-blank")
	sel := doc.Find(".thing")
	//fmt.Println("Sel: ", _+sel.Unwrap().Html())
	for i := range sel.Nodes {

		// use `single` as a selection of 1 node
		single := sel.Eq(i)

		// get the reddit post identifier
		reddit_post_id, exists := single.Attr("data-fullname")
		if exists == true {

			fmt.Println("reddit_post_id = " + reddit_post_id)

			title := single.Find("a[class] .title")

			html, err := title.Html()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Lenght ---------------- ", title.Length(), " index ", i)
			fmt.Println("title html=" + html)
			fmt.Println("title text=" + title.Text())

			//timestamp := sel.Find(".tagline")
			for iTime := range single.Find("a").Nodes {
				stamp := single.Eq(iTime)
				html, err := stamp.Html()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("time:" + html)
			}

			attribute, exists := single.Attr("href")
			if exists == true {

				fmt.Println("attribute=" + attribute)

				// create new post
				//post := rpcbotinterfaceobjects.NewPost(single.Text(), attribute)

				// insert rows - auto increment PKs will be set properly after the insert
				//err = dbmap.Insert(&post)
				//checkErr(err, "Insert failed")
			}
		}

	}
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func main() {
	ExampleScrape()
}
