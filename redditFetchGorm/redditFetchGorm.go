package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/kimxilxyong/intogooglego/post"
	"io"
	"log"
	"net/http"
	"unicode"
)

func RedditPostScraperWithGorm(sub string) (err error) {

	db, err := gorm.Open("mysql", "golang:golang@/golang?charset=utf8&parseTime=True&loc=Local")

	// Get database connection handle [*sql.DB](http://golang.org/pkg/database/sql/#DB)
	db.DB()

	// Then you could invoke `*sql.DB`'s functions with it
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Automating Migration
	db.AutoMigrate(&post.Post{})

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
	for _, p := range ps {
		if p.Err == nil {

			p.Site = "reddit"

			// check if post already exists
			var count int
			querypost := post.NewPost()
			// SELECT * FROM posts WHERE PostId = "xxx";
			db.Find(&querypost, "PostId = ?", p.PostId).Count(&count)
			err = db.Error
			if err != nil {
				return errors.New("select count(*) from posts failed: " + err.Error())
			}

			if count == 0 {
				foundnewposts = true
				db.Create(&p)
				err = db.Error
				if err != nil {
					return errors.New("insert into table posts failed: " + err.Error())
				}
				if err == nil {
					// Print out the crawled info
					fmt.Println("----------- INSERT ----------------------------")
					fmt.Println(p.String())
				}
			} else {
				// Post already exists, do an update
				//var err error

				score := p.Score

				// Get the first matched record
				// SELECT * FROM posts WHERE PostId = 'xxx' limit 1;
				db.Where("PostId = ?", p.PostId).First(&querypost)
				err = db.Error
				if err != nil {
					return errors.New("Failed: select Id from posts where PostId = " + p.PostId + ": " + err.Error())
				}

				if score != querypost.Score {
					db.Save(&p)
					err = db.Error
					if err != nil {
						return errors.New("update table 'posts' failed: " + err.Error())
					} else {
						updatedposts++
						// Print out the update info
						fmt.Println("----------- UPDATE SCORE-----------------------")
						fmt.Println(p.Title)
						fmt.Printf("From %d to %d\n", score, p.Score)
					}
				}
			}
		} else {
			fmt.Println("Single post error in " + geturl + ": " + p.Err.Error())
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
	err := RedditPostScraperWithGorm("golang")
	if err != nil {
		log.Fatalln("Failed to fetch from sub reddit golang: ", err)
	}
}
