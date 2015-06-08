// post.go
package post

import (
	"strconv"
	"time"
)

// holds a single post
type Post struct {
	Id        uint64    `db:"notnull, PID, primarykey, autoincrement"`
	Created   time.Time `db:"notnull"`
	PostDate  time.Time `db:"notnull"`
	Site      string    `db:"name: PostSite, notnull, size:50, index:idx_site"`
	WebPostId string    `db:"notnull, size:32, uniqueindex:idx_webpost"`
	Score     int       `db:"notnull"`
	Title     string    `db:"notnull"`
	Url       string    `db:"notnull"`
	User      string    `db:"index:idx_user, size:64"`
	PostSub   string    `db:"index:idx_user, size:128"`
	UserIP    string    `db:"notnull, index:idx_user, size:16"`
	BodyType  string    `db:"notnull, size:64"`
	Body      string    `db:"name:PostBody, size:16384"`
	Err       error     `db:"-"` // ignore this field when storing with gorp
	Comments  []*Comment
}

// holds a single comment bound to a post
type Comment struct {
	Id            uint64    `db:"notnull, primarykey, autoincrement"`
	PostId        uint64    `db:"notnull, index:idx_foreign_key_postid"` // points to post.id
	CommentDate   time.Time `db:"notnull"`
	User          string    `db:"size:64"`
	Title         string    `db:"size:256"`
	Body          string    `db:"name:CommentBody, size:16384"`
	ParseComplete bool      `db:"-"` // ignore this field when storing with gorp
	Err           error     `db:"-"` // ignore this field when storing with gorp
}

func (p *Post) String() (s string) {

	s = "Id = " + strconv.FormatUint(p.Id, 10) + "\n"
	s = s + "WebPostId = " + p.WebPostId + "\n"
	s = s + "Created = " + p.Created.String() + "\n"
	s = s + "Date = " + p.PostDate.String() + "\n"
	s = s + "User = " + p.User + "\n"
	s = s + "Title = " + p.Title + "\n"
	s = s + "Score = " + strconv.Itoa(p.Score) + "\n"
	s = s + "Url = " + p.Url
	return
}

func (c *Comment) String() (s string) {

	s = "Id = " + strconv.FormatUint(c.Id, 10) + "\n"
	s = s + "PostId = " + strconv.FormatUint(c.PostId, 10) + "\n"
	s = s + "Date = " + c.CommentDate.String() + "\n"
	s = s + "User = " + c.User + "\n"
	s = s + "Body = " + c.Body + "\n"
	return
}

// Set the data a post was posted - if it cannot be parsed use the current datetime
func (p *Post) SetPostDate(postdate string) {
	pd, err := time.Parse(time.RFC3339, postdate)
	if err != nil {
		p.PostDate = time.Now()
	} else {
		p.PostDate = pd
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

func NewPost() Post {

	return Post{
		Created: time.Now(),
	}
}

func NewComment() Comment {

	return Comment{
		CommentDate: time.Now(),
	}
}
