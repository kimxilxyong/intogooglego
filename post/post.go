// post.go
package post

import (
	"strconv"
	"time"
)

// holds a single post
type Post struct {
	Id       uint64    `gorm:"column:Id;primary_key"`
	Created  time.Time `gorm:"column:Created"`
	PostDate time.Time `gorm:"column:PostDate"`
	Site     string    `gorm:"column:Site"`
	PostId   string    `gorm:"column:PostId"`
	Score    int       `gorm:"column:Score"`
	Title    string    `gorm:"column:Title"`
	Url      string    `gorm:"column:Url"`
	User     string    `gorm:"column:User"`
	UserIP   string    `gorm:"column:UserIP"`
	BodyType string    `gorm:"column:BodyType"`
	Body     string    `gorm:"column:Body"`
	Err      error     `sql:"-", db:"-"` // ignore this field when storing with gorp or gorm
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
