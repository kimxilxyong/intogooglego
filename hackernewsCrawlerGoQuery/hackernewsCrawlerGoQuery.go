package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	_ "github.com/lib/pq"

	"io"
	"io/ioutil"

	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 3

func HackerNewsPostScraper(sub string) (err error) {
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
	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "fetch:", log.Lmicroseconds))

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name
	tablename := "posts_index_test"
	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	table := dbmap.AddTableWithName(post.Post{}, tablename)
	table.SetKeys(true, "PID")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	if err = dbmap.CreateTablesIfNotExists(); err != nil {
		return errors.New("Create tables failed: " + err.Error())
	}

	// Force create all indexes for this database
	if err = dbmap.CreateIndexes(); err != nil {
		return errors.New("Create indexes failed: " + err.Error())
	}

	// Get data from hackernews
	geturl := "http://news.ycombinator.com/" + sub
	body, err := GetHtmlBody(geturl)
	if err != nil {
		return errors.New("GetHtmlBody: " + err.Error())
	}

	// Create a new post slice and then parse the response body into ps
	ps := make([]*post.Post, 0)
	ps, err = ParseHtmlHackerNews(body, ps)
	if err != nil {
		return errors.New("ParseHtmlHackerNews: " + err.Error())
	}

	// Number of updated posts
	var updatedPostsCount uint32
	// Number of new posts
	var insertedPostsCount uint32

	// insert rows - auto increment PKs will be set properly after the insert
	for _, htmlpost := range ps {

		if htmlpost.PostId == "" {
			if DebugLevel > 1 {
				fmt.Printf("PostId not set in %s\n", htmlpost.Title)
			}
			// Fail early, continue with next post
			continue
		}

		if htmlpost.Err != nil {
			if DebugLevel > 1 {
				fmt.Println("Single post error in " + geturl + ": " + htmlpost.Err.Error())
			}
			// Fail early, continue with next post
			continue
		}

		// Store post sub
		htmlpost.PostSub = sub

		// check if post already exists
		intSelectResult := make([]int, 0)
		postcountsql := "select count(*) from " + dbmap.Dialect.QuoteField(tablename) +
			" where PostId = :post_id"
		_, err := dbmap.Select(&intSelectResult, postcountsql, map[string]interface{}{
			"post_id": htmlpost.PostId,
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Query: %s failed: %s\n", postcountsql, err.Error()))
		}
		if len(intSelectResult) == 0 {
			return errors.New(fmt.Sprintf("Query: %s returned no result\n", postcountsql))
		}
		postcount := intSelectResult[0]

		// New post? then insert
		if postcount == 0 {

			// Insert the new post into the database
			err = dbmap.Insert(htmlpost)

			if DebugLevel > 2 {
				// Print out the crawled info
				fmt.Println("----------- INSERT POST START -----------------")
				fmt.Println(htmlpost.String())
			}
			if err != nil {
				return errors.New("insert into table " + dbmap.Dialect.QuoteField(tablename) + " failed: " + err.Error())
			}
			if DebugLevel > 2 {
				// Print out the end of the crawled info
				fmt.Println("----------- INSERT POST END -------------------")
			}
			insertedPostsCount++

		} else {
			// Post already exists, do an update
			// Create a slice of posts to select into
			dbposts := make([]post.Post, 0)
			getpostsql := "select * from " + dbmap.Dialect.QuoteField(tablename) + " where PostId = :post_id"
			_, err := dbmap.Select(&dbposts, getpostsql, map[string]interface{}{
				"post_id": htmlpost.PostId,
			})
			if err != nil {
				return errors.New(fmt.Sprintf("Getting PostId %s from DB failed: %s\n", htmlpost.PostId, err.Error()))
			}
			var dbpost post.Post
			if len(dbposts) > 0 {
				dbpost = dbposts[0]
			} else {
				return errors.New(fmt.Sprintf("Query: %s returned no result\n", getpostsql))
			}

			if htmlpost.Score != dbpost.Score {
				// The post score changed, do an update into the database

				//fmt.Println("Post Date db: " + dbpost.PostDate.String() + ", html: " + htmlpost.PostDate.String())
				//fmt.Printf("Post Score db: %d, html: %d\n", dbpost.Score, htmlpost.Score)

				if DebugLevel > 2 {
					fmt.Println("----------- UPDATE POST START -----------------")
					fmt.Println(dbpost.String())
					fmt.Printf("From score %d to score %d\n", dbpost.Score, htmlpost.Score)
				}
				dbpost.Score = htmlpost.Score
				dbpost.PostDate = htmlpost.PostDate
				affectedrows, err := dbmap.Update(&dbpost)
				switch {
				case err != nil:
					return errors.New("update table " + tablename + " failed: " + err.Error())
				case affectedrows == 0:
					return errors.New(fmt.Sprintf("update table %s for Id %d did not affect any lines", tablename, dbpost.Id))
				default:
					updatedPostsCount++
					if DebugLevel > 2 {
						// Print out the update info
						fmt.Println("----------- UPDATE POST COMMIT -----------------")
						fmt.Println(dbpost.String())
						fmt.Println("----------- UPDATE POST END -------------------")
					}
				}
			}

		}
	}
	if insertedPostsCount == 0 && updatedPostsCount == 0 {
		if DebugLevel > 2 {
			fmt.Println("No new posts found at " + geturl)
		}
	}

	if updatedPostsCount > 0 && DebugLevel > 2 {
		fmt.Printf("%d existing posts have been updated from %s\n", updatedPostsCount, geturl)
	}

	if insertedPostsCount > 0 && DebugLevel > 2 {
		fmt.Printf("%d new posts have been inserted from %s\n", insertedPostsCount, geturl)
	}

	return
}

func GetHtmlBody(url string) (body io.Reader, err error) {

	// Get data from url
	resp, err := http.Get(url)
	if err != nil {
		err = errors.New("Failed to http.Get from " + url + ": " + err.Error())
		return
	}
	if resp != nil {
		defer resp.Body.Close()

		// capture all bytes from the response body
		buf, err := ioutil.ReadAll(resp.Body)
		body = bytes.NewReader(buf)

		if resp.StatusCode != 200 { // 200 = OK
			httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", url, resp.StatusCode, resp.Status)
			err = errors.New(httperr)
			return body, err
		}

		return body, err

	} else {
		err = errors.New("Response from " + url + " is nil")
		return
	}

	return
}

// Parse for posts in html from hackernews, input html is an io.Reader and returns recognized posts in a psout slice of posts.
// Errors which affect only a single post are stored in their post.Err
func ParseHtmlHackerNews(body io.Reader, ps []*post.Post) (psout []*post.Post, err error) {

	var html string
	// Create a qoquery document to parse from an io.Reader
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return ps, errors.New("Failed to parse HTML: " + err.Error())
	}

	// Find hackernews posts = elements with class "athing"
	thing := doc.Find(".athing")
	for iThing := range thing.Nodes {

		// Create a new post struct - if the crawling fails the post will have an Err attached
		// but will be added to the outgoing (psout) slice nevertheless
		post := post.NewPost()
		ps = append(ps, &post)

		// use `singlearticle` as a selection of one single post
		singlearticle := thing.Eq(iThing)

		// Get the next element containing additional info for this post
		scorenode := singlearticle.Next()
		if scorenode == nil {
			errhtml, _ := singlearticle.Html()
			post.Err = fmt.Errorf("Did not find next sibling for: %s\n", errhtml)
			continue
		}

		htmlpost := singlearticle.Find(".title a").First()
		if htmlpost.Size() == 0 {
			errhtml, _ := singlearticle.Html()
			post.Err = fmt.Errorf("Did not find title for: %s\n", errhtml)
			continue
		}

		post.Title = htmlpost.Text()
		var exists bool
		post.Url, exists = htmlpost.Attr("href")
		if exists == false {
			singlehtml, _ := htmlpost.Html()
			post.Err = fmt.Errorf("href not found in %s\n", singlehtml)
		}

		if DebugLevel > 3 {
			fmt.Printf("---------------------------\n")
			html, _ = scorenode.Html()
			fmt.Printf("HTML: %s\n", html)
			fmt.Printf("---------------------------\n")
		}

		// Get the score
		scoretag := scorenode.Find(".subtext .score").First()
		if scoretag.Size() == 0 {
			post.Err = fmt.Errorf("Did not find score for: %v\n", scorenode)
			continue
		}

		if DebugLevel > 3 {
			fmt.Printf("------- SCORE -------------\n")
			html, _ = scoretag.Html()
			fmt.Printf("HTML: %s\n", html)
			score := scoretag.Text()
			fmt.Printf("TEXT: %s\n", score)
			fmt.Printf("---------------------------\n")
		}

		post.SetScore(strings.Split(scoretag.Text(), " ")[0])

		postid, exists := scoretag.Attr("id")
		if !exists {
			html, _ = scoretag.Html()
			post.Err = fmt.Errorf("Did not find postid in %s\n", html)
		}

		if DebugLevel > 3 {
			fmt.Printf("------- POST ID -----------\n")
			fmt.Printf("TEXT: %s\n", postid)
			fmt.Printf("---------------------------\n")
		}

		post.PostId = strings.Split(postid, "_")[1]

		// Get the username and postdate
		hrefs := scorenode.Find(".subtext a")
		if hrefs.Size() == 0 {
			errhtml, _ := scorenode.Html()
			post.Err = fmt.Errorf("Did not find user and date in %s\n", errhtml)
			continue
		}

		for i := range hrefs.Nodes {
			href := hrefs.Eq(i)
			t, _ := href.Html()
			s, exists := href.Attr("href")
			if exists {
				if strings.HasPrefix(s, "user?id") {
					post.User = t
					continue
				}
				if strings.HasPrefix(s, "item?id") {
					if strings.Contains(t, "ago") {
						var postDate time.Time
						postDate, err = GetDateFromCreatedAgo(t)
						if err != nil {
							post.Err = errors.New(fmt.Sprintf("Failed to convert to date: %s: %s\n", t, err.Error()))
							continue
						}
						post.PostDate = postDate
						post.Err = err
					}
				}
			}
			if DebugLevel > 3 {
				fmt.Printf("------- HREF --------------\n")
				fmt.Printf("TEXT: %s\n", t)
				fmt.Printf("HREF: %s\n", s)
				fmt.Printf("---------------------------\n")
			}
		}
		if DebugLevel > 3 {
			fmt.Printf("---------------------------\n")
			fmt.Printf("POST: %s\n", post.String())
			fmt.Printf("---------------------------\n")
		}
	}

	return ps, err
}

func GetDateFromCreatedAgo(c string) (created time.Time, err error) {

	var amount int64
	var dateunit string
	created = time.Now()

	splitted := strings.Split(c, " ")
	if len(splitted) > 1 {
		amount, err = strconv.ParseInt(splitted[0], 10, 0)
		amount = amount * -1 // Back to the future
		if err != nil {
			err = errors.New(fmt.Sprintf("GetDateFromCreatedAgo: Failed to convert %s: ", c))
			return
		}
		dateunit = splitted[1]
		switch strings.ToLower(dateunit) {
		case "minutes", "minute":
			created = created.Add(time.Duration(amount) * time.Minute)
		case "hours", "hour":
			created = created.Add(time.Duration(amount) * time.Hour)
		case "days", "day":
			created = created.AddDate(0, 0, int(amount))
		case "months", "month":
			created = created.AddDate(0, int(amount), 0)
		case "years", "year":
			created = created.AddDate(int(amount), 0, 0)
		}
	}
	return
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
	err := HackerNewsPostScraper("newest")
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Failed to fetch from hackernews: ", err)
			panic(err)
		}
	}
}
