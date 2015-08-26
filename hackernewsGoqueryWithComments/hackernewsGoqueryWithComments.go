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
	"net/url"
	//"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 1

func HackerNewsPostScraper(sub string) (err error) {
	//drivername := "postgres"
	//dsn := "user=golang password=golang dbname=golang sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	drivername := "mysql"
	dsn := "golang:golang@/golang?parseTime=true"
	dialect := gorp.MySQLDialect{"InnoDB", "utf8mb4"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data using ping
	if err = db.Ping(); err != nil {
		return errors.New("db.Ping failed: " + err.Error())
	}

	// Set the connection to use utf8bmb4
	if dialect.Engine == "InnoDB" {
		fmt.Println("Setting connection to utf8mb4")
		_, err = db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			return errors.New("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci: " + err.Error())
		}
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	defer dbmap.Db.Close()
	dbmap.DebugLevel = DebugLevel
	// Will log all SQL statements + args as they are run
	// The first arg is a string prefix to prepend to all log messages
	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "Trace:", log.Lmicroseconds))

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name

	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	table := dbmap.AddTableWithName(post.Post{}, "posts_index_test")
	table.SetKeys(true, "PID")

	// Add the comments table
	table = dbmap.AddTableWithName(post.Comment{}, "comments_index_test")
	table.SetKeys(true, "Id")

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
	// DEBUG for a special thread
	//geturl := "https://news.ycombinator.com/item?id=10056146"
	body, err := GetHtmlBody(geturl)
	if err != nil {
		return errors.New("GetHtmlBody: " + err.Error())
	}

	// Create a new post slice and then parse the response body into ps
	ps := make([]*post.Post, 0)
	//cs := make([]*post.Comment, 0)
	ps, err = ParseHtmlHackerNews(body, ps)
	if err != nil {
		return errors.New("ParseHtmlHackerNews: " + err.Error())
	}

	// Number of updated posts
	var updatedPostsCount int64
	// Number of new posts
	var insertedPostsCount int64

	var insertedPostsCommentCount int64
	var updatedPostsCommentCount int64

	// Number of post parsing errors
	var htmlPostErrorCount uint32
	// Number of comment parsing errors
	var htmlCommentErrorCount uint32

	// loop over all parsed posts
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
			htmlPostErrorCount++
			continue
		}

		if len(htmlpost.CommentParseErrors) > 0 {
			for _, c := range htmlpost.CommentParseErrors {
				htmlCommentErrorCount++
				if DebugLevel > 2 {
					fmt.Println("Single comment error in '" + geturl + "' for WebPostId '" + htmlpost.WebPostId + ": " + c.Err.Error())
				}

			}
		}

		// Store post sub
		htmlpost.PostSub = sub

		tm, err := dbmap.TableFor(reflect.TypeOf(*htmlpost), true)
		if err != nil {
			return errors.New("Failed to get reflection type: " + err.Error())
		}
		if DebugLevel > 3 {
			fmt.Println("TABLEMAP: " + tm.TableName)
		}
		// check if post already exists
		dbposts := make([]post.Post, 0)
		getpostsql := "select * from " + dbmap.Dialect.QuotedTableForQuery("", tm.TableName) + " where WebPostId = :post_id"
		_, err = dbmap.Select(&dbposts, getpostsql, map[string]interface{}{
			"post_id": htmlpost.WebPostId,
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Getting PostId %s from DB failed: %s", htmlpost.WebPostId, err.Error()))
		}
		var dbpost *post.Post
		if len(dbposts) == 1 {
			dbpost = &dbposts[0]
		} else if len(dbposts) > 1 {
			return errors.New(fmt.Sprintf("Query: %s returned %d rows", getpostsql, len(dbposts)))
		}
		postcount := len(dbposts)

		// New post? then insert
		if postcount == 0 {

			if DebugLevel > 2 {
				fmt.Printf("New post found, inserting htmlpost.WebPostId '%s'\n", htmlpost.WebPostId)
			}

			// Reset the rowcount info
			dbmap.LastOpInfo.Reset()
			htmlpost.CommentCount = uint64(len(htmlpost.Comments))
			// Insert the new post into the database
			err = dbmap.InsertWithChilds(htmlpost)

			if DebugLevel > 2 {
				// Print out the crawled info
				fmt.Println("----------- INSERT POST START -----------------")
				fmt.Println(htmlpost.String("INSERT: "))
			}
			if err != nil {
				return errors.New("insert into table " + dbmap.Dialect.QuoteField(tm.TableName) + " failed: " + err.Error())
			}
			if DebugLevel > 2 {
				// Print out the end of the crawled info
				fmt.Println("----------- INSERT POST END -------------------")
			}
			insertedPostsCount += dbmap.LastOpInfo.RowCount
			insertedPostsCommentCount += dbmap.LastOpInfo.ChildInsertRowCount

		} else {
			// Post already exists, get the full post with its comments from the db

			res, err := dbmap.GetWithChilds(post.Post{}, 9999999999, 0, dbpost.Id)
			if err != nil {
				return errors.New("get failed: " + err.Error())
			}
			if res == nil {
				return errors.New(fmt.Sprintf("Get post for id %d did not return any rows ", dbpost.Id))
			}
			dbpost = res.(*post.Post)

			// Check if an update is needed
			var updateNeeded bool
			updateNeeded, err = AddUpdatableChilds(htmlpost, dbpost, dbmap)
			if err != nil {
				return errors.New(fmt.Sprintf("CheckIfDataChanged for post '%s' failed: %s", htmlpost.WebPostId, err.Error()))
			}
			//if htmlpost.Score != dbpost.Score {
			if updateNeeded {
				// The post changed, do an update into the database

				//fmt.Println("Post Date db: " + dbpost.PostDate.String() + ", html: " + htmlpost.PostDate.String())
				//fmt.Printf("Post Score db: %d, html: %d\n", dbpost.Score, htmlpost.Score)

				if DebugLevel > 2 {
					fmt.Println("----------- UPDATE POST START -----------------")
					fmt.Println(dbpost.String("UPDATE1: "))
					fmt.Printf("From score %d to score %d\n", dbpost.Score, htmlpost.Score)
				}
				dbpost.Score = htmlpost.Score
				dbpost.PostDate = htmlpost.PostDate

				// Reset the rowcount info
				dbmap.LastOpInfo.Reset()

				// Update the posts together with its comments
				affectedrows, err := dbmap.UpdateWithChilds(dbpost)

				switch {
				case err != nil:
					return errors.New("update table " + tm.TableName + " failed: " + err.Error())
				case affectedrows == 0:
					return errors.New(fmt.Sprintf("update table %s for Id %d did not affect any lines", tm.TableName, dbpost.Id))
				default:

					updatedPostsCount += dbmap.LastOpInfo.RowCount
					insertedPostsCommentCount += dbmap.LastOpInfo.ChildInsertRowCount
					updatedPostsCommentCount += dbmap.LastOpInfo.ChildUpdateRowCount

					dbpost.CommentCount += uint64(dbmap.LastOpInfo.ChildInsertRowCount)
					_, err = dbmap.Update(dbpost)

					if err != nil {
						return errors.New(fmt.Sprintf("Update for post '%s' failed: %s", dbpost.WebPostId, err.Error()))
					}

					if DebugLevel > 2 {
						// Print out the update info
						fmt.Println("----------- UPDATE POST COMMIT -----------------")
						fmt.Println(dbpost.String("UPDATE2: "))
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
		return
	}

	if DebugLevel > 2 {
		fmt.Printf("%d new posts have been inserted from %s\n", insertedPostsCount, geturl)
		fmt.Printf("%d posts have been updated from %s\n", updatedPostsCount, geturl)
		fmt.Printf("%d new comments have been inserted from %s\n", insertedPostsCommentCount, geturl)
		fmt.Printf("%d comments have been updated from %s\n", updatedPostsCommentCount, geturl)
		fmt.Printf("%d comment errors\n", htmlCommentErrorCount)

	}

	return
}

func AddUpdatableChilds(htmlpost *post.Post, dbpost *post.Post, dbmap *gorp.DbMap) (updateNeeded bool, err error) {
	// Check if there are aleady comments in dbpost
	// If not get them from the database

	if len(dbpost.Comments) == 0 {
		pk := dbpost.Id
		if pk == 0 {
			err = errors.New("primary key not set in dbpost")
			return
		}
		var res interface{}
		res, err = dbmap.GetWithChilds(post.Post{}, 9999999999, 0, pk)
		if err != nil {
			err = errors.New("get failed: " + err.Error())
			return
		}
		if res == nil {
			err = errors.New(fmt.Sprintf("Get post for id %d did not return any rows ", pk))
			return
		}

		dbpost := res.(*post.Post)
		if DebugLevel > 3 {
			// Print out the update info
			fmt.Println("----------- DB POST -----------------")
			fmt.Println(dbpost.String("CHECK DB: "))
			fmt.Println("----------- DB POST END -------------------")
		}
	}
	if DebugLevel > 3 {
		// Print out the update info
		fmt.Println("----------- HTML POST -----------------")
		fmt.Println(htmlpost.String("CHECK HTML: "))
		fmt.Println("----------- HTML POST END -------------------")
	}

	updateNeeded = htmlpost.Hash() != dbpost.Hash()

	if updateNeeded {
		var UpdatedComments []*post.Comment
		var found bool

		if DebugLevel > 2 {
			fmt.Printf("**** UpdatedComments len %d\n", len(UpdatedComments))
		}
		for _, h := range htmlpost.Comments {
			found = false
			htmlHash := h.Hash()
			for _, d := range dbpost.Comments {
				if DebugLevel > 2 {
					fmt.Printf("**** COMPARE\n")
					fmt.Printf("**** **** d.Hash():%d htmlHash %d\n", d.Hash(), htmlHash)
					fmt.Printf("**** **** d.Date '%s' h.Date '%s'\n", d.GetCommentDate().String(), h.GetCommentDate().String())
					fmt.Printf("**** COMPARE END\n")
				}
				if d.Hash() == htmlHash {
					// post with identical content has been found - do not store this comment
					found = true
					if DebugLevel > 2 {
						fmt.Printf("**** ***************** MATCH d.Hash() == htmlHash %d\n", d.Hash())
					}
					break
				}
				if h.WebCommentId == d.WebCommentId {
					// external unique comment id found - this comment is already stored
					// but the comment content has been changed - update needed
					if DebugLevel > 3 {
						fmt.Printf("**** COMPARE h.WebCommentId\n")
						fmt.Printf("**** **** h '%s' d '%s'\n", h.WebCommentId, d.WebCommentId)
						fmt.Printf("**** COMPARE h.WebCommentId END\n")
					}
					h.Id = d.Id
					h.PostId = d.PostId
					h.Title = d.Title
					h.Body = d.Body
					break
				}
			}
			if !found {
				UpdatedComments = append(UpdatedComments, h)
				if DebugLevel > 2 {
					fmt.Printf("**** UpdatedComments len %d\n", len(UpdatedComments))
					fmt.Printf("**** **** append(UpdatedComments, h) %s\n", h.String("APP: "))
				}
			}

		}
		fmt.Printf("**** htmlpost.Comments len %d\n", len(htmlpost.Comments))
		fmt.Printf("**** UpdatedComments len %d\n", len(UpdatedComments))
		dbpost.Comments = make([]*post.Comment, len(UpdatedComments), len(UpdatedComments))
		fmt.Printf("**** dbpost.Comments1 len %d\n", len(dbpost.Comments))

		copy(dbpost.Comments, UpdatedComments)

		fmt.Printf("**** dbpost.Comments2 len %d\n", len(dbpost.Comments))
	}
	if (DebugLevel > 3) && updateNeeded {
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

func ParseHtmlComments(p *post.Post) (err error) {

	if p.WebPostId == "" {
		return errors.New(fmt.Sprintf("p.WebPostId is empty in post '%s'", p.String("PC: ")))
	}
	// Get comments from hackernews
	geturl := fmt.Sprintf("http://news.ycombinator.com/item?id=%s", p.WebPostId)
	// DEBUG
	//geturl := fmt.Sprintf("https://news.ycombinator.com/item?id=9751858")

	if DebugLevel > 2 {
		fmt.Printf("START GET COMMENTS FROM '%s'\n", geturl)
	}

	body, err := GetHtmlBody(geturl)
	if err != nil {
		return errors.New("GetHtmlBody: " + err.Error())
	}
	// Create a qoquery document to parse from an io.Reader
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return errors.New("Failed to parse HTML: " + err.Error())
	}
	// Find hackernews comments = elements with class "athing"
	thing := doc.Find(".athing")
	for iThing := range thing.Nodes {
		// use `singlecomment` as a selection of one single post
		singlecomment := thing.Eq(iThing)

		comment := post.NewComment()
		//p.Comments = append(p.Comments, &comment)

		comheads := singlecomment.Find(".comhead a")
		for i := range comheads.Nodes {

			comhead := comheads.Eq(i)
			t, _ := comhead.Html()
			s, exists := comhead.Attr("href")
			if exists {
				if strings.HasPrefix(s, "user?id") {
					comment.User = t
					continue
				}
				if strings.HasPrefix(s, "item?id") {
					if strings.Contains(t, "ago") {
						var commentDate time.Time
						commentDate, err = GetDateFromCreatedAgo(t)
						if err != nil {
							comment.Err = errors.New(fmt.Sprintf("Failed to convert to date: %s: %s\n", t, err.Error()))
							err = nil
							continue
						}
						comment.CommentDate = commentDate
						if len(strings.Split(s, "=")) > 1 {
							comment.WebCommentId = strings.Split(s, "=")[1]
						}
						//comment.Err = err
					}
				}
			}

			comments := singlecomment.Find("span.comment")

			removeReplySelection := comments.Find("span div.reply")
			removeReplySelection.Remove()

			var sep string
			for iComment, _ := range comments.Nodes {
				s := comments.Eq(iComment)

				h, _ := s.Html()

				if !utf8.ValidString(s.Text()) {
					comment.Err = errors.New(fmt.Sprintf("Ignoring invalid UTF-8: '%s'", s.Text()))
					break
				}

				h, err = HtmlToMarkdown(h)
				if err != nil {
					comment.Err = errors.New(fmt.Sprintf("Ignoring markdownifier: '%s'", err.Error()))
					break
				}

				if h != "" {
					comment.Body = comment.Body + sep + h
				}
				sep = "\n"
			}
			//fmt.Printf("POST %s BODY = %s\n", p.WebPostId, comment.Body)

			if comment.Err == nil && len(comment.WebCommentId) > 0 && len(comment.Body) > 0 {
				p.Comments = append(p.Comments, &comment)
			} else {
				p.CommentParseErrors = append(p.CommentParseErrors, &comment)
			}
		}
	}

	if DebugLevel > 0 {
		fmt.Printf("GET COMMENTS FROM '%s' yielded %d comments\n", geturl, len(p.Comments))
	}

	return err
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
		post.Site = "hackernews"
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
		post.Url = stringMinifier(post.Url)

		if !(strings.HasPrefix(post.Url, "http")) {
			post.Url = "https://news.ycombinator.com/" + post.Url
		}

		if DebugLevel > 2 {
			fmt.Printf("**** URL post.Url: %s\n", post.Url)
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

		post.WebPostId = strings.Split(postid, "_")[1]

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
		if post.Err == nil {
			err = ParseHtmlComments(&post)
		}
		if DebugLevel > 2 && err == nil {
			fmt.Printf("------ POST DUMP -----------\n")
			fmt.Print(post.String("PARSED: "))
			fmt.Printf("------ POST DUMP END -------\n")
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

	// Convert time to UTC
	created = created.UTC()
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

// Markdownifier using the wonderfull online api service at http://heckyesmarkdown.com/
// This example converts html provided as a string to markdown, the html does not need to be complete,
// just a part/snippet is enough
// Thanks to Brett for providing this service to the world!
func HtmlToMarkdown(htmlInput string) (markdownResult string, err error) {

	// Get starttime for measuring how long this functions takes
	timeStart := time.Now()

	fmt.Printf("Start HtmlToMarkdown - ")

	serviceEndPoint := "http://heckyesmarkdown.com/go/"
	postParams := url.Values{}
	postParams.Set("html", htmlInput) // the html input string
	postParams.Set("read", "0")       // turn readability off, default is on
	postParams.Set("md", "1")         // Run Markdownify, default on

	timeout := time.Duration(30 * time.Second)
	client := &http.Client{}
	client.Timeout = timeout

	resp, err := client.PostForm(serviceEndPoint, postParams)
	if err != nil {
		requestDuration := (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))
		fmt.Printf("HtmlToMarkdown ERROR %s, duration %d\n", err.Error(), requestDuration)
		return
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	markdownResult = string(body)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", resp.Status)
	}

	requestDuration := (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))
	fmt.Printf("HtmlToMarkdown duration %d\n", requestDuration)
	return
}

func main() {
	err := HackerNewsPostScraper("newest")
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Failed to fetch from hackernews newest: ", err)
			panic(err)
		}
	}
	err = HackerNewsPostScraper("")
	if err != nil {
		if DebugLevel > 0 {
			log.Fatalln("Failed to fetch from hackernews: ", err)
			panic(err)
		}
	}
}
