package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	"io/ioutil"
	"net/http"
	//"net/url"
	//"os"
	"bytes"
	"github.com/jeffail/gabs"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

type RedditJsonCommentList []struct {
	Data struct {
		After    interface{} `json:"after"`
		Before   interface{} `json:"before"`
		Children []struct {
			Data struct {
				ApprovedBy          interface{} `json:"approved_by"`
				Archived            bool        `json:"archived"`
				Author              string      `json:"author"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				BannedBy            interface{} `json:"banned_by"`
				Clicked             bool        `json:"clicked"`
				Created             float64     `json:"created"`
				CreatedUtc          float64     `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Domain              string      `json:"domain"`
				Downs               int         `json:"downs"`
				Edited              float64     `json:"edited"`
				From                interface{} `json:"from"`
				FromID              interface{} `json:"from_id"`
				FromKind            interface{} `json:"from_kind"`
				Gilded              int         `json:"gilded"`
				Hidden              bool        `json:"hidden"`
				HideScore           bool        `json:"hide_score"`
				ID                  string      `json:"id"`
				IsSelf              bool        `json:"is_self"`
				Likes               interface{} `json:"likes"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				LinkFlairText       interface{} `json:"link_flair_text"`
				Locked              bool        `json:"locked"`
				Media               struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"media"`
				MediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"media_embed"`
				ModReports    []interface{} `json:"mod_reports"`
				Name          string        `json:"name"`
				NumComments   int           `json:"num_comments"`
				NumReports    interface{}   `json:"num_reports"`
				Over18        bool          `json:"over_18"`
				Permalink     string        `json:"permalink"`
				Quarantine    bool          `json:"quarantine"`
				RemovalReason interface{}   `json:"removal_reason"`
				ReportReasons interface{}   `json:"report_reasons"`
				Saved         bool          `json:"saved"`
				Score         int           `json:"score"`
				SecureMedia   struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"secure_media"`
				SecureMediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"secure_media_embed"`
				Selftext      string        `json:"selftext"`
				SelftextHTML  interface{}   `json:"selftext_html"`
				Stickied      bool          `json:"stickied"`
				Subreddit     string        `json:"subreddit"`
				SubredditID   string        `json:"subreddit_id"`
				SuggestedSort interface{}   `json:"suggested_sort"`
				Thumbnail     string        `json:"thumbnail"`
				Title         string        `json:"title"`
				Ups           int           `json:"ups"`
				URL           string        `json:"url"`
				UserReports   []interface{} `json:"user_reports"`
				Visited       bool          `json:"visited"`
			} `json:"data"`
			Kind string `json:"kind"`
		} `json:"children"`
		Modhash string `json:"modhash"`
	} `json:"data"`
	Kind string `json:"kind"`
}

type RedditJsonPostList struct {
	Data struct {
		After    interface{} `json:"after"`
		Before   interface{} `json:"before"`
		Children []struct {
			Data struct {
				ApprovedBy          interface{} `json:"approved_by"`
				Archived            bool        `json:"archived"`
				Author              string      `json:"author"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				BannedBy            interface{} `json:"banned_by"`
				Clicked             bool        `json:"clicked"`
				Created             float64     `json:"created"`
				CreatedUtc          float64     `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Domain              string      `json:"domain"`
				Downs               int         `json:"downs"`
				Edited              float64     `json:"edited"`
				From                interface{} `json:"from"`
				FromID              interface{} `json:"from_id"`
				FromKind            interface{} `json:"from_kind"`
				Gilded              int         `json:"gilded"`
				Hidden              bool        `json:"hidden"`
				HideScore           bool        `json:"hide_score"`
				ID                  string      `json:"id"`
				IsSelf              bool        `json:"is_self"`
				Likes               interface{} `json:"likes"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				LinkFlairText       interface{} `json:"link_flair_text"`
				Locked              bool        `json:"locked"`
				Media               struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"media"`
				MediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"media_embed"`
				ModReports    []interface{} `json:"mod_reports"`
				Name          string        `json:"name"`
				NumComments   int           `json:"num_comments"`
				NumReports    interface{}   `json:"num_reports"`
				Over18        bool          `json:"over_18"`
				Permalink     string        `json:"permalink"`
				Quarantine    bool          `json:"quarantine"`
				RemovalReason interface{}   `json:"removal_reason"`
				ReportReasons interface{}   `json:"report_reasons"`
				Saved         bool          `json:"saved"`
				Score         int           `json:"score"`
				SecureMedia   struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"secure_media"`
				SecureMediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"secure_media_embed"`
				Selftext      string        `json:"selftext"`
				SelftextHTML  interface{}   `json:"selftext_html"`
				Stickied      bool          `json:"stickied"`
				Subreddit     string        `json:"subreddit"`
				SubredditID   string        `json:"subreddit_id"`
				SuggestedSort interface{}   `json:"suggested_sort"`
				Thumbnail     string        `json:"thumbnail"`
				Title         string        `json:"title"`
				Ups           int           `json:"ups"`
				URL           string        `json:"url"`
				UserReports   []interface{} `json:"user_reports"`
				Visited       bool          `json:"visited"`
			} `json:"data"`
			Kind string `json:"kind"`
		} `json:"children"`
		Modhash string `json:"modhash"`
	} `json:"data"`
	Kind string `json:"kind"`
}

type UnmarshallBuffer struct {
	object interface{}
}

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 1

var debugFromFile bool = true

// Connection to the database, gets initialized by InitDatabase
var dbmap *gorp.DbMap

func ParseJsonComments(buf []byte, post *post.Post) (err error) {

	// Remove BOM
	buf = bytes.TrimPrefix(buf, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}

	if !utf8.ValidString(string(buf)) {
		fmt.Printf("INVALID UTF8: %s\n", string(buf))
		return
	}

	//jsonParsed, err := gabs.ParseJSON([]byte(`{"a":1}`))
	jsonParsed, err := gabs.ParseJSON(buf)
	//jsonParsed, err := gabs.ParseJSONFile("testcomments.json.txt")
	if err != nil {
		fmt.Println("-----------------------")
		fmt.Printf("Failed to parse json comments: %s\n", err.Error())
		fmt.Printf("%s", post.String("X "))
		fmt.Println("-----------------------")
		return err
	}
	//fmt.Printf("%s\n", jsonParsed.StringIndent("", " "))

	// S is shorthand for Search
	//children, _ := jsonParsed.Path("data.children").Children()
	rootnodes, _ := jsonParsed.Children()

	for _, rootnode := range rootnodes {

		//fmt.Printf("%s\n", child.String())
		//fmt.Printf("%s\n", rootnode.Path("kind").String())

		if rootnode.Path("kind").String() == "\"Listing\"" {
			fmt.Printf("%s\n", rootnode.Path("kind").String())
			err = ParseCommentKindListing(0, rootnode, post)

		}
		//TraceRedditCommentJson(child, 1)

	}
	return err
}

func ParseCommentKindListing(level int, listing *gabs.Container, post *post.Post) (err error) {

	childs := listing.Search("data")

	//fmt.Printf("ParseListing %s\n", childs.String())

	childCount, err := childs.ArrayCount("children")
	if err != nil {
		fmt.Printf("Failed to get child count: %s\n", err.Error())
		return err
	}
	fmt.Println(strings.Repeat(" ", level*3) + fmt.Sprintf("Found %d children", childCount))

	for i := 0; i < childCount; i++ {
		child, _ := childs.ArrayElement(i, "children")
		//fmt.Printf("level %d author %d: %s\n", level, i, child.Path("data.author").String())
		//fmt.Printf("%s\n", child.String())

		fmt.Println(strings.Repeat(" ", level*3) + "Kind: " + child.Path("kind").String())
		fmt.Println(strings.Repeat(" ", level*3) + "Name: " + child.Path("data.name").String())
		fmt.Println(strings.Repeat(" ", level*3) + "subreddit_id: " + child.Path("data.subreddit_id").String())
		fmt.Println(strings.Repeat(" ", level*3) + "author: " + child.Path("data.author").String())
		if child.Path("kind").String() == "\"t3\"" {
			fmt.Println(strings.Repeat(" ", level*3) + "body: " + child.Path("data.selftext").String())

		} else {
			fmt.Println(strings.Repeat(" ", level*3) + "body: " + child.Path("data.body").String())
		}
		//fmt.Printf("SubPath %s\n", child.Path("data.replies").String())
		if child.Path("data.replies.kind").String() == "\"Listing\"" {
			err = ParseCommentKindListing(level+1, child.Path("data.replies"), post)

		}
	}

	return err
}

func main() {

	/*
		buf, err := ioutil.ReadFile("testcomments.json.txt")
		if err != nil {
			fmt.Printf("Error reading json test file: %s\n", err.Error())
			return
		}
		TraceJosonListing(buf)
		return
	*/
	var err error
	dbmap, err = InitDatabase()
	defer dbmap.Db.Close()
	if err != nil {
		fmt.Printf("Failed to init database: %s\n", err.Error())
		return
	}

	var postsub string
	postsub = "golang"

	//uri := "https://www.reddit.com/r/golang/controversial.json"
	uri := "https://www.reddit.com/r/" + postsub + ".json"
	fmt.Println("fetching", uri)
	redditPostList, err := GetJsonPostList(uri)
	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		fmt.Println(err)
		return
	}

	fmt.Printf("Children len: %d\n", len(redditPostList.Data.Children))

	// Create a new post slice to be stored in the database later
	PostList := make([]*post.Post, 0)

	// Loop over posts and get the comments
	for index, child := range redditPostList.Data.Children {
		fmt.Printf("%d, Title: %s, ID: %s\n", index, child.Data.Title, child.Data.ID)
		//GetJsonCommentList(child.Data.ID)

		// Create a new post struct - if the crawling fails the post will have an Err attached
		// but will be added to the outgoing (psout) slice nevertheless
		post := post.NewPost()

		post.Title = child.Data.Title
		post.WebPostId = child.Data.ID
		post.Url = child.Data.URL
		post.User = child.Data.Author
		post.Score = child.Data.Score
		post.Body = child.Data.Selftext
		post.PostDate = time.Unix(int64(child.Data.CreatedUtc), 0)
		post.PostSub = postsub
		post.CommentCount = uint64(child.Data.NumComments)
		post.Site = "reddit"
		// Add to the crawled post list
		PostList = append(PostList, &post)
	}

	tm, err := dbmap.TableFor(reflect.TypeOf(post.NewPost()), true)
	if err != nil {
		fmt.Println("Failed to get reflection type: " + err.Error())
		return
	}

	// loop over all parsed posts
	for _, parsedpost := range PostList {

		// check if post already exists
		dbposts := make([]post.Post, 0)
		getpostsql := "select * from " + dbmap.Dialect.QuotedTableForQuery("", tm.TableName) + " where WebPostId = :post_id"
		_, err = dbmap.Select(&dbposts, getpostsql, map[string]interface{}{
			"post_id": parsedpost.WebPostId,
		})
		if err != nil {
			fmt.Printf(fmt.Sprintf("Getting PostId %s from DB failed: %s", parsedpost.WebPostId, err.Error()))
		}
		var dbpost *post.Post
		if len(dbposts) == 1 {
			dbpost = &dbposts[0]
		} else if len(dbposts) > 1 {
			fmt.Printf(fmt.Sprintf("Query: %s returned %d rows", getpostsql, len(dbposts)))
		}
		postcount := len(dbposts)

		//var count int
		var buf []byte
		buf, err = GetJsonCommentList(parsedpost.WebPostId)
		if err != nil {
			fmt.Printf("GetJsonCommentList %s: failed%s\n", parsedpost.WebPostId, err.Error())
		}

		// Parse the comments into post structure
		err = ParseJsonComments(buf, parsedpost)
		if err != nil {
			fmt.Printf("ParseCommentsInto %s: failed%s\n", parsedpost.WebPostId, err.Error())
		}
		// New post? then insert
		if postcount == 0 {
			// Insert the new post into the database
			err = dbmap.InsertWithChilds(parsedpost)
			if err != nil {
				fmt.Println("insert into database failed: " + err.Error())
			}
		} else {
			// Update
			dbpost.CommentCount = parsedpost.CommentCount
			dbpost.Score = parsedpost.Score
			//dbpost.Comments = append(dbpost.Comments, &parsedpost.Comments)
			dbpost.Comments = parsedpost.Comments

			dbpost.CommentCount += uint64(len(dbpost.Comments))
			// Update the post into the database
			_, err = dbmap.UpdateWithChilds(dbpost)
			if err != nil {
				fmt.Println("update into database failed: " + err.Error())
			}
		}
	}

	fmt.Println("exit", uri)
}

func GetJsonCommentList(ID string) (buf []byte, err error) {

	if debugFromFile {
		var testfile string
		testfile = "testcomments." + ID + ".json.txt"
		buf, err := ioutil.ReadFile(testfile)
		if err != nil {
			fmt.Printf("Error reading json comment file %s: %s\n", testfile, err.Error())
			return nil, err
		}
		fmt.Println(string(buf))
		return buf, err
	}

	// Get data from url
	uri := fmt.Sprintf("https://www.reddit.com/r/golang/comments/%s.json", ID)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)

	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()

		// capture all bytes from the response body
		buf, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 { // 200 = OK
			httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", uri, resp.StatusCode, resp.Status)
			err = errors.New(httperr)
			return nil, err
		}
		//TraceJsonComments(buf)
		fmt.Println(string(buf))

		ioutil.WriteFile("testcomments."+ID+".json.txt", buf, 0)
		return buf, err

	} else {
		err = errors.New("Response from " + uri + " is nil")
		return nil, err
	}

	return nil, errors.New("Uncatched error in GetJsonCommentList")
}

func GetJsonPostList(uri string) (redditPostList *RedditJsonPostList, err error) {

	if debugFromFile {
		buf, err := ioutil.ReadFile("testposts.json.txt")
		if err != nil {
			fmt.Printf("Error reading json test file: %s\n", err.Error())
			return nil, err
		}

		var rpl RedditJsonPostList
		fmt.Println(string(buf))
		err = json.Unmarshal(buf, &rpl)
		return &rpl, err
	}
	// Get data from url
	//resp, err := http.Get(url)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)

	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()

		// capture all bytes from the response body
		buf, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 { // 200 = OK
			httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", uri, resp.StatusCode, resp.Status)
			err = errors.New(httperr)
			return nil, err
		}
		// Fix Reddit API type mismatch error
		// edited is reported as false(bool) if it really should be 0(float)
		// if the post was not edited after creation
		buf = bytes.Replace(buf, []byte(`"edited": false`), []byte(`"edited": 0`), -1)

		var rpl RedditJsonPostList
		fmt.Println(string(buf))
		err = json.Unmarshal(buf, &rpl)
		// Debug, write respones
		ioutil.WriteFile("testposts.json.txt", buf, 0)
		return &rpl, err

	} else {
		err = errors.New("Response from " + uri + " is nil")
		return nil, err
	}

	return nil, errors.New("Uncatched error in GetJsonPostList")
}

func InitDatabase() (*gorp.DbMap, error) {
	//drivername := "postgres"
	//dsn := "user=golang password=golang dbname=golang sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	drivername := "mysql"
	dsn := "golang:golang@/golang?parseTime=true"
	dialect := gorp.MySQLDialect{"InnoDB", "utf8mb4"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return nil, errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data using ping
	if err = db.Ping(); err != nil {
		return nil, errors.New("db.Ping failed: " + err.Error())
	}

	// Set the connection to use utf8bmb4
	if dialect.Engine == "InnoDB" {
		fmt.Println("Setting connection to utf8mb4")
		_, err = db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			return nil, errors.New("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci: " + err.Error())
		}
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	//defer dbmap.Db.Close()
	dbmap.DebugLevel = DebugLevel
	// Will log all SQL statements + args as they are run
	// The first arg is a string prefix to prepend to all log messages
	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "Trace:", log.Lmicroseconds))

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name

	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	table := dbmap.AddTableWithName(post.Post{}, "posts_reddit_test")
	table.SetKeys(true, "PID")

	// Add the comments table
	table = dbmap.AddTableWithName(post.Comment{}, "comments_reddit_test")
	table.SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	if err = dbmap.CreateTablesIfNotExists(); err != nil {
		return dbmap, errors.New("Create tables failed: " + err.Error())
	}

	// Force create all indexes for this database
	if err = dbmap.CreateIndexes(); err != nil {
		return dbmap, errors.New("Create indexes failed: " + err.Error())
	}

	return dbmap, nil
}
