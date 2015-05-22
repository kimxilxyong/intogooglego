package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	"io"
	"log"
	"net/http"
	"unicode"
)

func RedditPostScraper(sub string) (err error) {

	// connect to db using standard Go database/sql API
	//db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", "golang:golang@/golang")
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return errors.New("db.Ping failed: " + err.Error())
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
	_ = dbmap.AddTableWithName(post.Post{}, "posts")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		return errors.New("Create table 'posts' failed: " + err.Error())
	}

	// Get data from reddit
	geturl := "http://www.reddit.com/r/" + sub + "/new"
	resp, err := http.Get(geturl)
	if err != nil {
		return errors.New("Failed to http.Get from " + geturl + ": " + err.Error())
	}
	if resp != nil {
		if resp.Body == nil {
			return errors.New("Body from " + geturl + " is nil!")
		} else {
			defer resp.Body.Close()
		}
	} else {
		return errors.New("Response from " + geturl + " is nil!")
	}
	if resp.StatusCode != 200 { // 200 = OK
		httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", geturl, resp.StatusCode, resp.Status)
		return errors.New(httperr)
	}

	// Create a new post slice and then parse the response body into ps
	ps := make([]post.Post, 0)
	ps, err = ParseHtmlReddit(resp.Body, ps)
	if err != nil {
		return errors.New("Error in RedditParseHtml: " + geturl + ": " + err.Error())
	}
	foundnewposts := false
	updatedposts := 0

	// insert rows - auto increment PKs will be set properly after the insert
	for _, post := range ps {
		if post.Err == nil {

			// check if post already exists
			count, err := dbmap.SelectInt("select count(*) from posts where PostId = ?", post.PostId)
			if err != nil {
				return errors.New("select count(*) from posts failed: " + err.Error())
			}

			if count == 0 {
				foundnewposts = true
				err = dbmap.Insert(&post)
				if err != nil {
					return errors.New("insert into table posts failed: " + err.Error())
				}
				if err == nil {
					// Print out the crawled info
					fmt.Println("----------- INSERT ----------------------------")
					fmt.Println(post.String())
				}
			} else {
				// Post already exists, do an update
				var updateid int64
				var score int64
				var err error
				updateid, err = dbmap.SelectInt("select PID from posts where PostId = ?", post.PostId)
				if err != nil {
					return errors.New("Failed: select PID from posts where PostId = " + post.PostId + ": " + err.Error())
				}
				post.Id = uint64(updateid)
				score, err = dbmap.SelectInt("select Score from posts where PID = ?", post.Id)
				if err != nil {
					return errors.New(fmt.Sprintf("Failed: select Score from posts where PID = %d", post.Id) + ": " + err.Error())
				}
				if score != int64(post.Score) {
					_, err = dbmap.Exec("update posts set Score = ? where PID = ?", post.Score, post.Id)

					if err != nil {
						return errors.New("update table 'posts' failed: " + err.Error())
					} else {
						updatedposts++
						// Print out the update info
						fmt.Println("----------- UPDATE SCORE-----------------------")
						fmt.Println(post.Title)
						fmt.Printf("From score %d to score %d\n", score, post.Score)
					}
				}
			}
		} else {
			fmt.Println("Single post error in " + geturl + ": " + post.Err.Error())
		}
	}
	if !foundnewposts {
		fmt.Println("No new posts found at " + geturl)
	}

	if updatedposts > 0 {
		fmt.Printf("%d posts have been updated from %s\n", updatedposts, geturl)
	}

	return
}

// Parse for posts in html from reddit, input html is an io.Reader and returns recognized posts in a psout slice of posts.
// Errors which affect only a single post are stored in their post.Err
func ParseHtmlReddit(io io.Reader, ps []post.Post) (psout []post.Post, err error) {

	// Create a qoquery document to parse from an io.Reader
	doc, err := goquery.NewDocumentFromReader(io)
	if err != nil {
		return ps, errors.New("Failed to parse HTML: " + err.Error())
	}

	// Find reddit posts = elements with class "thing"
	thing := doc.Find(".thing")
	for iThing := range thing.Nodes {

		// Create a new post struct - if the crawling fails the post will have an Err attached
		// but will be added to the outgoing (psout) slice nevertheless
		post := post.NewPost()

		// use `singlething` as a selection of one single post
		singlething := thing.Eq(iThing)

		// get the reddit post identifier
		reddit_post_id, exists := singlething.Attr("data-fullname")
		if exists == false {
			singlehtml, _ := singlething.Html()
			post.Err = fmt.Errorf("data-fullname not found in %s", singlehtml)
		} else {
			post.PostId = reddit_post_id
			// find an element with class title and a child with class may-blank
			// and remove CRLF and unnecessary whitespaces
			post.Title = stringMinifier(singlething.Find(".title .may-blank").Text())
			// Get the post user
			post.User = singlething.Find(".author").Text()
			// Get the post url
			post.Url, _ = singlething.Find(".comments.may-blank").Attr("href")
			// Get the post likes score
			post.SetScore(singlething.Find(".score.likes").Text())
			// Get the post date
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

func main() {
	err := RedditPostScraper("golang")
	if err != nil {
		log.Fatalln("Failed to fetch from sub reddit golang: ", err)
		panic(err)
	}
}
