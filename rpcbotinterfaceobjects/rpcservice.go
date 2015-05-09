/*
	Defines the structures holding all info about one post
*/

package rpcbotinterfaceobjects

import (
	"strconv"
	"strings"
	"time"
)

type Bot struct {
	ServiceCallCount int
}

// test ProcessPost function - turn the input to uppercase
func (this *Bot) ProcessPost(in *BotInput, out *BotOutput) error {
	out.SetContent(strings.ToUpper(in.GetContent()) + " = " + strconv.Itoa(this.ServiceCallCount))
	this.ServiceCallCount++
	return nil
}

// holds a single post
type Post struct {
	Id       int64
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
}

func NewPost(site, postid, postdate, score, title, url, user string) Post {

	posttime, err := time.Parse(time.RFC3339, postdate)
	if err != nil {
		posttime = time.Now()
	}

	postscore, err := strconv.ParseInt(score, 10, 32)
	if err != nil {
		postscore = -1
	}

	return Post{
		Created:  time.Now(),
		PostId:   postid,
		PostDate: posttime,
		Site:     site,
		Score:    int(postscore),
		Title:    title,
		Url:      url,
		User:     user,
	}
}
