package main

import (
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
	"log"
	"net/http"
	"unicode"
)

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 3

func HackerNewsPostScraper(sub string) (err error) {
	drivername := "postgres"
	dsn := "user=golang password=golang dbname=golang sslmode=disable"
	dialect := gorp.PostgresDialect{}

	//drivername := "mysql"
	//dsn := "golang:golang@/golang?parseTime=true"
	//dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data:
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
	defer body.Close()

	// Create a new post slice and then parse the response body into ps
	ps := make([]post.Post, 0)
	ps, err = ParseHtmlHackerNews(body, ps)
	if err != nil {
		return errors.New("ParseHtmlHackerNews: " + err.Error())
	}
	foundnewposts := false
	updatedposts := 0

	// insert rows - auto increment PKs will be set properly after the insert
	for _, htmlpost := range ps {
		if htmlpost.Err == nil {
			var postcount int

			// Store reddit sub
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
			postcount = intSelectResult[0]

			// DEBUG
			if DebugLevel > 3 {
				fmt.Println("HTMLpost.PostId: " + htmlpost.PostId)
				fmt.Printf("HTMLpost.Id: %v\n", htmlpost.Id)
				fmt.Printf("DBpost count: %v \n", postcount)
			}

			// New post? then insert
			if postcount == 0 {
				foundnewposts = true
				err = dbmap.Insert(&htmlpost)

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
			} else {
				// Post already exists, do an update
				dbposts := make([]post.Post, 0)
				getpostsql := "select * from " + dbmap.Dialect.QuoteField(tablename) + " where PostId = :post_id"
				_, err := dbmap.Select(&dbposts, getpostsql, map[string]interface{}{
					"post_id": htmlpost.PostId,
				})
				if err != nil {
					return errors.New(fmt.Sprintf("Getting PostId %s from DB failes\n", htmlpost.PostId, err.Error()))
				}
				var dbpost post.Post
				if len(dbposts) > 0 {
					dbpost = dbposts[0]
				} else {
					return errors.New(fmt.Sprintf("Query: %s returned no result\n", getpostsql))
				}
				// DEBUG
				if DebugLevel > 3 {
					fmt.Printf("DBPOST: %s\n", dbpost.String())
					fmt.Printf("DBpost.Id: %v\n", dbpost.Id)
					fmt.Printf("DBpost.Score: %v\n", dbpost.Score)
				}

				if htmlpost.Score != dbpost.Score {

					if DebugLevel > 2 {
						// Print out the update info
						fmt.Println("----------- UPDATE POST START -----------------")
						fmt.Println("Title: " + dbpost.Title)
						fmt.Printf("Id: %v\n", dbpost.Id)
						fmt.Printf("From score %d to score %d\n", dbpost.Score, htmlpost.Score)
						fmt.Println("----------- UPDATE POST END -------------------")
					}

					dbpost.Score = htmlpost.Score
					affectedrows, err := dbmap.Update(&dbpost)
					switch {
					case err != nil:
						return errors.New("update table " + tablename + " failed: " + err.Error())
					case affectedrows == 0:
						return errors.New(fmt.Sprintf("update table %s for Id %d did not affect any lines", tablename, dbpost.Id))
					default:
						updatedposts++
					}
				}
			}
		} else {
			if DebugLevel > 1 {
				fmt.Println("Single post error in " + geturl + ": " + htmlpost.Err.Error())
			}
		}
	}
	if !foundnewposts {
		if DebugLevel > 2 {
			fmt.Println("No new posts found at " + geturl)
		}
	}

	if updatedposts > 0 {
		if DebugLevel > 2 {
			fmt.Printf("%d posts have been updated from %s\n", updatedposts, geturl)
		}
	}

	return
}

func GetHtmlBody(url string) (body io.ReadCloser, err error) {

	// Get data from url
	resp, err := http.Get(url)
	if err != nil {
		err = errors.New("Failed to http.Get from " + url + ": " + err.Error())
		return
	}
	if resp != nil {
		if resp.Body == nil {
			err = errors.New("Body from " + url + " is nil!")
			return
		} else {
			//defer resp.Body.Close()
		}
	} else {
		err = errors.New("Response from " + url + " is nil")
		return
	}
	if resp.StatusCode != 200 { // 200 = OK
		httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", url, resp.StatusCode, resp.Status)
		err = errors.New(httperr)
		return
	}

	body = resp.Body
	return
}

// Parse for posts in html from reddit, input html is an io.Reader and returns recognized posts in a psout slice of posts.
// Errors which affect only a single post are stored in their post.Err
func ParseHtmlHackerNews(body io.Reader, ps []post.Post) (psout []post.Post, err error) {

	// Create a new post struct - if the crawling fails the post will have an Err attached
	// but will be added to the outgoing (psout) slice nevertheless
	post := post.NewPost()
	post.Site = "hackernews"

	root, err := html.Parse(body)
	if err != nil {
		err = errors.New("Failed to html.Parse: " + err.Error())
		return
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		//if n.DataAtom == atom.A && n.Parent != nil && n.Parent.Parent != nil {
		//	return scrape.Attr(n.Parent.Parent, "class") == "athing"
		//}

		//if n.DataAtom == atom.Tr && n.Parent != nil && n.Parent.DataAtom == atom.Table {
		//	return scrape.Attr(n, "class") == "athing"
		//}

		if n.DataAtom == atom.Tr {
			return scrape.Attr(n, "class") == "athing"
			//return true
		}

		return false
	}
	// grab all articles and print them
	articles := scrape.FindAll(root, matcher)
	for i, article := range articles {
		fmt.Printf("--INDEX %2d\n", i)

		titlenode, ok := scrape.Find(article,
			func(n *html.Node) bool {
				if n.DataAtom == atom.A && n.Parent != nil && scrape.Attr(n.Parent, "class") == "title" {
					return true
				}
				return false
			})
		if ok {
			fmt.Printf("---TITLE %s\n", scrape.Text(titlenode))
			fmt.Printf("---URL %s\n", scrape.Attr(titlenode, "href"))
		}

		fmt.Printf("---NEXT SIBLING %s\n", scrape.Text(article.NextSibling))

		post.Title = scrape.Text(article)
		post.Url = scrape.Attr(article, "href")
	}

	ps = append(ps, post)

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
	err := HackerNewsPostScraper("golang")
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Failed to fetch from sub hackernews golang: ", err)
			panic(err)
		}
	}
}
