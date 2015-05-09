package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"strings"
	"unicode"
)

func ExampleGoQuery() {

	// Get an io.Reader with HTML content
	io := getHtmlInputReader()

	// Create a qoquery document to parse from
	doc, err := goquery.NewDocumentFromReader(io)
	checkErr(err, "Failed to parse HTML")

	fmt.Println("---- Starting to parse ------------------------")

	// Find reddit posts = elements with class "thing"
	thing := doc.Find(".thing")
	for iThing := range thing.Nodes {

		// use `single` as a selection of 1 node
		singlething := thing.Eq(iThing)

		// get the reddit post identifier
		reddit_post_id, exists := singlething.Attr("data-fullname")
		if exists == true {

			// find an element with class title and a child with class may-blank
			reddit_post_title := singlething.Find(".title .may-blank").Text()
			reddit_post_user := singlething.Find(".author").Text()
			// find an element with class comments and may-blank (in the same element, note the space!)
			reddit_post_url, _ := singlething.Find(".comments.may-blank").Attr("href")
			reddit_post_score := singlething.Find(".score.likes").Text()
			reddit_postdate, exists := singlething.Find("time").Attr("datetime")

			if exists == true {

				// Remove CRLF and unnecessary whitespaces
				reddit_post_title = stringMinifier(reddit_post_title)

				// Print out the crawled info
				fmt.Println("Id = " + reddit_post_id)
				fmt.Println("Date = " + reddit_postdate)
				fmt.Println("User = " + reddit_post_user)
				fmt.Println("Title = " + reddit_post_title)
				fmt.Println("Score = " + reddit_post_score)
				fmt.Println("Url = " + reddit_post_url)
				fmt.Println("-----------------------------------------------")

			}
		}

	}
}

// Removes all unnecessary whitespaces
func stringMinifier(in string) (out string) {

	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}
	return
}

// returns an io.Reader with dummy test html
func getHtmlInputReader() io.Reader {
	s := `
<html>
  <head>
    <meta name="generator"
    content="HTML Tidy for HTML5 (experimental) for Windows https://github.com/w3c/tidy-html5/tree/c63cc39" />
    <title></title>
  </head>
  <body>
    <div class="thing id-t3_34z9xo odd link" onclick="click_thing(this)" data-fullname="t3_34z9xo">
      <p class="parent"></p>
      <span class="rank">1</span>
      <div class="midcol unvoted">
        <div class="arrow up login-required" onclick="$(this).vote(r.config.vote_hash, null, event)" role="button"
        aria-label="upvote" tabindex="0"></div>
        <div class="score dislikes">10</div>
        <div class="score unvoted">11</div>
        <div class="score likes">12</div>
        <div class="arrow down login-required" onclick="$(this).vote(r.config.vote_hash, null, event)" role="button"
        aria-label="downvote" tabindex="0"></div>
      </div>
      <div class="entry unvoted">
        <p class="title">
        <a class="title may-blank loggedin" href="https://github.com/dariubs/GoBooks" tabindex="1">dariubs/GoBooks: list of paper
        and electronic books on Go</a> 
        <span class="domain">(
        <a href="/domain/github.com/">github.com</a>)</span></p>
        <p class="tagline">submitted 
        <time title="Tue May 5 20:13:08 2015 UTC" datetime="2015-05-05T20:13:08+00:00" class="live-timestamp">10 hours ago</time>
        <a href="http://www.reddit.com/user/dgryski" class="author may-blank id-t2_3hcmx">dgryski</a></p>
        <ul class="flat-list buttons">
          <li class="first">
            <a href="http://www.reddit.com/r/golang/comments/34z9xo/dariubsgobooks_list_of_paper_and_electronic_books/"
            class="comments may-blank">1 comment</a>
          </li>
          <li class="share">
            <span class="share-button toggle" style="">
              <a class="option active login-required" href="#" tabindex="100"
              onclick="return toggle(this, share, cancelShare)">share</a>
              <a class="option" href="#">cancel</a>
            </span>
          </li>
          <li class="link-save-button save-button">
            <a href="#">save</a>
          </li>
          <li>
            <form action="/post/hide" method="post" class="state-button hide-button">
              <input name="executed" value="hidden" type="hidden" />
              <span>
                <a href="javascript:void(0)" onclick="change_state(this, &#39;hide&#39;, hide_thing);">hide</a>
              </span>
            </form>
          </li>
          <li class="report-button">
            <a href="javascript:void(0)" class="action-thing" data-action-form="#report-action-form">report</a>
          </li>
        </ul>
        <div class="expando" style="display: none">
          <span class="error">loading...</span>
        </div>
      </div>
      <div class="child"></div>
      <div class="clearleft"></div>
    </div>
   <div class="thing id-t3_359t2l odd link self" onclick="click_thing(this)" data-fullname="t3_359t2l">
      <p class="parent"></p>
      <span class="rank">1</span>
      <div class="midcol unvoted">
        <div class="arrow up login-required" onclick="$(this).vote(r.config.vote_hash, null, event)" role="button"
        aria-label="upvote" tabindex="0"></div>
        <div class="score likes">•</div>
        <div class="score unvoted">•</div>
        <div class="score dislikes">•</div>
        <div class="arrow down login-required" onclick="$(this).vote(r.config.vote_hash, null, event)" role="button"
        aria-label="downvote" tabindex="0"></div>
      </div>
      <div class="entry unvoted">
        <p class="title">
        <a class="title may-blank loggedin" href="/r/golang/comments/359t2l/bulding_api_services_in_go/" tabindex="1"
        rel="nofollow">Bulding API services in Go</a> 
        <span class="domain">(
        <a href="/r/golang/">self.golang</a>)</span></p>
        <div class="expando-button collapsed selftext" onclick="expando_child(this)"></div>
        <p class="tagline">submitted 
        <time title="Fri May 8 08:41:00 2015 UTC" datetime="2015-05-08T08:41:00+00:00" class="live-timestamp">a minute ago</time>
        by 
        <a href="http://www.reddit.com/user/jan1024188" class="author may-blank id-t2_8bbqm">jan1024188</a></p>
        <ul class="flat-list buttons">
          <li class="first">
            <a href="http://www.reddit.com/r/golang/comments/359t2l/bulding_api_services_in_go/"
            class="comments empty may-blank">comment</a>
          </li>
          <li class="share">
            <span class="share-button toggle" style="">
              <a class="option active login-required" href="#" tabindex="100"
              onclick="return toggle(this, share, cancelShare)">share</a>
              <a class="option" href="#">cancel</a>
            </span>
          </li>
          <li class="link-save-button save-button">
            <a href="#">save</a>
          </li>
          <li>
            <form action="/post/hide" method="post" class="state-button hide-button">
              <input name="executed" value="hidden" type="hidden" />
              <span>
                <a href="javascript:void(0)" onclick="change_state(this, &#39;hide&#39;, hide_thing);">hide</a>
              </span>
            </form>
          </li>
          <li class="report-button">
            <a href="javascript:void(0)" class="action-thing" data-action-form="#report-action-form">report</a>
          </li>
        </ul>
        <div class="expando" style="display: none">
          <span class="error">loading...</span>
        </div>
      </div>
      <div class="child"></div>
      <div class="clearleft"></div>
    </div>	
  </body>
</html>
`
	return strings.NewReader(s)
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func main() {
	ExampleGoQuery()
}
