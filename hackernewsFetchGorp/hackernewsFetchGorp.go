package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	_ "github.com/lib/pq"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	geturl := "http://news.ycombinator.com/"
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

		if htmlpost.WebPostId == "" {
			if DebugLevel > 1 {
				fmt.Printf("WebPostId not set in %s\n", htmlpost.Title)
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
			" where WebPostId = :post_id"
		_, err := dbmap.Select(&intSelectResult, postcountsql, map[string]interface{}{
			"post_id": htmlpost.WebPostId,
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
			getpostsql := "select * from " + dbmap.Dialect.QuoteField(tablename) + " where WebPostId = :post_id"
			_, err := dbmap.Select(&dbposts, getpostsql, map[string]interface{}{
				"post_id": htmlpost.WebPostId,
			})
			if err != nil {
				return errors.New(fmt.Sprintf("Getting WebPostId %s from DB failed: %s\n", htmlpost.WebPostId, err.Error()))
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

	root, err := html.Parse(body)
	if err != nil {
		err = errors.New("Failed to html.Parse: " + err.Error())
		return
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Tr && n.Parent != nil && n.Parent.DataAtom == atom.Tbody {
			matched := scrape.Attr(n, "class") == "athing"
			return matched
		}
		return false
	}

	// grab all articles and loop over them
	articles := scrape.FindAll(root, matcher)
	for _, article := range articles {
		var ok bool
		// Get one post entry
		var titlenode *html.Node

		titlenode, ok = scrape.Find(article,
			func(n *html.Node) bool {
				if n.DataAtom == atom.A && n.Parent != nil && scrape.Attr(n.Parent, "class") == "title" {
					return true
				}
				return false
			})
		if !ok {
			continue
		}
		// Create a new post struct - if the crawling fails the post will have an Err attached
		// but will be added to the outgoing (psout) slice nevertheless
		post := post.NewPost()
		post.Site = "hackernews"

		post.Title = scrape.Text(titlenode)
		post.Url = scrape.Attr(titlenode, "href")
		if strings.HasPrefix(post.Url, "item?id=") {
			post.Url = "https://news.ycombinator.com/" + post.Url
		}

		ps = append(ps, &post)

		// Get additional info for this post
		scorenode := article.NextSibling
		if scorenode == nil {
			post.Err = errors.New("Did not find score for: %s\n" + scrape.Text(article))
			continue
		}

		// Get the subtext containing scores, user and date
		subtext, ok := scrape.Find(scorenode,
			func(n *html.Node) bool {
				if scrape.Attr(n, "class") == "subtext" {
					return true
				}
				return false
			})
		if !ok {
			post.Err = errors.New(fmt.Sprintf("Did not find siblings for subtext %s\n", scorenode.Data))
			continue
		}

		subs := scrape.FindAll(subtext,
			func(n *html.Node) bool {
				// Get the PostId and Score
				// span class="score" id="score_9643579">92 points</span>
				if n.DataAtom == atom.Span && scrape.Attr(n, "class") == "score" && n.Parent != nil && scrape.Attr(n.Parent, "class") == "subtext" {

					// Get score
					var scoreid int
					scorestr := strings.Split(scrape.Text(n), " ")[0]
					scoreid, err = strconv.Atoi(scorestr)
					if err != nil {
						fmt.Printf("Failed to convert score %s to int: %s\n", scorestr, err.Error())
						return false
					}
					post.Score = scoreid

					// Get PostId
					postidstr := scrape.Attr(n, "id")
					if len(strings.Split(postidstr, "_")) > 1 {
						post.WebPostId = strings.Split(postidstr, "_")[1]
						return true
					}
				}
				// Get the Username and Creation Date for this post
				if scrape.Attr(n.Parent, "class") == "subtext" && n.DataAtom == atom.A && n.Parent != nil {
					href := strings.ToLower(scrape.Attr(n, "href"))
					if href != "" {
						s := strings.Split(href, "?")
						if s[0] == "user" && len(s) > 1 {
							// Username
							u := strings.Split(s[1], "=")
							if len(u) > 1 {
								post.User = u[1]
								return true
							}
						} else {
							if s[0] == "item" && len(s) > 1 {
								// Created date
								createdago := scrape.Text(n)
								if strings.Contains(createdago, "ago") {
									var postDate time.Time

									postDate, err = GetDateFromCreatedAgo(createdago)
									if err != nil {
										err = errors.New(fmt.Sprintf("Failed to convert to date: %V\n", createdago))
										return false
									}
									post.PostDate = postDate

									return true
								}
							}
						}
					}
				} // end "class" == "subtext"
				return false
			})

		if len(subs) == 0 {
			var w bytes.Buffer
			if rerr := html.Render(&w, subtext); rerr != nil {
				fmt.Printf("Render error: %s\n", rerr)
			}
			post.Err = errors.New(fmt.Sprintf("Unable to parse score,user,date from %s:\n %s\n", post.Title, w.String()))
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
			created = created.AddDate(0, 0, int(amount*-1))
		case "months", "month":
			created = created.AddDate(0, int(amount*-1), 0)
		case "years", "year":
			created = created.AddDate(int(amount*-1), 0, 0)
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
	err := HackerNewsPostScraper("golang")
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Failed to fetch from sub hackernews golang: ", err)
			panic(err)
		}
	}
}
