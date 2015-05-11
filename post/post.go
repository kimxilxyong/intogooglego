// post.go
package post

import (
	"fmt"
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

func (p *Post) String() {

	fmt.Println("Id = " + strconv.FormatUint(p.Id, 10))
	fmt.Println("Created = " + p.PostDate.String())
	fmt.Println("Date = " + p.PostDate.String())
	fmt.Println("User = " + p.User)
	fmt.Println("Title = " + p.Title)
	fmt.Println("Score = " + strconv.Itoa(p.Score))
	fmt.Println("Url = " + p.Url)
	fmt.Println("-----------------------------------------------")

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
