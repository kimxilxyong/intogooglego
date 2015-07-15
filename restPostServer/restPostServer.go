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
	"os"
	"strconv"
	"sync"
)

const (
	table_comments = "comments_index_test"
	table_posts    = "posts_index_test"
)

type Impl struct {
	Dbmap *gorp.DbMap
}

var lock = sync.RWMutex{}

func main() {

	i := Impl{}
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

		rest.Get("/p/:orderby", i.GetAllPosts),
		rest.Get("/p", i.GetAllPosts),
		rest.Get("/t/:postid", i.GetPostThreadComments),
		rest.Get("/img/#filename", i.GetImage),
		rest.Get("/", i.SendStaticMainHtml),
		rest.Get("/c/:postid", i.SendStaticCommentsHtml),
		rest.Get("/b/:postid", i.SendStaticBlapbHtml),
		rest.Get("/css", i.SendStaticCss),
		rest.Get("/js/#jsfile", i.SendStaticJS),
		rest.Get("/jtable/*jtfile", i.SendStaticJTable),
		rest.Get("/js", i.SendStaticJS),

		rest.Get("/test/*filename", i.GetHtmlFile),

		rest.Get("/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),

		rest.Delete("/countries/:code", DeleteCountry),
		rest.Get("/countries", GetAllCountries),
		rest.Post("/countries", PostCountry),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	//http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func (i *Impl) GetAllPosts(w rest.ResponseWriter, r *rest.Request) {

	sort := "desc"
	orderby := r.PathParam("orderby")
	if orderby == "title" {
		sort = "asc"
	}

	i.SetContentType(&w)
	lock.RLock()
	postlist := post.Posts{}

	//	postlist.Posts := make([]post.Post, 0)
	getpostsql := "select * from " + i.Dbmap.Dialect.QuotedTableForQuery("", table_posts)
	if orderby != "" {
		getpostsql = getpostsql + " order by " + orderby + " " + sort
	}

	fmt.Printf("Postlist len=%d\n", len(postlist.Posts))
	fmt.Printf("GetAllPosts: '%s'\n", getpostsql)

	_, err := i.Dbmap.Select(&postlist.Posts, getpostsql)
	if err != nil {
		err = errors.New(fmt.Sprintf("Getting posts from DB failed: %s", err.Error()))
		postlist.Posts = append(postlist.Posts, &post.Post{Err: err})

	}
	fmt.Printf("Postlist len=%d\n", len(postlist.Posts))
	lock.RUnlock()
	w.WriteJson(&postlist)
}

func (i *Impl) GetPostThreadComments(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)
	i.SetContentType(&w)

	//w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")

	postid := r.PathParam("postid")
	pid, err := strconv.ParseUint(postid, 10, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pid = pid

	lock.RLock()

	res, err := i.Dbmap.GetWithChilds(post.Post{}, pid)
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
	dbpost := res.(*post.Post)
	postlist.Posts = append(postlist.Posts, dbpost)

	lock.RUnlock()

	w.WriteJson(&postlist)
}

func (i *Impl) SendStaticMainHtml(w rest.ResponseWriter, r *rest.Request) {
	req := r.Request
	rw := w.(http.ResponseWriter)
	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, "index.html")
}

func (i *Impl) SendStaticBlapbHtml(w rest.ResponseWriter, r *rest.Request) {

	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)

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
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
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
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmldata)
	if err != nil {
		rest.Error(w, fmt.Sprintf("Failed to write %d bytes: %s", x, err.Error()), http.StatusNoContent)
		return
	}

}

func (i *Impl) SendStaticCss(w rest.ResponseWriter, r *rest.Request) {
	req := r.Request
	rw := w.(http.ResponseWriter)
	// ServeFile replies to the request with the contents of the named file or directory.
	http.ServeFile(rw, req, "default.css")
}

func (i *Impl) SendStaticJS(w rest.ResponseWriter, r *rest.Request) {

	jsfile := r.PathParam("jsfile")
	if jsfile == "" {
		jsfile = "default.js"
	}

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

func (i *Impl) GetImage(w rest.ResponseWriter, r *rest.Request) {

	filename := r.PathParam("filename")
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

func (i *Impl) SetContentType(w *rest.ResponseWriter) {
	r := *w
	r.Header().Set("Content-Type", "application/json; charset=utf-8")
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
	fmt.Printf("Request: %s\n", string(b))
	fmt.Printf("RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Println("----------------------------------------")

	var bufWriter bytes.Buffer
	err = r.Header.Write(&bufWriter)
	if err != nil {
		fmt.Printf("r.Header.Write error: %s\n", err.Error())
		return err
	}

	s := bufWriter.String()
	fmt.Printf("Request Header: %s\n", s)
	fmt.Println("----------------------------------------")

	//fmt.Printf("Username: %s\n", r.URL.User.Username())

	return err
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	countries := make([]Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = *country
		i++
	}
	lock.RUnlock()
	w.WriteJson(&countries)
}

func PostCountry(w rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	lock.Lock()
	store[country.Code] = &country
	lock.Unlock()
	w.WriteJson(&country)
}

func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	lock.Lock()
	delete(store, code)
	lock.Unlock()
	w.WriteHeader(http.StatusOK)
}
