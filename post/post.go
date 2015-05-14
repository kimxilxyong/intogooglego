// post.go
package post

import (
	"strconv"
	"time"
)

// holds a single post
type Post struct {
	Id       uint64
	Created  time.Time
	PostDate time.Time
	Site     string
	PostId   string
	Score    int
	Title    string
	Url      string
	User     string
	UserIP   string
	BodyType string
	Body     string
	Err      error `db:"-"` // ignore this field when storing with gorp
}

func (p *Post) String() (s string) {

	s = "Id = " + strconv.FormatUint(p.Id, 10) + "\n"
	s = s + "PostId = " + p.PostId + "\n"
	s = s + "Created = " + p.PostDate.String() + "\n"
	s = s + "Date = " + p.PostDate.String() + "\n"
	s = s + "User = " + p.User + "\n"
	s = s + "Title = " + p.Title + "\n"
	s = s + "Score = " + strconv.Itoa(p.Score) + "\n"
	s = s + "Url = " + p.Url
	return
}

func (p *Post) SetPostDate(postdate string) {
	pd, err := time.Parse(time.RFC3339, postdate)
	if err != nil {
		p.PostDate = time.Now()
	} else {
		p.PostDate = pd
	}
}

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
