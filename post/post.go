// post.go
package post

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"time"
)

const API_VERSION = 1

// Posts holds a slice of Post
type Posts struct {
	JsonApiVersion   int
	RequestDuration  int64
	RequestErrorCode int
	RequestErrorMsg  string
	Posts            []*Post
}

// Post holds a single post
// You can use ether db or gorp as tag
type Post struct {
	Id        uint64     `db:"notnull, PID, primarykey, autoincrement"`
	Created   time.Time  `db:"notnull"`
	PostDate  time.Time  `db:"notnull"`
	Site      string     `db:"name: PostSite, enforcenotnull, size:50, index:idx_site"`
	WebPostId string     `db:"enforcenotnull, size:32, uniqueindex:idx_webpost"`
	Score     int        `db:"notnull"`
	Title     string     `gorp:"notnull, size: 512"`
	Url       string     `db:"notnull, size:1024"`
	User      string     `db:"index:idx_user, size:64"`
	PostSub   string     `db:"index:idx_user, size:128"`
	Ignored   int        `gorp:"ignorefield"`
	UserIP    string     `db:"notnull, index:idx_user, size:16"`
	BodyType  string     `gorp:"notnull, size:64"`
	Body      string     `db:"name:PostBody, type:mediumtext"`
	Err       error      `db:"-"`               // ignore this field when storing with gorp
	Comments  []*Comment `db:"relation:PostId"` // will create a table Comment as a detail table with foreignkey PostId
	// if you want a different name just issue a: table = dbmap.AddTableWithName(post.Comment{}, "comments_embedded_test")
	// after: table := dbmap.AddTableWithName(post.Post{}, "posts_embedded_test")
	// but before: dbmap.CreateTablesIfNotExists()
	CommentParseErrors []*Comment `db:"-"`
	CommentCount       uint64
}

// holds a single comment bound to a post
type Comment struct {
	Id            uint64    `db:"notnull, primarykey, autoincrement"`
	PostId        uint64    `db:"notnull, index:idx_foreign_key_postid, uniqueindex:idx_webcomment"` // points to post.id
	WebCommentId  string    `db:"enforcenotnull, size:32, uniqueindex:idx_webcomment"`
	CommentDate   time.Time `db:"notnull"`
	User          string    `db:"size:64"`
	Title         string    `db:"size:4196"`
	Body          string    `db:"name:CommentBody, type:mediumtext"` //size:16383"`
	ParseComplete bool      `db:"-"`                                 // ignore this field when storing with gorp
	Err           error     `db:"-"`                                 // ignore this field when storing with gorp
}

func (p *Post) String(tag string) (s string) {

	s = tag + "Id = " + strconv.FormatUint(p.Id, 10) + "\n"
	s = s + tag + "WebPostId = " + p.WebPostId + "\n"
	s = s + tag + "Created = " + p.Created.String() + "\n"
	s = s + tag + "Date = " + p.PostDate.String() + "\n"
	s = s + tag + "User = " + p.User + "\n"
	s = s + tag + "Title = " + p.Title + "\n"
	s = s + tag + "Score = " + strconv.Itoa(p.Score) + "\n"
	s = s + tag + "Url = " + p.Url + "\n"

	for i, c := range p.Comments {
		s = s + fmt.Sprintf("---------- Comment %d START --------------\n", i)
		s = s + c.String(tag)
		s = s + fmt.Sprintf("---------- Comment %d END ----------------\n", i)
	}
	return
}

func (c *Comment) String(tag string) (s string) {

	tag = tag + "C: "
	s = tag + "Id = " + strconv.FormatUint(c.Id, 10) + "\n"
	s = s + tag + "PostId = " + strconv.FormatUint(c.PostId, 10) + "\n"
	s = s + tag + "WebCommentId = " + c.WebCommentId + "\n"
	s = s + tag + "Date = " + c.GetCommentDate().String() + "\n"
	s = s + tag + "Title = " + c.Title + "\n"
	s = s + tag + "User = " + c.User + "\n"
	s = s + tag + "Body = " + c.Body + "\n"
	s = s + tag + fmt.Sprintf("Hash = %d\n", c.Hash())
	return
}

func (p *Post) Hash() (h uint64) {
	h = Hash(strconv.Itoa(p.Score))
	for _, c := range p.Comments {
		h = h + c.Hash()
	}
	return
}

func (c *Comment) Hash() (h uint64) {
	h = Hash(
		c.GetCommentDate().String() +
			c.Title +
			c.User +
			c.Body)
	return
}
func (c *Comment) GetCommentDate() time.Time {
	return c.CommentDate.UTC()
}

func Hash(s string) uint64 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return uint64(h.Sum32())
}

// Set the date a post was posted - if it cannot be parsed use the current datetime
func (p *Post) SetPostDate(postdate string) {
	pd, err := time.Parse(time.RFC3339, postdate)
	if err != nil {
		p.PostDate = time.Unix(time.Now().Unix(), 0).UTC()
	} else {
		p.PostDate = time.Unix(pd.Unix(), 0).UTC()
	}
}

// Set the data a post was posted - if it cannot be parsed use the current datetime
func (c *Comment) SetCommentDate(commentdate string) {
	cd, err := time.Parse(time.RFC3339, commentdate)
	if err != nil {
		c.CommentDate = time.Unix(time.Now().Unix(), 0).UTC()
	} else {
		c.CommentDate = time.Unix(cd.Unix(), 0).UTC()
	}
}

// Set the score of a post or -1 if not convertable from string to int
func (p *Post) SetScore(score string) {
	ps, err := strconv.Atoi(score)
	if err != nil {
		p.Score = -1
	} else {
		p.Score = ps
	}
}

// Add a new comment to the post
func (p *Post) AddComment() *Comment {
	newComment := NewComment()
	p.Comments = append(p.Comments, &newComment)
	return &newComment
}

func NewPost() Post {

	return Post{
		Created: time.Unix(time.Now().Unix(), 0).UTC(),
	}
}

func NewComment() Comment {

	return Comment{
		CommentDate: time.Unix(time.Now().Unix(), 0).UTC(),
	}
}
