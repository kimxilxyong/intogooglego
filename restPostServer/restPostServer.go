package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/go-json-rest-middleware-jwt"
	"github.com/kimxilxyong/gorp"
	"github.com/kimxilxyong/intogooglego/post"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"golang.org/x/net/http2"
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
	table_users    = "users_index_test"
	debug          = true
)

type Impl struct {
	Dbmap          *gorp.DbMap
	DebugSleep     int64
	DebugLevel     int
	jwt_middleware *jwt.JWTMiddleware
}

var lock = sync.RWMutex{}

var debugLevel = 3

func main() {

	i := Impl{DebugLevel: 3}

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

	i.jwt_middleware = &jwt.JWTMiddleware{
		Key:        []byte("secretdamnsecretfuostukeysff"),
		Realm:      "HolyRealm",
		DebugLevel: 3,
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(username string, password string) bool {

			if username == "admin" && password == "admin" {
				return true
			}

			var dbUser post.User
			err := i.Dbmap.SelectOne(&dbUser, "select * from "+table_users+" where name=?", username)
			if err != nil {
				if i.DebugLevel >= 2 {
					fmt.Printf("Cannot find user " + username + ", ERR: " + err.Error())
				}
				return false
			}

			return password == dbUser.Password
		},

		Authorizator: func(username string, request *rest.Request) bool {
			if username == "" {
				return false
			}
			return true
		},
		// Payload / claims
		PayloadFunc: func(userId string) map[string]interface{} {
			claims := make(map[string]interface{})
			claims["UserLevel"] = "9001"
			claims["SortOrder"] = "postdate" // Possible values: commentcount, score, postdate
			return claims
		},
	}

	api := rest.NewApi()

	var MiddleWareStack = []rest.Middleware{
		&rest.AccessLogApacheMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
			EnableLogAsJson:          true,
		},
		&rest.JsonIndentMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	}
	statusMw := &rest.StatusMiddleware{}

	api.Use(statusMw)

	api.Use(MiddleWareStack...)

	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			if debugLevel > 2 {
				fmt.Printf("AUTH Request.URL.Path: '%s' returning '%b'\n", request.URL.Path, request.URL.Path != "/login")
			}
			// Allow unauthenticated urls
			allowNonAuth := request.URL.Path == "/login" ||
				request.URL.Path == "/register" ||
				request.URL.Path == "/" ||
				request.URL.Path == "" ||
				strings.HasPrefix(request.URL.Path, "/favicon") ||
				strings.HasPrefix(request.URL.Path, "/css") ||
				strings.HasPrefix(request.URL.Path, "/js") ||
				strings.HasPrefix(request.URL.Path, "/html") ||
				strings.HasPrefix(request.URL.Path, "/t") ||
				strings.HasPrefix(request.URL.Path, "/img")
			return !allowNonAuth
			//return false // allow all for debug
		},
		IfTrue: i.jwt_middleware,
	})

	router, err := rest.MakeRouter(

		// JSON
		rest.Get("/j/t/:postid", i.JsonGetPostThreadComments),
		rest.Get("/j/p/:orderby", i.JsonGetPosts),
		rest.Get("/j/p/:orderby.:filterby", i.JsonGetPosts),

		// HTML
		rest.Get("/", i.SendStaticMainHtml),
		rest.Get("/t/:postid", i.SendStaticCommentsHtml),

		// Auth JWT
		rest.Post("/login", i.jwt_middleware.LoginHandler),
		rest.Post("/register", i.JwtRegisterUser),
		rest.Get("/jwttest", i.JwtTest),
		rest.Post("/jwtposttest", i.JwtPostTest),
		rest.Get("/refresh_token", i.jwt_middleware.RefreshHandler),

		// JSON Depricated
		//rest.Get("/p/:orderby", i.JsonGetAllPosts),
		//rest.Get("/p", i.JsonGetAllPosts),

		// HTML, Images, CSS and JS
		rest.Get("/img/*filename", i.SendStaticImage),
		rest.Get("/css", i.SendStaticCss),
		rest.Get("/css/#cssfile", i.SendStaticCss),
		rest.Get("/js/*jsfile", i.SendStaticJS),
		//rest.Get("/js", i.SendStaticJS),
		rest.Get("/#filename", i.GetStaticFile),

		rest.Get("/html/*filename", i.GetHtmlFile),
		rest.Get("/test/*filename", i.GetTestFile),

		rest.Get("/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	//http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))
	//http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
	http.Handle("/", api.MakeHandler())

	if debugLevel > 2 {
		fmt.Println("Starting http.ListenAndServe :8080")
	}
	//log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (i *Impl) LoginHandler(w rest.ResponseWriter, r *rest.Request) {
	i.DumpRequestHeader(r)
}

type registerUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (i *Impl) JwtRegisterUser(w rest.ResponseWriter, r *rest.Request) {

	//if debugLevel > 2 {
	fmt.Printf("ENV: %v\n", r.Env)
	//fmt.Printf("JWT Auth User: %s\n", r.Env["REMOTE_USER"].(string))
	//}
	//w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})

	i.DumpRequestHeader(r)

	register_vals := registerUser{}
	err := r.DecodeJsonPayload(&register_vals)
	if err != nil {
		// DEBUG
		fmt.Printf("*** DecodeJsonPayload Error: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.WriteJson(map[string]interface{}{
			"Error":                err.Error(),
			"JwtValidationMessage": "Decode failed",
			"JwtValidationCode":    99})
		return
	}

	// Check if the user already exists
	count, err := i.Dbmap.SelectInt("select count(*) from "+table_users+" where name = :name",
		map[string]interface{}{
			"name": register_vals.Username,
		})
	if err != nil {
		// DEBUG
		fmt.Printf("*** Register i.Dbmap.SelectInt Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteJson(map[string]interface{}{
			"Error":                err.Error(),
			"JwtValidationMessage": "Select count failed",
			"JwtValidationCode":    99})
		return
	}

	if count > 0 {
		// User already exists
		// DEBUG
		fmt.Printf("*** Register User " + register_vals.Username + " already exists")
		w.WriteHeader(http.StatusNotAcceptable)
		w.WriteJson(map[string]interface{}{
			"Error":                "user " + register_vals.Username + " already exists",
			"JwtValidationMessage": "Register failed",
			"JwtValidationCode":    99})
		return
	}

	newUser := post.User{
		Name:     register_vals.Username,
		Password: register_vals.Password,
		Created:  time.Unix(time.Now().Unix(), 0).UTC(),
	}

	err = i.Dbmap.Insert(&newUser)
	if err != nil {
		// DEBUG
		fmt.Printf("*** Register i.Dbmap.Insert Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteJson(map[string]interface{}{
			"Error":                err.Error(),
			"JwtValidationMessage": "Insert failed",
			"JwtValidationCode":    99})
		return
	}
	w.WriteJson(&newUser)

	//rest.Error(w, "Invalid Endpoint", http.StatusBadRequest)
}

func (i *Impl) JwtPostTest(w rest.ResponseWriter, r *rest.Request) {

	//if debugLevel > 2 {
	fmt.Printf("ENV: %v\n", r.Env)
	//fmt.Printf("JWT Auth User: %s\n", r.Env["REMOTE_USER"].(string))
	//}
	//w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})

	i.DumpRequestHeader(r)

	/*user := post.User{
		Name:   "TestUser",
		Id:     123,
		Avatar: "AVATAR",
	}
	w.WriteJson(&user)
	*/
	//rest.Error(w, "Invalid Endpoint", http.StatusBadRequest)

	i.jwt_middleware.LoginHandler(w, r)
}

func (i *Impl) JwtTest(w rest.ResponseWriter, r *rest.Request) {

	//if debugLevel > 2 {
	fmt.Printf("ENV: %v\n", r.Env)
	//fmt.Printf("JWT Auth User: %s\n", r.Env["REMOTE_USER"].(string))
	//}
	//w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})

	/*user := post.User{
		Name:   "TestUser",
		Id:     123,
		Avatar: "AVATAR",
	}
	*/
	i.jwt_middleware.LoginHandler(w, r)
	//w.WriteJson(&user)
	//rest.Error(w, "Invalid Endpoint", http.StatusBadRequest)
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

func (i *Impl) JsonGetPosts(w rest.ResponseWriter, r *rest.Request) {

	// Get starttime for measuring how long this functions takes
	timeStart := time.Now()

	i.DumpRequestHeader(r)

	// Sleep for debugging delay . DEBUG
	time.Sleep(500 * time.Millisecond)

	i.SetResponseContentType("application/json", &w)

	sort := "desc"
	orderby := r.PathParam("orderby")
	if (orderby != "postdate") && (orderby != "commentcount") && (orderby != "score") {
		rest.Error(w, "Invalid Endpoint", http.StatusBadRequest)
		return
	}

	filterby := r.PathParam("filterby")
	if filterby != "" {
		filterby = strings.ToLower(filterby)
		filterby = strings.Replace(filterby, "=", ") like '", 1)
		//filterby = strings.Replace(filterby, "delete", "", 99)
		//filterby = strings.Replace(filterby, "insert", "", 99)
		//filterby = strings.Replace(filterby, "update", "", 99)
		filterby = "lower(" + filterby + string('%') + "'"
		log.Printf("************* JsonGetPosts: filter :%s:\n", filterby)

	}

	const PARAM_OFFSET = "offset"
	const PARAM_LIMIT = "limit"
	const PARAM_FILTER_BY_POSTER = "fbp"
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
	postlist := post.Posts{}
	postlist.JsonApiVersion = post.API_VERSION

	//	postlist.Posts := make([]post.Post, 0)
	getpostsql := "select * from " + i.Dbmap.Dialect.QuotedTableForQuery("", table_posts)

	if filterby != "" {
		println("----------")
		println(filterby)
		println(getpostsql)
		filterby = " user like \"a%\""
		getpostsql = getpostsql + " where" + filterby
		println(getpostsql)
		println("----------")
	}
	if orderby != "" {
		getpostsql = getpostsql + " order by " + orderby + " " + sort
	}

	if (Limit > -1) && (Offset > -1) {
		getpostsql = fmt.Sprintf(getpostsql+" limit %d offset %d", Limit, Offset)
	} else {
		if (Limit < 0) && (Offset > 0) {
			Limit = 999999999999999999
			getpostsql = fmt.Sprintf(getpostsql+" limit %d offset %d", Limit, Offset)
		}
	}

	//getpostsql = "select * from `posts_index_test`"
	getpostsql = "select * from " + i.Dbmap.Dialect.QuotedTableForQuery("", table_posts)

	//getpostsql = getpostsql + " where user like \"a%\""
	filterby = " user like \"a%\""
	getpostsql = getpostsql + " where " + filterby + " order by commentcount desc limit 2 offset 0"
	//getpostsql = "select * from `posts_index_test`" + " where user like \"a%\""
	println(getpostsql)

	_, err = i.Dbmap.Select(&postlist.Posts, getpostsql)
	if err != nil {
		err = errors.New(fmt.Sprintf("Getting posts from DB failed: %s", err.Error()))
		postlist.Posts = append(postlist.Posts, &post.Post{Err: err})

	}
	if i.DebugLevel > 2 {
		log.Printf("Postlist len=%d\n", len(postlist.Posts))
		log.Printf("JsonGetPosts: '%s'\n", getpostsql)
	}

	// Get and set the execution time in milliseconds
	postlist.RequestDuration = (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))

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

	for _, comment := range dbpost.Comments {
		markedDown := blackfriday.MarkdownCommon([]byte(comment.Body))
		comment.Body = string(bluemonday.UGCPolicy().SanitizeBytes(markedDown))
	}
	postlist.Posts = append(postlist.Posts, dbpost)

	// Get and set the execution time in milliseconds
	postlist.RequestDuration = (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))

	w.WriteJson(&postlist)
}

// ReplaceStaticTemplates replaces {{header-template}}, {{aside-template}}, aso
// the actual html from holy-batman-header.html, holy-batman-aside.html, aso
func (i *Impl) ReplaceStaticTemplates(htmlData []byte) ([]byte, error) {

	// Header template
	htmlHeader, err := ioutil.ReadFile("holy-batman-header.html")
	if err != nil {
		return htmlData, err
	}
	template := []byte("{{header-template}}")
	htmlData = bytes.Replace(htmlData, template, htmlHeader, 1)
	// Header template end

	// Aside template
	htmlHeader, err = ioutil.ReadFile("holy-batman-aside.html")
	if err != nil {
		return htmlData, err
	}
	template = []byte("{{aside-template}}")
	htmlData = bytes.Replace(htmlData, template, htmlHeader, 1)
	// Aside template end

	// Nav template
	htmlHeader, err = ioutil.ReadFile("holy-batman-nav.html")
	if err != nil {
		return htmlData, err
	}
	template = []byte("{{nav-template}}")
	htmlData = bytes.Replace(htmlData, template, htmlHeader, 1)
	// Nav template end

	// Footer template
	htmlHeader, err = ioutil.ReadFile("holy-batman-footer.html")
	if err != nil {
		return htmlData, err
	}
	template = []byte("{{footer-template}}")
	htmlData = bytes.Replace(htmlData, template, htmlHeader, 1)
	// footer template end

	return htmlData, nil
}

func (i *Impl) SendStaticMainHtml(w rest.ResponseWriter, r *rest.Request) {
	rw := w.(http.ResponseWriter)

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	htmlData, err := ioutil.ReadFile("index.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	htmlData, err = i.ReplaceStaticTemplates(htmlData)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmlData)
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

	htmlData, err := ioutil.ReadFile("t.html")
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	htmlData, err = i.ReplaceStaticTemplates(htmlData)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	template := []byte("{{postid}}")
	postreplace := []byte(postid)
	htmlData = bytes.Replace(htmlData, template, postreplace, 1)

	// Write the bytes back
	rw.WriteHeader(http.StatusOK)
	var x int
	x, err = rw.Write(htmlData)
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

	if debugLevel >= 2 {
		fmt.Printf("SendStaticJS: '%s'\n", jsfile)
	}
	req := r.Request
	rw := w.(http.ResponseWriter)

	if jsfile != "" {
		jsfile = "js/" + jsfile
		if _, err := os.Stat(jsfile); os.IsNotExist(err) {
			errormsg := fmt.Sprintf("no such file: %s", jsfile)
			fmt.Println(errormsg)
			http.Error(rw, errormsg, http.StatusNotFound)
		} else {
			// ServeFile replies to the request with the contents of the named file
			i.SetResponseContentType("text/javascript", &w)
			http.ServeFile(rw, req, jsfile)
		}
	} else {
		//http.Error(rw, "File not found", http.StatusNotFound)
		http.Error(rw, "", http.StatusNotFound)
	}
}

func (i *Impl) SendStaticImage(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)

	filename := r.PathParam("filename")
	extension := path.Ext(filename)

	if extension != "" {
		extension = extension[1:]
		if extension == "svg" {
			extension = extension + "+xml"
		}
		i.SetResponseContentType("image/"+extension, &w)
	}

	if debugLevel > 2 {
		fmt.Printf("SendStaticImage filename: '%s'\n", filename)
		fmt.Printf("SendStaticImage extension: '%s'\n", extension)
	}
	req := r.Request
	rw := w.(http.ResponseWriter)
	if filename != "" {
		filename = "images/" + filename
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			errormsg := fmt.Sprintf("SendStaticImage: no such file: %s", filename)
			fmt.Println(errormsg)
			http.Error(rw, errormsg, http.StatusNotFound)
		} else {
			// ServeFile replies to the request with the contents of the named file
			http.ServeFile(rw, req, filename)
		}
	} else {
		//http.Error(rw, "File not found", http.StatusNotFound)
		http.Error(rw, "", http.StatusNotFound)
	}
}

func (i *Impl) GetHtmlFile(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)
	i.SetResponseContentType("text/html", &w)

	filename := r.PathParam("filename")

	fmt.Printf("GetHtmlTestFile: '%s'\n", filename)

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

func (i *Impl) GetStaticFile(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)
	//i.SetResponseContentType("text/html", &w)

	filename := r.PathParam("filename")

	req := r.Request
	rw := w.(http.ResponseWriter)
	if filename != "" {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			errormsg := fmt.Sprintf("GetTestFile: no such file or directory: %s", filename)
			fmt.Println(errormsg)
			http.Error(rw, errormsg, http.StatusNotFound)
		} else {
			// ServeFile replies to the request with the contents of the named file
			http.ServeFile(rw, req, filename)
		}
	} else {
		//http.Error(rw, "File not found", http.StatusNotFound)
		http.Error(rw, "", http.StatusNotFound)
	}
}

func (i *Impl) GetTestFile(w rest.ResponseWriter, r *rest.Request) {

	i.DumpRequestHeader(r)
	//i.SetResponseContentType("text/html", &w)

	filename := r.PathParam("filename")

	fmt.Printf("GetTestFile: '%s'\n", filename)

	req := r.Request
	rw := w.(http.ResponseWriter)
	if filename != "" {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			errormsg := fmt.Sprintf("GetTestFile: no such file or directory: %s", filename)
			fmt.Println(errormsg)
			http.Error(rw, errormsg, http.StatusNotFound)
		} else {
			// ServeFile replies to the request with the contents of the named file
			http.ServeFile(rw, req, filename)
		}
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

	// Add the users table
	table = i.Dbmap.AddTableWithName(post.User{}, table_users)
	table.SetKeys(true, "Id")

	// Add the comments table
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
	fmt.Println("----------------------------------------")
	fmt.Printf("Request Header: %s\n", s)
	fmt.Println("----------------------------------------")

	//fmt.Printf("Username: %s\n", r.URL.User.Username())

	return err
}
