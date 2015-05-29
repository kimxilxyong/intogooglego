// post.go
package post

import (
	"strconv"
	"time"
)

// holds a single post
type Post struct {
	Id           uint64    `db:"notnull, PID, primarykey, autoincrement"`
	SecondTestID int       `db:"notnull, name: SID"`
	Created      time.Time `db:"notnull, primarykey"`
	PostDate     time.Time `db:"notnull"`
	Site         string    `db:"name: PostSite, notnull, size:50, index:idx_site"`
	PostId       string    `db:"notnull, size:32, unique, index:idx_site"`
	Score        int       `db:"notnull"`
	Title        string    `db:"notnull"`
	Url          string    `db:"notnull"`
	User         string    `db:"index:idx_user, size:64"`
	PostSub      string    `db:"index:idx_user, size:128"`
	UserIP       string    `db:"notnull, index:idx_user, size:16"`
	BodyType     string    `db:"notnull, size:64"`
	Body         string    `db:"name:PostBody, size:16384"`
	Err          error     `db:"-"` // ignore this field when storing with gorp
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
