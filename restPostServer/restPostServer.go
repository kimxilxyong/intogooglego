package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	table_comments = "comments_index_test"
	table_posts    = "posts_index_test"
	debug          = true
)

type Impl struct {
	Dbmap      *gorp.DbMap
	DebugSleep int64
}

var lock = sync.RWMutex{}

func main() {

	i := Impl{}

	i.DebugSleep = 5000

	err := i.InitDB()
	if err != nil {
		panic("Failed to init database: " + err.Error())
	}
	if i.Dbmap != nil {
		defer i.Dbmap.Db.Close()
	} else {
		panic("Failed to init database, dbmap is nil: " + err.Error())
	}

	api := rest.NewApi()

	var MiddleWareStack = []rest.Middleware{
		&rest.AccessLogApacheMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		//&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
		},
		&rest.JsonIndentMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	}
	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)
	api.Use(MiddleWareStack...)
	router, err := rest.MakeRouter(

		// JSON
		rest.Get("/j/t/:postid", i.JsonGetPostThreadComments),

		// HTML
		rest.Get("/t/:postid", i.SendStaticCommentsHtml),

		// JSON Depricated
		rest.Get("/p/:orderby", i.JsonGetAllPosts),
		rest.Get("/p", i.JsonGetAllPosts),
		//rest.Get("/j/t/:postid", i.JsonGetPostThreadComments),

		// HTML, Images, CSS and JS

		rest.Get("/", i.SendStaticMainHtml),
		rest.Get("/b/:postid", i.SendStaticBlapbHtml),
		rest.Get("/l/:postid", i.SendStaticLazyHtml),
		rest.Get("/l2/:postid", i.SendStaticLazyHtml2),
		rest.Get("/l3/:postid", i.SendStaticLazyHtml3),
		rest.Get("/lastworking/:postid", i.SendStaticLastWorkingHtml),

		rest.Get("/img/#filename", i.SendStaticImage),
		rest.Get("/css", i.SendStaticCss),
		rest.Get("/css/#cssfile", i.SendStaticCss),
		rest.Get("/js/#jsfile", i.SendStaticJS),
		rest.Get("/jtable/*jtfile", i.SendStaticJTable),
		rest.Get("/api/names", i.SendStaticLazyJSONTable),
		rest.Get("/js", i.SendStaticJS),

		rest.Get("/test/*filename", i.GetHtmlFile),

		rest.Get("/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	//http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func (i *Impl) JsonGetAllPosts(w rest.ResponseWriter, r *rest.Request) {

	sort := "desc"
	orderby := r.PathParam("orderby")
	if orderby == "title" {
		sort = "asc"
	}

	i.SetResponseContentType("application/json", &w)
	lock.RLock()
	postlist := post.Posts{}

	//	postlist.Posts := make([]post.Post, 0)
	getpostsql := "select * from " + i.Dbmap.Dialect.QuotedTableForQuery("", table_posts)
	if orderby != "" {
		getpostsql = getpostsql + " order by " + orderby + " " + sort
	}

	log.Printf("Postlist len=%d\n", len(postlist.Posts))
	log.Printf("GetAllPosts: '%s'\n", getpostsql)

	_, err := i.Dbmap.Select(&postlist.Posts, getpostsql)
	if err != nil {
		err = errors.New(fmt.Sprintf("Getting posts from DB failed: %s", err.Error()))
		postlist.Posts = append(postlist.Posts, &post.Post{Err: err})

	}
	log.Printf("Postlist len=%d\n", len(postlist.Posts))
	lock.RUnlock()
	w.WriteJson(&postlist)
}

func (i *Impl) JsonGetPostThreadComments(w rest.ResponseWriter, r *rest.Request) {

	// Get starttime for measuring how long this functions takes
	timeStart := time.Now()

	i.DumpRequestHeader(r)

	// Sleep for debugging delay . DEBUG
	time.Sleep(500 * time.Millisecond)

	i.SetResponseContentType("application/json", &w)

	postid := r.PathParam("postid")
	pid, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	const PARAM_OFFSET = "offset"
	const PARAM_LIMIT = "limit"
	var Offset int64
	var Limit int64

	Offset = -1
	Limit = -1

	// set all map query strings to lowercase
	m, err := url.ParseQuery(strings.ToLower(r.URL.RawQuery))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the query params for limit and offset (if exists)
	if m.Get(PARAM_OFFSET) != "" {
		Offset, err = strconv.ParseInt(m.Get(PARAM_OFFSET), 10, 64)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if m.Get(PARAM_LIMIT) != "" {
		Limit, err = strconv.ParseInt(m.Get(PARAM_LIMIT), 10, 64)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	lock.RLock()
	defer lock.RUnlock()

	log.Printf("Limit %d, Offset %d\n", Limit, Offset)
	res, err := i.Dbmap.GetWithChilds(post.Post{}, Limit, Offset, pid)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res == nil {
		err = errors.New(fmt.Sprintf("Get post id %d not found", pid))
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	postlist := post.Posts{}
	postlist.JsonApiVersion = post.API_VERSION
	dbpost := res.(*post.Post)
	postlist.Posts = append(postlist.Posts, dbpost)

	// Get and set the execution time in milliseconds
	postlist.RequestDuration = (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))

	w.WriteJson(&postlist)
}

func (i *Impl) SendStaticMainHtml(w rest.ResponseWriter, r *rest.Request) {
	req := r.Request
	rw := w.(http.ResponseWriter)
	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, "index.html")
}

func (i *Impl) SendStaticLazyHtml(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("lazy.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticLazyHtml2(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("lazy2.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticLazyHtml3(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("lazy3.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticLastWorkingHtml(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("layout_working.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticBlapbHtml(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("comments.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticCommentsHtml(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	postid := r.PathParam("postid")
	_, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	htmldata, err := ioutil.ReadFile("t.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmldata = bytes.Replace(htmldata, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}
}

func (i *Impl) SendStaticCss(w rest.ResponseWriter, r *rest.Request) {

	cssfile := r.PathParam("cssfile")
	if cssfile == "" {
		cssfile = "default.css"
	}

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/css", &w)

	cssfile = "css/" + cssfile
	fmt.Printf("SendStaticCSS: '%s'\n", cssfile)
	req := r.Request
	rw := w.(http.ResponseWriter)
	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, cssfile)
}

func (i *Impl) SendStaticJS(w rest.ResponseWriter, r *rest.Request) {

	jsfile := r.PathParam("jsfile")
	if jsfile == "" {
		jsfile = "default.js"
	}

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/javascript", &w)

	jsfile = "js/" + jsfile
	fmt.Printf("SendStaticJS: '%s'\n", jsfile)

	req := r.Request
	rw := w.(http.ResponseWriter)
	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, jsfile)
}

func (i *Impl) SendStaticJTable(w rest.ResponseWriter, r *rest.Request) {
	req := r.Request
	rw := w.(http.ResponseWriter)
	jtfile := r.PathParam("jtfile")
	if jtfile == "" {
		http.Error(rw, "", http.StatusNotFound)
		return
	}
	jtfile = "jtable/" + jtfile
	fmt.Printf("SendStaticJTable: '%s'\n", jtfile)

	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, jtfile)
}

func (i *Impl) SendStaticLazyJSONTable(w rest.ResponseWriter, r *rest.Request) {
	req := r.Request
	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	//si.SetContentType(&w)

	jtfile := "static_json_api_names.txt"
	fmt.Printf("SendStaticLazyJSONTable: '%s'\n", jtfile)

	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, jtfile)
}

func (i *Impl) SendStaticImage(w rest.ResponseWriter, r *rest.Request) {

	filename := r.PathParam("filename")
	extension := path.Ext(filename)

	i.SetResponseContentType("image/"+extension[1:], &w)

	req := r.Request
	rw := w.(http.ResponseWriter)
	if filename != "" {
		// ServeFile replies to the request with the contents of the named file or directory.
		http.ServeFile(rw, req, "images/"+filename)
	} else {
		//http.Error(rw, "File not found", http.StatusNotFound)
		http.Error(rw, "", http.StatusNotFound)
	}
}

func (i *Impl) GetHtmlFile(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	filename := r.PathParam("filename")

	fmt.Printf("GetTestFile: '%s'\n", filename)

	req := r.Request
	rw := w.(http.ResponseWriter)
	if filename != "" {
		// ServeFile replies to the request with the contents of the named file or directory.
		filename = filename
		http.ServeFile(rw, req, filename)
	} else {
		//http.Error(rw, "File not found", http.StatusNotFound)
		http.Error(rw, "", http.StatusNotFound)
	}
}

func (i *Impl) InitDB() (err error) {
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
	i.Dbmap = &gorp.DbMap{Db: db, Dialect: dialect}

	i.Dbmap.DebugLevel = 3
	// Will log all SQL statements + args as they are run
	// The first arg is a string prefix to prepend to all log messages
	i.Dbmap.TraceOn("[gorp]", log.New(os.Stdout, "fetch:", log.Lmicroseconds))

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name

	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	//table := i.Dbmap.AddTableWithName(post.Post{}, "posts_index_test")
	table := i.Dbmap.AddTableWithName(post.Post{}, table_posts)

	table.SetKeys(true, "PID")

	// Add the comments table
	//table = i.Dbmap.AddTableWithName(post.Comment{}, "comments_index_test")
	table = i.Dbmap.AddTableWithName(post.Comment{}, table_comments)
	table.SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	if err = i.Dbmap.CreateTablesIfNotExists(); err != nil {
		return errors.New("Create tables failed: " + err.Error())
	}

	// Force create all indexes for this database
	if err = i.Dbmap.CreateIndexes(); err != nil {
		return errors.New("Create indexes failed: " + err.Error())
	}

	// Show Connection info

	return
}

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

func (i *Impl) SetResponseContentType(ctype string, w *rest.ResponseWriter) {
	r := *w
	ct := fmt.Sprintf("%s; charset=utf-8", ctype)
	r.Header().Set("Content-Type", ct)
}

func (i *Impl) DumpRequestHeader(r *rest.Request) error {

	//var req *http.Request

	req := r.Request
	//req := http.Request(*r)
	b, err := httputil.DumpRequest(req, false)
	if err != nil {
		fmt.Printf("DumpRequestHeader error: %s\n", err.Error())
		return err
	}
	fmt.Println("------- DumpRequestHeader --------------")
	fmt.Printf("Request: %s\n", string(b))
	fmt.Printf("RemoteAddr: %s\n", r.RemoteAddr)

	var bufWriter bytes.Buffer
	err = r.Header.Write(&bufWriter)
	if err != nil {
		fmt.Printf("r.Header.Write error: %s\n", err.Error())
		fmt.Println("----------------------------------------")
		return err
	}

	s := bufWriter.String()
	fmt.Printf("Request Header: %s\n", s)
	fmt.Println("----------------------------------------")

	//fmt.Printf("Username: %s\n", r.URL.User.Username())

	return err
}
