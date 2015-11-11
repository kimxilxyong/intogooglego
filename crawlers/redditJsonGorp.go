package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	//"net/url"
	//"os"
	"bytes"
	"strings"
)

type OriginalRedditJsonCommentList struct {
	Data struct {
		After    interface{} `json:"after"`
		Before   interface{} `json:"before"`
		Children []struct {
			Data struct {
				ApprovedBy          interface{} `json:"approved_by"`
				Archived            bool        `json:"archived"`
				Author              string      `json:"author"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				BannedBy            interface{} `json:"banned_by"`
				Clicked             bool        `json:"clicked"`
				Created             int         `json:"created"`
				CreatedUtc          int         `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Domain              string      `json:"domain"`
				Downs               int         `json:"downs"`
				Edited              bool        `json:"edited"`
				From                interface{} `json:"from"`
				FromID              interface{} `json:"from_id"`
				FromKind            interface{} `json:"from_kind"`
				Gilded              int         `json:"gilded"`
				Hidden              bool        `json:"hidden"`
				HideScore           bool        `json:"hide_score"`
				ID                  string      `json:"id"`
				IsSelf              bool        `json:"is_self"`
				Likes               interface{} `json:"likes"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				LinkFlairText       interface{} `json:"link_flair_text"`
				Locked              bool        `json:"locked"`
				Media               struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"media"`
				MediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"media_embed"`
				ModReports    []interface{} `json:"mod_reports"`
				Name          string        `json:"name"`
				NumComments   int           `json:"num_comments"`
				NumReports    interface{}   `json:"num_reports"`
				Over18        bool          `json:"over_18"`
				Permalink     string        `json:"permalink"`
				Quarantine    bool          `json:"quarantine"`
				RemovalReason interface{}   `json:"removal_reason"`
				ReportReasons interface{}   `json:"report_reasons"`
				Saved         bool          `json:"saved"`
				Score         int           `json:"score"`
				SecureMedia   struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"secure_media"`
				SecureMediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"secure_media_embed"`
				Selftext      string        `json:"selftext"`
				SelftextHTML  interface{}   `json:"selftext_html"`
				Stickied      bool          `json:"stickied"`
				Subreddit     string        `json:"subreddit"`
				SubredditID   string        `json:"subreddit_id"`
				SuggestedSort interface{}   `json:"suggested_sort"`
				Thumbnail     string        `json:"thumbnail"`
				Title         string        `json:"title"`
				Ups           int           `json:"ups"`
				URL           string        `json:"url"`
				UserReports   []interface{} `json:"user_reports"`
				Visited       bool          `json:"visited"`
			} `json:"data"`
			Kind string `json:"kind"`
		} `json:"children"`
		Modhash string `json:"modhash"`
	} `json:"data"`
	Kind string `json:"kind"`
}

type RedditJsonCommentList []struct {
	Data struct {
		After    interface{} `json:"after"`
		Before   interface{} `json:"before"`
		Children []struct {
			Data struct {
				ApprovedBy          interface{} `json:"approved_by"`
				Archived            bool        `json:"archived"`
				Author              string      `json:"author"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				BannedBy            interface{} `json:"banned_by"`
				Clicked             bool        `json:"clicked"`
				Created             float64     `json:"created"`
				CreatedUtc          float64     `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Domain              string      `json:"domain"`
				Downs               int         `json:"downs"`
				Edited              float64     `json:"edited"`
				From                interface{} `json:"from"`
				FromID              interface{} `json:"from_id"`
				FromKind            interface{} `json:"from_kind"`
				Gilded              int         `json:"gilded"`
				Hidden              bool        `json:"hidden"`
				HideScore           bool        `json:"hide_score"`
				ID                  string      `json:"id"`
				IsSelf              bool        `json:"is_self"`
				Likes               interface{} `json:"likes"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				LinkFlairText       interface{} `json:"link_flair_text"`
				Locked              bool        `json:"locked"`
				Media               struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"media"`
				MediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"media_embed"`
				ModReports    []interface{} `json:"mod_reports"`
				Name          string        `json:"name"`
				NumComments   int           `json:"num_comments"`
				NumReports    interface{}   `json:"num_reports"`
				Over18        bool          `json:"over_18"`
				Permalink     string        `json:"permalink"`
				Quarantine    bool          `json:"quarantine"`
				RemovalReason interface{}   `json:"removal_reason"`
				ReportReasons interface{}   `json:"report_reasons"`
				Saved         bool          `json:"saved"`
				Score         int           `json:"score"`
				SecureMedia   struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"secure_media"`
				SecureMediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"secure_media_embed"`
				Selftext      string        `json:"selftext"`
				SelftextHTML  interface{}   `json:"selftext_html"`
				Stickied      bool          `json:"stickied"`
				Subreddit     string        `json:"subreddit"`
				SubredditID   string        `json:"subreddit_id"`
				SuggestedSort interface{}   `json:"suggested_sort"`
				Thumbnail     string        `json:"thumbnail"`
				Title         string        `json:"title"`
				Ups           int           `json:"ups"`
				URL           string        `json:"url"`
				UserReports   []interface{} `json:"user_reports"`
				Visited       bool          `json:"visited"`
			} `json:"data"`
			Kind string `json:"kind"`
		} `json:"children"`
		Modhash string `json:"modhash"`
	} `json:"data"`
	Kind string `json:"kind"`
}

type RedditJsonPostList struct {
	Data struct {
		After    interface{} `json:"after"`
		Before   interface{} `json:"before"`
		Children []struct {
			Data struct {
				ApprovedBy          interface{} `json:"approved_by"`
				Archived            bool        `json:"archived"`
				Author              string      `json:"author"`
				AuthorFlairCSSClass interface{} `json:"author_flair_css_class"`
				AuthorFlairText     interface{} `json:"author_flair_text"`
				BannedBy            interface{} `json:"banned_by"`
				Clicked             bool        `json:"clicked"`
				Created             float64     `json:"created"`
				CreatedUtc          float64     `json:"created_utc"`
				Distinguished       interface{} `json:"distinguished"`
				Domain              string      `json:"domain"`
				Downs               int         `json:"downs"`
				Edited              float64     `json:"edited"`
				From                interface{} `json:"from"`
				FromID              interface{} `json:"from_id"`
				FromKind            interface{} `json:"from_kind"`
				Gilded              int         `json:"gilded"`
				Hidden              bool        `json:"hidden"`
				HideScore           bool        `json:"hide_score"`
				ID                  string      `json:"id"`
				IsSelf              bool        `json:"is_self"`
				Likes               interface{} `json:"likes"`
				LinkFlairCSSClass   interface{} `json:"link_flair_css_class"`
				LinkFlairText       interface{} `json:"link_flair_text"`
				Locked              bool        `json:"locked"`
				Media               struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"media"`
				MediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"media_embed"`
				ModReports    []interface{} `json:"mod_reports"`
				Name          string        `json:"name"`
				NumComments   int           `json:"num_comments"`
				NumReports    interface{}   `json:"num_reports"`
				Over18        bool          `json:"over_18"`
				Permalink     string        `json:"permalink"`
				Quarantine    bool          `json:"quarantine"`
				RemovalReason interface{}   `json:"removal_reason"`
				ReportReasons interface{}   `json:"report_reasons"`
				Saved         bool          `json:"saved"`
				Score         int           `json:"score"`
				SecureMedia   struct {
					Oembed struct {
						AuthorName      string `json:"author_name"`
						AuthorURL       string `json:"author_url"`
						Description     string `json:"description"`
						Height          int    `json:"height"`
						HTML            string `json:"html"`
						ProviderName    string `json:"provider_name"`
						ProviderURL     string `json:"provider_url"`
						ThumbnailHeight int    `json:"thumbnail_height"`
						ThumbnailURL    string `json:"thumbnail_url"`
						ThumbnailWidth  int    `json:"thumbnail_width"`
						Title           string `json:"title"`
						Type            string `json:"type"`
						URL             string `json:"url"`
						Version         string `json:"version"`
						Width           int    `json:"width"`
					} `json:"oembed"`
					Type string `json:"type"`
				} `json:"secure_media"`
				SecureMediaEmbed struct {
					Content   string `json:"content"`
					Height    int    `json:"height"`
					Scrolling bool   `json:"scrolling"`
					Width     int    `json:"width"`
				} `json:"secure_media_embed"`
				Selftext      string        `json:"selftext"`
				SelftextHTML  interface{}   `json:"selftext_html"`
				Stickied      bool          `json:"stickied"`
				Subreddit     string        `json:"subreddit"`
				SubredditID   string        `json:"subreddit_id"`
				SuggestedSort interface{}   `json:"suggested_sort"`
				Thumbnail     string        `json:"thumbnail"`
				Title         string        `json:"title"`
				Ups           int           `json:"ups"`
				URL           string        `json:"url"`
				UserReports   []interface{} `json:"user_reports"`
				Visited       bool          `json:"visited"`
			} `json:"data"`
			Kind string `json:"kind"`
		} `json:"children"`
		Modhash string `json:"modhash"`
	} `json:"data"`
	Kind string `json:"kind"`
}

func Debug() {

	//GetJsonCommentList("3sewvb")
	//return

	testJsonString := `[{"kind": "Listing", "data": {"modhash": "", "children": [{"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I am excited by the possibilities of Go. But I am need of a good resource to help me learn.&lt;/p&gt;\n\n&lt;p&gt;As someone with no Go background at all but a background in PHP and MySQL I am kind of understanding what I am reading on the Gin website but I could really do with a super comprehensive hold-my-hand guide on how to build a full website (not an api) with stuff like user logins, session management, CRSF, input validation etc all covered.&lt;/p&gt;\n\n&lt;p&gt;Any books or tutorials? I really want to learn Gin because of it&amp;#39;s high performance and it seems like I&amp;#39;ll learn more about under the hood compared to another framework?&lt;/p&gt;\n\n&lt;p&gt;It seems like a lot of packages have been written which are great. Is there a definitive list of packages? i.e. best one for handling CRSF, best one for sessions, best on for preventing XSS, validating JSON etc.&lt;/p&gt;\n\n&lt;p&gt;About my background I have been a web developer for 5 years and know html, css, php, javascript and use both windows and linux.&lt;/p&gt;\n\n&lt;p&gt;Thank you&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "I am excited by the possibilities of Go. But I am need of a good resource to help me learn.\n\nAs someone with no Go background at all but a background in PHP and MySQL I am kind of understanding what I am reading on the Gin website but I could really do with a super comprehensive hold-my-hand guide on how to build a full website (not an api) with stuff like user logins, session management, CRSF, input validation etc all covered.\n\nAny books or tutorials? I really want to learn Gin because of it's high performance and it seems like I'll learn more about under the hood compared to another framework?\n\nIt seems like a lot of packages have been written which are great. Is there a definitive list of packages? i.e. best one for handling CRSF, best one for sessions, best on for preventing XSS, validating JSON etc.\n\nAbout my background I have been a web developer for 5 years and know html, css, php, javascript and use both windows and linux.\n\nThank you\n\n", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sewvb", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "blackcoinprophet", "media": null, "name": "t3_3sewvb", "score": 4, "approved_by": null, "over_18": false, "hidden": false, "thumbnail": "", "subreddit_id": "t5_2rc7j", "edited": 1447257836.0, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "mod_reports": [], "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3sewvb/best_resource_to_learn_how_to_use_gin_framework/", "locked": false, "hide_score": false, "created": 1447284416.0, "url": "https://www.reddit.com/r/golang/comments/3sewvb/best_resource_to_learn_how_to_use_gin_framework/", "author_flair_text": null, "quarantine": false, "title": "Best resource to learn how to use Gin framework to build a website? (complete Go noob)", "created_utc": 1447255616.0, "ups": 4, "upvote_ratio": 0.83, "num_comments": 4, "visited": false, "num_reports": null, "distinguished": null}}], "after": null, "before": null}}, {"kind": "Listing", "data": {"modhash": "", "children": [{"kind": "t1", "data": {"subreddit_id": "t5_2rc7j", "banned_by": null, "removal_reason": null, "link_id": "t3_3sewvb", "likes": null, "replies": "", "user_reports": [], "saved": false, "id": "cwwoyaw", "gilded": 0, "archived": false, "report_reasons": null, "author": "koalefant", "parent_id": "t3_3sewvb", "score": 4, "approved_by": null, "controversiality": 0, "body": "not a framework, but the Gorilla toolkit is useful for composing elements of what you want [Gorilla Toolkit](http://www.gorillatoolkit.org). Used it myself when starting out and implemented the session management, logins etc in less than 20 lines of code.", "edited": false, "author_flair_css_class": null, "downs": 0, "body_html": "&lt;div class=\"md\"&gt;&lt;p&gt;not a framework, but the Gorilla toolkit is useful for composing elements of what you want &lt;a href=\"http://www.gorillatoolkit.org\"&gt;Gorilla Toolkit&lt;/a&gt;. Used it myself when starting out and implemented the session management, logins etc in less than 20 lines of code.&lt;/p&gt;\n&lt;/div&gt;", "subreddit": "golang", "score_hidden": false, "name": "t1_cwwoyaw", "created": 1447291691.0, "author_flair_text": null, "created_utc": 1447262891.0, "distinguished": null, "mod_reports": [], "num_reports": null, "ups": 4}}, {"kind": "t1", "data": {"subreddit_id": "t5_2rc7j", "banned_by": null, "removal_reason": null, "link_id": "t3_3sewvb", "likes": null, "replies": "", "user_reports": [], "saved": false, "id": "cwwnbhe", "gilded": 0, "archived": false, "report_reasons": null, "author": "yarrowy", "parent_id": "t3_3sewvb", "score": 1, "approved_by": null, "controversiality": 0, "body": "I was looking for something like this too but had no luck finding any.", "edited": false, "author_flair_css_class": null, "downs": 0, "body_html": "&lt;div class=\"md\"&gt;&lt;p&gt;I was looking for something like this too but had no luck finding any.&lt;/p&gt;\n&lt;/div&gt;", "subreddit": "golang", "score_hidden": false, "name": "t1_cwwnbhe", "created": 1447289214.0, "author_flair_text": null, "created_utc": 1447260414.0, "distinguished": null, "mod_reports": [], "num_reports": null, "ups": 1}}, {"kind": "t1", "data": {"subreddit_id": "t5_2rc7j", "banned_by": null, "removal_reason": null, "link_id": "t3_3sewvb", "likes": null, "replies": "", "user_reports": [], "saved": false, "id": "cwwr7da", "gilded": 0, "archived": false, "report_reasons": null, "author": "no1youknowz", "parent_id": "t3_3sewvb", "score": 1, "approved_by": null, "controversiality": 0, "body": "Try checking out the issues list.  There are a lot of people asking for help to do certain things.\n\nLike you, I am coming from a php background.  I got frustrated with gin as it doesn't come with MVC out of the gate.  Even though there are issues asking for this, it hasn't been implemented.", "edited": false, "author_flair_css_class": null, "downs": 0, "body_html": "&lt;div class=\"md\"&gt;&lt;p&gt;Try checking out the issues list.  There are a lot of people asking for help to do certain things.&lt;/p&gt;\n\n&lt;p&gt;Like you, I am coming from a php background.  I got frustrated with gin as it doesn&amp;#39;t come with MVC out of the gate.  Even though there are issues asking for this, it hasn&amp;#39;t been implemented.&lt;/p&gt;\n&lt;/div&gt;", "subreddit": "golang", "score_hidden": false, "name": "t1_cwwr7da", "created": 1447295048.0, "author_flair_text": null, "created_utc": 1447266248.0, "distinguished": null, "mod_reports": [], "num_reports": null, "ups": 1}}, {"kind": "t1", "data": {"subreddit_id": "t5_2rc7j", "banned_by": null, "removal_reason": null, "link_id": "t3_3sewvb", "likes": null, "replies": "", "user_reports": [], "saved": false, "id": "cwwymxc", "gilded": 0, "archived": false, "report_reasons": null, "author": "tvmaly", "parent_id": "t3_3sewvb", "score": 1, "approved_by": null, "controversiality": 0, "body": "I build the core services of bestfoodnearme.com with gin.  I would recommend looking at the examples and then the middleware examples.  Start out small and build just one thing.  use the standard template library.  I would be more than happy to answer any specific questions you have", "edited": false, "author_flair_css_class": null, "downs": 0, "body_html": "&lt;div class=\"md\"&gt;&lt;p&gt;I build the core services of bestfoodnearme.com with gin.  I would recommend looking at the examples and then the middleware examples.  Start out small and build just one thing.  use the standard template library.  I would be more than happy to answer any specific questions you have&lt;/p&gt;\n&lt;/div&gt;", "subreddit": "golang", "score_hidden": false, "name": "t1_cwwymxc", "created": 1447306193.0, "author_flair_text": null, "created_utc": 1447277393.0, "distinguished": null, "mod_reports": [], "num_reports": null, "ups": 1}}], "after": null, "before": null}}]`
	testJsonString = strings.Replace(testJsonString, `"edited": false`, `"edited": 0`, -1)
	testJson := []byte(testJsonString)
	var rpl RedditJsonCommentList
	err := json.Unmarshal(testJson, &rpl)
	if err != nil {
		fmt.Printf("Failed to parse: %s\n", err.Error())
	}
	fmt.Printf("Children len: %d\n", len(rpl[0].Data.Children))

	// Loop over posts and get the comments
	for index, child := range rpl[0].Data.Children {
		fmt.Printf("%d, Title: %s, ID: %s, Edited: %d\n", index, child.Data.Title, child.Data.ID, child.Data.Edited)
		//GetJsonCommentList(child.Data.ID)
	}
	fmt.Printf("Children 1 len: %d\n", len(rpl[1].Data.Children))

	// Loop over posts and get the comments
	for index, child := range rpl[1].Data.Children {
		fmt.Printf("%d, Title: %s, ID: %s, Edited: %d\n", index, child.Data.Title, child.Data.ID, child.Data.Edited)
		//GetJsonCommentList(child.Data.ID)
	}
	fmt.Println("exit")
	return
	/*
		testJsonString := `{"kind": "Listing", "data": {"modhash": "", "children": [{"kind": "t3", "data": {"domain": "davidnix.io", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sfjho", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "dgryski", "media": null, "score": 18, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 12, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sfjho/gos_error_handling_is_elegant/", "locked": false, "name": "t3_3sfjho", "created": 1447293782.0, "url": "http://davidnix.io/post/error-handling-in-go/", "author_flair_text": null, "quarantine": false, "title": "Go's Error Handling is Elegant", "created_utc": 1447264982.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 18}}, {"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I am excited by the possibilities of Go. But I am need of a good resource to help me learn.&lt;/p&gt;\n\n&lt;p&gt;As someone with no Go background at all but a background in PHP and MySQL I am kind of understanding what I am reading on the Gin website but I could really do with a super comprehensive hold-my-hand guide on how to build a full website (not an api) with stuff like user logins, session management, CRSF, input validation etc all covered.&lt;/p&gt;\n\n&lt;p&gt;Any books or tutorials? I really want to learn Gin because of it&amp;#39;s high performance and it seems like I&amp;#39;ll learn more about under the hood compared to another framework?&lt;/p&gt;\n\n&lt;p&gt;It seems like a lot of packages have been written which are great. Is there a definitive list of packages? i.e. best one for handling CRSF, best one for sessions, best on for preventing XSS, validating JSON etc.&lt;/p&gt;\n\n&lt;p&gt;About my background I have been a web developer for 5 years and know html, css, php, javascript and use both windows and linux.&lt;/p&gt;\n\n&lt;p&gt;Thank you&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "I am excited by the possibilities of Go. But I am need of a good resource to help me learn.\n\nAs someone with no Go background at all but a background in PHP and MySQL I am kind of understanding what I am reading on the Gin website but I could really do with a super comprehensive hold-my-hand guide on how to build a full website (not an api) with stuff like user logins, session management, CRSF, input validation etc all covered.\n\nAny books or tutorials? I really want to learn Gin because of it's high performance and it seems like I'll learn more about under the hood compared to another framework?\n\nIt seems like a lot of packages have been written which are great. Is there a definitive list of packages? i.e. best one for handling CRSF, best one for sessions, best on for preventing XSS, validating JSON etc.\n\nAbout my background I have been a web developer for 5 years and know html, css, php, javascript and use both windows and linux.\n\nThank you\n\n", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sewvb", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "blackcoinprophet", "media": null, "score": 4, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 3, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": 1447257836.0, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3sewvb/best_resource_to_learn_how_to_use_gin_framework/", "locked": false, "name": "t3_3sewvb", "created": 1447284416.0, "url": "https://www.reddit.com/r/golang/comments/3sewvb/best_resource_to_learn_how_to_use_gin_framework/", "author_flair_text": null, "quarantine": false, "title": "Best resource to learn how to use Gin framework to build a website? (complete Go noob)", "created_utc": 1447255616.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 4}}, {"kind": "t3", "data": {"domain": "sift-tool.org", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sc6k2", "from_kind": null, "gilded": 1, "archived": false, "clicked": false, "report_reasons": null, "author": "Rican7", "media": null, "score": 64, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 18, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sc6k2/sift_a_fast_and_powerful_open_source_alternative/", "locked": false, "name": "t3_3sc6k2", "created": 1447226283.0, "url": "https://sift-tool.org", "author_flair_text": null, "quarantine": false, "title": "sift - a fast and powerful open source alternative to grep", "created_utc": 1447197483.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 64}}, {"kind": "t3", "data": {"domain": "github.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sedv2", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "jcuga", "media": null, "score": 5, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 3, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sedv2/golongpoll_a_golang_http_longpolling_library/", "locked": false, "name": "t3_3sedv2", "created": 1447274093.0, "url": "https://github.com/jcuga/golongpoll", "author_flair_text": null, "quarantine": false, "title": "golongpoll: a golang HTTP longpolling library. Makes web pub-sub easy", "created_utc": 1447245293.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 5}}, {"kind": "t3", "data": {"domain": "stackoverflow.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3se9l4", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "HectorJ", "media": null, "score": 4, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3se9l4/is_there_any_way_to_use_mysql_temp_tables_in_go/", "locked": false, "name": "t3_3se9l4", "created": 1447270738.0, "url": "https://stackoverflow.com/q/33578271/1685538", "author_flair_text": null, "quarantine": false, "title": "Is there any way to use MySQL Temp Tables in Go?", "created_utc": 1447241938.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 4}}, {"kind": "t3", "data": {"domain": "blog.golang.org", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s9k1h", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "dgryski", "media": null, "score": 113, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 8, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s9k1h/six_years_of_go_the_go_blog/", "locked": false, "name": "t3_3s9k1h", "created": 1447187049.0, "url": "https://blog.golang.org/6years", "author_flair_text": null, "quarantine": false, "title": "Six years of Go - The Go Blog", "created_utc": 1447158249.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 113}}, {"kind": "t3", "data": {"domain": "github.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sbpcb", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "poitrus", "media": null, "score": 16, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 1, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sbpcb/globally_unique_id_generator_thought_for_the_web/", "locked": false, "name": "t3_3sbpcb", "created": 1447219490.0, "url": "https://github.com/rs/xid", "author_flair_text": null, "quarantine": false, "title": "Globally unique ID generator thought for the web", "created_utc": 1447190690.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 16}}, {"kind": "t3", "data": {"domain": "integralist.co.uk", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sahg7", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "outlearn", "media": null, "score": 12, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 5, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sahg7/http2/", "locked": false, "name": "t3_3sahg7", "created": 1447202747.0, "url": "http://www.integralist.co.uk/posts/http2.html", "author_flair_text": null, "quarantine": false, "title": "HTTP/2", "created_utc": 1447173947.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 12}}, {"kind": "t3", "data": {"domain": "reddit.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s9l6n", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "diegobernardes", "media": null, "score": 12, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 22, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s9l6n/which_language_has_the_brightest_future_in/", "locked": false, "name": "t3_3s9l6n", "created": 1447187851.0, "url": "https://www.reddit.com/r/rust/comments/3s6mxr/which_language_has_the_brightest_future_in/", "author_flair_text": null, "quarantine": false, "title": "Which language has the brightest future in replacement of C between D, Go and Rust? And Why?", "created_utc": 1447159051.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 12}}, {"kind": "t3", "data": {"domain": "speakerdeck.com", "banned_by": null, "media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F58c4a0600d1f4aa399125142f037c7ba&amp;url=https%3A%2F%2Fspeakerdeck.com%2Ffarslan%2Ftools-for-working-with-go-code&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F58c4a0600d1f4aa399125142f037c7ba%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 401}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": {"oembed": {"provider_url": "https://speakerdeck.com/", "description": "Go tools are special. Go was designed from the beginning to make tools easy to write. This caused the Go ecosystem to have dozens of well built tools. We highlight here some of the best and most used tools and show how to use them from an editor.", "title": "Tools for working with Go Code", "type": "rich", "thumbnail_width": 1024, "height": 401, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F58c4a0600d1f4aa399125142f037c7ba&amp;url=https%3A%2F%2Fspeakerdeck.com%2Ffarslan%2Ftools-for-working-with-go-code&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F58c4a0600d1f4aa399125142f037c7ba%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "author_name": "Fatih Arslan", "version": "1.0", "provider_name": "Speaker Deck", "thumbnail_url": "https://i.embed.ly/1/image?url=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F58c4a0600d1f4aa399125142f037c7ba%2Fslide_0.jpg&amp;key=b1e305db91cf4aa5a86b732cc9fffceb", "thumbnail_height": 576, "author_url": "https://speakerdeck.com/farslan"}, "type": "speakerdeck.com"}, "link_flair_text": null, "id": "3s9syw", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "farslan", "media": {"oembed": {"provider_url": "https://speakerdeck.com/", "description": "Go tools are special. Go was designed from the beginning to make tools easy to write. This caused the Go ecosystem to have dozens of well built tools. We highlight here some of the best and most used tools and show how to use them from an editor.", "title": "Tools for working with Go Code", "type": "rich", "thumbnail_width": 1024, "height": 401, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F58c4a0600d1f4aa399125142f037c7ba&amp;url=https%3A%2F%2Fspeakerdeck.com%2Ffarslan%2Ftools-for-working-with-go-code&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F58c4a0600d1f4aa399125142f037c7ba%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "author_name": "Fatih Arslan", "version": "1.0", "provider_name": "Speaker Deck", "thumbnail_url": "https://speakerd.s3.amazonaws.com/presentations/58c4a0600d1f4aa399125142f037c7ba/slide_0.jpg", "thumbnail_height": 576, "author_url": "https://speakerdeck.com/farslan"}, "type": "speakerdeck.com"}, "score": 5, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 7, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F58c4a0600d1f4aa399125142f037c7ba&amp;url=https%3A%2F%2Fspeakerdeck.com%2Ffarslan%2Ftools-for-working-with-go-code&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F58c4a0600d1f4aa399125142f037c7ba%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 401}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s9syw/tools_for_working_with_go_code/", "locked": false, "name": "t3_3s9syw", "created": 1447192208.0, "url": "https://speakerdeck.com/farslan/tools-for-working-with-go-code", "author_flair_text": null, "quarantine": false, "title": "Tools for working with Go Code", "created_utc": 1447163408.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 5}}, {"kind": "t3", "data": {"domain": "github.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3sbsiz", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "adililhan", "media": null, "score": 0, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3sbsiz/zabbix_desktop_notification_with_golang/", "locked": false, "name": "t3_3sbsiz", "created": 1447220707.0, "url": "https://github.com/adililhan/Zabbix-Desktop-Notification-with-Golang", "author_flair_text": null, "quarantine": false, "title": "Zabbix Desktop Notification with Golang", "created_utc": 1447191907.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 0}}, {"kind": "t3", "data": {"domain": "speakerdeck.com", "banned_by": null, "media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F0feaedadd2ed4ec280698ea13405092f&amp;url=https%3A%2F%2Fspeakerdeck.com%2Fcampoy%2Ffunctional-go&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F0feaedadd2ed4ec280698ea13405092f%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 401}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": {"type": "speakerdeck.com", "oembed": {"provider_url": "https://speakerdeck.com/", "description": "In this talk I discuss whether applying some of the principles of programming language to Go is possible or makes sense. Can we build the Maybe and Many monads in Go? Well, yes! Is it a good idea? Well, probably not Presented at http://dotgo.eu 2015 in Paris", "title": "Functional Go", "thumbnail_width": 1024, "height": 401, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F0feaedadd2ed4ec280698ea13405092f&amp;url=https%3A%2F%2Fspeakerdeck.com%2Fcampoy%2Ffunctional-go&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F0feaedadd2ed4ec280698ea13405092f%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "author_name": "Francesc Campoy Flores", "version": "1.0", "provider_name": "Speaker Deck", "thumbnail_url": "https://i.embed.ly/1/image?url=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F0feaedadd2ed4ec280698ea13405092f%2Fslide_0.jpg&amp;key=b1e305db91cf4aa5a86b732cc9fffceb", "type": "rich", "thumbnail_height": 576, "author_url": "https://speakerdeck.com/campoy"}}, "link_flair_text": null, "id": "3s9a1n", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "campoy", "media": {"type": "speakerdeck.com", "oembed": {"provider_url": "https://speakerdeck.com/", "description": "In this talk I discuss whether applying some of the principles of programming language to Go is possible or makes sense. Can we build the Maybe and Many monads in Go? Well, yes! Is it a good idea? Well, probably not Presented at http://dotgo.eu 2015 in Paris", "title": "Functional Go", "thumbnail_width": 1024, "height": 401, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F0feaedadd2ed4ec280698ea13405092f&amp;url=https%3A%2F%2Fspeakerdeck.com%2Fcampoy%2Ffunctional-go&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F0feaedadd2ed4ec280698ea13405092f%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "author_name": "Francesc Campoy Flores", "version": "1.0", "provider_name": "Speaker Deck", "thumbnail_url": "https://speakerd.s3.amazonaws.com/presentations/0feaedadd2ed4ec280698ea13405092f/slide_0.jpg", "type": "rich", "thumbnail_height": 576, "author_url": "https://speakerdeck.com/campoy"}}, "score": 8, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fspeakerdeck.com%2Fplayer%2F0feaedadd2ed4ec280698ea13405092f&amp;url=https%3A%2F%2Fspeakerdeck.com%2Fcampoy%2Ffunctional-go&amp;image=https%3A%2F%2Fspeakerd.s3.amazonaws.com%2Fpresentations%2F0feaedadd2ed4ec280698ea13405092f%2Fslide_0.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=speakerdeck\" width=\"600\" height=\"401\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 401}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s9a1n/functional_go_speaker_deck/", "locked": false, "name": "t3_3s9a1n", "created": 1447179490.0, "url": "https://speakerdeck.com/campoy/functional-go", "author_flair_text": null, "quarantine": false, "title": "Functional Go // Speaker Deck", "created_utc": 1447150690.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 8}}, {"kind": "t3", "data": {"domain": "github.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s8pv6", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "y4m4b4", "media": null, "score": 11, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 4, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s8pv6/minio_go_library_for_amazon_s3_compatible_cloud/", "locked": false, "name": "t3_3s8pv6", "created": 1447164915.0, "url": "https://github.com/minio/minio-go", "author_flair_text": null, "quarantine": false, "title": "Minio Go Library for Amazon S3 Compatible Cloud Storage.", "created_utc": 1447136115.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 11}}, {"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;Hi guys I&amp;#39;m looking for a mentor to help me with a project I&amp;#39;m currently wanting to do.&lt;/p&gt;\n\n&lt;p&gt;I would like to recreate &lt;a href=\"https://github.com/joshuaferrara/node-csgo\"&gt;node-csgo&lt;/a&gt; in golang, it is a plugin for &lt;a href=\"https://github.com/seishun/node-steam\"&gt;node-steam&lt;/a&gt;.&lt;/p&gt;\n\n&lt;p&gt;&lt;a href=\"https://github.com/seishun/node-steam\"&gt;node-steam&lt;/a&gt; is already programmed in Go here: &lt;a href=\"https://github.com/Philipp15b/go-steam\"&gt;go-steam&lt;/a&gt; so basically I want to recreate the &lt;a href=\"https://github.com/joshuaferrara/node-csgo\"&gt;node-csgo&lt;/a&gt; plugin into go so that I can use it with the &lt;a href=\"https://github.com/Philipp15b/go-steam\"&gt;go-steam&lt;/a&gt; package but I haven&amp;#39;t the first clue on how to convert the javascript into golang, though I am well off with javascript, but I don&amp;#39;t know what packages would really be used and looking through  &lt;a href=\"https://github.com/Philipp15b/go-steam\"&gt;go-steam&lt;/a&gt; I don&amp;#39;t know what packages and what not to use to accomplish this. I am very new to Go and I would like somebody to help me along, I could pay if necessary for your services as this would be a very good learning experience. Anyways if anybody is interested please PM me or post here with any pricing or if you just want to help.&lt;/p&gt;\n\n&lt;p&gt;I was thinking we could communicate via VOIP (Skype, Mumble, Ventrilo, w/e) and share screens or something. Anyways, if anybody is interested let me know. Thanks.&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "Hi guys I'm looking for a mentor to help me with a project I'm currently wanting to do.\n\nI would like to recreate [node-csgo](https://github.com/joshuaferrara/node-csgo) in golang, it is a plugin for [node-steam](https://github.com/seishun/node-steam).\n\n[node-steam](https://github.com/seishun/node-steam) is already programmed in Go here: [go-steam](https://github.com/Philipp15b/go-steam) so basically I want to recreate the [node-csgo](https://github.com/joshuaferrara/node-csgo) plugin into go so that I can use it with the [go-steam](https://github.com/Philipp15b/go-steam) package but I haven't the first clue on how to convert the javascript into golang, though I am well off with javascript, but I don't know what packages would really be used and looking through  [go-steam](https://github.com/Philipp15b/go-steam) I don't know what packages and what not to use to accomplish this. I am very new to Go and I would like somebody to help me along, I could pay if necessary for your services as this would be a very good learning experience. Anyways if anybody is interested please PM me or post here with any pricing or if you just want to help.\n\nI was thinking we could communicate via VOIP (Skype, Mumble, Ventrilo, w/e) and share screens or something. Anyways, if anybody is interested let me know. Thanks.", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s8c6c", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "Gacnt", "media": null, "score": 4, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 4, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": 1447129338.0, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3s8c6c/looking_for_a_mentor/", "locked": false, "name": "t3_3s8c6c", "created": 1447157952.0, "url": "https://www.reddit.com/r/golang/comments/3s8c6c/looking_for_a_mentor/", "author_flair_text": null, "quarantine": false, "title": "Looking for a mentor.", "created_utc": 1447129152.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 4}}, {"kind": "t3", "data": {"domain": "youtube.com", "banned_by": null, "media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FGVQzgy8AD30%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DGVQzgy8AD30%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FGVQzgy8AD30%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "description": "Golang Send Email - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang Send Email", "url": "http://www.youtube.com/watch?v=GVQzgy8AD30", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FGVQzgy8AD30%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DGVQzgy8AD30%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FGVQzgy8AD30%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.embed.ly/1/image?url=https%3A%2F%2Fi.ytimg.com%2Fvi%2FGVQzgy8AD30%2Fhqdefault.jpg&amp;key=b1e305db91cf4aa5a86b732cc9fffceb", "type": "video", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}}, "link_flair_text": null, "id": "3s8jpu", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "tscottmcleod", "media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "description": "Golang Send Email - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang Send Email", "url": "http://www.youtube.com/watch?v=GVQzgy8AD30", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FGVQzgy8AD30%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DGVQzgy8AD30%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FGVQzgy8AD30%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.ytimg.com/vi/GVQzgy8AD30/hqdefault.jpg", "type": "video", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}}, "score": 3, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 1, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FGVQzgy8AD30%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DGVQzgy8AD30%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FGVQzgy8AD30%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s8jpu/golang_send_email/", "locked": false, "name": "t3_3s8jpu", "created": 1447161629.0, "url": "https://www.youtube.com/attribution_link?a=mw0Clf1b0YE&amp;u=%2Fwatch%3Fv%3DGVQzgy8AD30%26feature%3Dshare", "author_flair_text": null, "quarantine": false, "title": "Golang Send Email", "created_utc": 1447132829.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 3}}, {"kind": "t3", "data": {"domain": "reddit.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s8w80", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "Mittalmailbox", "media": null, "score": 2, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s8w80/xposti_made_a_sublime_package_for_hugo_rgohugo/", "locked": false, "name": "t3_3s8w80", "created": 1447168892.0, "url": "https://www.reddit.com/r/gohugo/comments/3s8u0y/i_made_a_sublime_package_for_hugo/?ref=share&amp;ref_source=link", "author_flair_text": null, "quarantine": false, "title": "[X-Post]I made a sublime package for Hugo \u2022 /r/gohugo", "created_utc": 1447140092.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 2}}, {"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I&amp;#39;m just getting started with Go, looking for good open source projects using Go for API backend (with database). Not really looking to dive into frameworks right now, something with just standard library would be great. Any links to tutorial(s) would also be great.&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "I'm just getting started with Go, looking for good open source projects using Go for API backend (with database). Not really looking to dive into frameworks right now, something with just standard library would be great. Any links to tutorial(s) would also be great.", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s6j1m", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "akshayaurora", "media": null, "score": 13, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 9, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3s6j1m/developing_api_backend_in_go/", "locked": false, "name": "t3_3s6j1m", "created": 1447131109.0, "url": "https://www.reddit.com/r/golang/comments/3s6j1m/developing_api_backend_in_go/", "author_flair_text": null, "quarantine": false, "title": "Developing API backend in Go", "created_utc": 1447102309.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 13}}, {"kind": "t3", "data": {"domain": "youtube.com", "banned_by": null, "media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FFwDa-naV_ls%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DFwDa-naV_ls%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FFwDa-naV_ls%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": {"oembed": {"provider_url": "https://www.youtube.com/", "description": "Golang Datastore - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang Datastore", "url": "http://www.youtube.com/watch?v=FwDa-naV_ls", "type": "video", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FFwDa-naV_ls%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DFwDa-naV_ls%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FFwDa-naV_ls%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.embed.ly/1/image?url=https%3A%2F%2Fi.ytimg.com%2Fvi%2FFwDa-naV_ls%2Fhqdefault.jpg&amp;key=b1e305db91cf4aa5a86b732cc9fffceb", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}, "type": "youtube.com"}, "link_flair_text": null, "id": "3s8rhg", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "tscottmcleod", "media": {"oembed": {"provider_url": "https://www.youtube.com/", "description": "Golang Datastore - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang Datastore", "url": "http://www.youtube.com/watch?v=FwDa-naV_ls", "type": "video", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FFwDa-naV_ls%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DFwDa-naV_ls%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FFwDa-naV_ls%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.ytimg.com/vi/FwDa-naV_ls/hqdefault.jpg", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}, "type": "youtube.com"}, "score": 2, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 1, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FFwDa-naV_ls%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DFwDa-naV_ls%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FFwDa-naV_ls%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s8rhg/golang_datastore/", "locked": false, "name": "t3_3s8rhg", "created": 1447165882.0, "url": "https://www.youtube.com/attribution_link?a=TW-e6nRoc5E&amp;u=%2Fwatch%3Fv%3DFwDa-naV_ls%26feature%3Dshare", "author_flair_text": null, "quarantine": false, "title": "Golang Datastore", "created_utc": 1447137082.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 2}}, {"kind": "t3", "data": {"domain": "youtube.com", "banned_by": null, "media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FM_miFP9N-w0%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DM_miFP9N-w0%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FM_miFP9N-w0%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "description": "golang image jpeg processing - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang image analysis", "url": "http://www.youtube.com/watch?v=M_miFP9N-w0", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FM_miFP9N-w0%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DM_miFP9N-w0%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FM_miFP9N-w0%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.embed.ly/1/image?url=https%3A%2F%2Fi.ytimg.com%2Fvi%2FM_miFP9N-w0%2Fhqdefault.jpg&amp;key=b1e305db91cf4aa5a86b732cc9fffceb", "type": "video", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}}, "link_flair_text": null, "id": "3s8kzs", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "tscottmcleod", "media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "description": "golang image jpeg processing - Best Golang books: http://amzn.to/1RIM5HP http://amzn.to/1kGGsPv", "title": "Golang image analysis", "url": "http://www.youtube.com/watch?v=M_miFP9N-w0", "author_name": "Todd McLeod", "height": 450, "width": 600, "html": "&lt;iframe class=\"embedly-embed\" src=\"//cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FM_miFP9N-w0%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DM_miFP9N-w0%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FM_miFP9N-w0%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "thumbnail_width": 480, "version": "1.0", "provider_name": "YouTube", "thumbnail_url": "https://i.ytimg.com/vi/M_miFP9N-w0/hqdefault.jpg", "type": "video", "thumbnail_height": 360, "author_url": "https://www.youtube.com/user/toddmcleod"}}, "score": 2, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {"content": "&lt;iframe class=\"embedly-embed\" src=\"https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FM_miFP9N-w0%3Ffeature%3Doembed&amp;url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DM_miFP9N-w0%26feature%3Dshare&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FM_miFP9N-w0%2Fhqdefault.jpg&amp;key=2aa3c4d5f3de4f5b9120b660ad850dc9&amp;type=text%2Fhtml&amp;schema=youtube\" width=\"600\" height=\"450\" scrolling=\"no\" frameborder=\"0\" allowfullscreen&gt;&lt;/iframe&gt;", "width": 600, "scrolling": false, "height": 450}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s8kzs/golang_image_analysis/", "locked": false, "name": "t3_3s8kzs", "created": 1447162240.0, "url": "https://www.youtube.com/attribution_link?a=GWjLJ4SbBKs&amp;u=%2Fwatch%3Fv%3DM_miFP9N-w0%26feature%3Dshare", "author_flair_text": null, "quarantine": false, "title": "Golang image analysis", "created_utc": 1447133440.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 2}}, {"kind": "t3", "data": {"domain": "outlearn.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s5y24", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "outlearn", "media": null, "score": 14, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 2, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s5y24/learn_go_from_scratch/", "locked": false, "name": "t3_3s5y24", "created": 1447123015.0, "url": "https://www.outlearn.com/learn/matryer/golang-from-scratch", "author_flair_text": null, "quarantine": false, "title": "Learn Go from Scratch", "created_utc": 1447094215.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 14}}, {"kind": "t3", "data": {"domain": "github.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s63ao", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "cbarrick", "media": null, "score": 13, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s63ao/evo_genetic_evolutionary_algorithms_in_go/", "locked": false, "name": "t3_3s63ao", "created": 1447125012.0, "url": "https://github.com/cbarrick/evo", "author_flair_text": null, "quarantine": false, "title": "Evo: Genetic &amp; Evolutionary Algorithms in Go", "created_utc": 1447096212.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 13}}, {"kind": "t3", "data": {"domain": "getgb.io", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s3p14", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "lstokeworth", "media": null, "score": 22, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 4, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s3p14/gb_version_031_released/", "locked": false, "name": "t3_3s3p14", "created": 1447077745.0, "url": "http://getgb.io/news/gb-version-0.3.1-released/", "author_flair_text": null, "quarantine": false, "title": "gb version 0.3.1 released", "created_utc": 1447048945.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 22}}, {"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I&amp;#39;ve created a very small queue library over Redis, but this is my first &amp;quot;complete&amp;quot; Go project and I don&amp;#39;t feel it very idiomatic. (Especially the testing, maybe I should mock the client and forget the helper...)&lt;/p&gt;\n\n&lt;p&gt;Anyway, I&amp;#39;m open for every tips and helps how to make this more &amp;quot;goish&amp;quot;. :)&lt;/p&gt;\n\n&lt;p&gt;&lt;a href=\"https://github.com/Gerifield/go-little-red-queue\"&gt;https://github.com/Gerifield/go-little-red-queue&lt;/a&gt;&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "I've created a very small queue library over Redis, but this is my first \"complete\" Go project and I don't feel it very idiomatic. (Especially the testing, maybe I should mock the client and forget the helper...)\n\nAnyway, I'm open for every tips and helps how to make this more \"goish\". :)\n\nhttps://github.com/Gerifield/go-little-red-queue\n", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s7dlj", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "gergo254", "media": null, "score": 0, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3s7dlj/yet_another_queue_over_redis/", "locked": false, "name": "t3_3s7dlj", "created": 1447143104.0, "url": "https://www.reddit.com/r/golang/comments/3s7dlj/yet_another_queue_over_redis/", "author_flair_text": null, "quarantine": false, "title": "Yet another queue over redis", "created_utc": 1447114304.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 0}}, {"kind": "t3", "data": {"domain": "medium.com", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": null, "selftext": "", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s1ngp", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "saturnism2", "media": null, "score": 45, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 15, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": false, "from_id": null, "permalink": "/r/golang/comments/3s1ngp/my_go_ide_in_a_container/", "locked": false, "name": "t3_3s1ngp", "created": 1447043731.0, "url": "https://medium.com/google-cloud/my-ide-in-a-container-49d4f177de", "author_flair_text": null, "quarantine": false, "title": "My Go IDE in a Container", "created_utc": 1447014931.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 45}}, {"kind": "t3", "data": {"domain": "self.golang", "banned_by": null, "media_embed": {}, "subreddit": "golang", "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I was wondering if you would help my with my final year project at university by filling in a short survey (&lt;a href=\"https://clarkeash.typeform.com/to/vKeHjW\"&gt;https://clarkeash.typeform.com/to/vKeHjW&lt;/a&gt;) on how you setup you server and deploy your apps.&lt;/p&gt;\n\n&lt;p&gt;Your help and time would be very appreciated&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "selftext": "I was wondering if you would help my with my final year project at university by filling in a short survey (https://clarkeash.typeform.com/to/vKeHjW) on how you setup you server and deploy your apps.\n\nYour help and time would be very appreciated", "likes": null, "suggested_sort": null, "user_reports": [], "secure_media": null, "link_flair_text": null, "id": "3s6ii1", "from_kind": null, "gilded": 0, "archived": false, "clicked": false, "report_reasons": null, "author": "clarkeash", "media": null, "score": 0, "approved_by": null, "over_18": false, "hidden": false, "num_comments": 0, "thumbnail": "", "subreddit_id": "t5_2rc7j", "hide_score": false, "edited": false, "link_flair_css_class": null, "author_flair_css_class": null, "downs": 0, "secure_media_embed": {}, "saved": false, "removal_reason": null, "stickied": false, "from": null, "is_self": true, "from_id": null, "permalink": "/r/golang/comments/3s6ii1/server_setup_and_application_deployments/", "locked": false, "name": "t3_3s6ii1", "created": 1447130887.0, "url": "https://www.reddit.com/r/golang/comments/3s6ii1/server_setup_and_application_deployments/", "author_flair_text": null, "quarantine": false, "title": "Server setup and application deployments", "created_utc": 1447102087.0, "distinguished": null, "mod_reports": [], "visited": false, "num_reports": null, "ups": 0}}], "after": "t3_3s6ii1", "before": null}}`
		testJsonString = strings.Replace(testJsonString, `"edited": false`, `"edited": 0`, -1)
		testJson := []byte(testJsonString)
		var rpl RedditJsonCommentList
		err := json.Unmarshal(testJson, &rpl)
		if err != nil {
			fmt.Printf("Failed to parse: %s\n", err.Error())
		}
		fmt.Printf("Children len: %d\n", len(rpl.Data.Children))

		// Loop over posts and get the comments
		for index, child := range rpl.Data.Children {
			fmt.Printf("%d, Title: %s, ID: %s, Edited: %d\n", index, child.Data.Title, child.Data.ID, child.Data.Edited)
			//GetJsonCommentList(child.Data.ID)
		}
		fmt.Println("exit")
		return
	*/
}

func main() {

	Debug()
	return

	//uri := "https://www.reddit.com/r/golang/controversial.json"
	uri := "https://www.reddit.com/r/golang.json"
	fmt.Println("fetching", uri)
	redditPostList, err := GetJsonPostList(uri)
	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		fmt.Println(err)
		return
	}

	fmt.Printf("Children len: %d\n", len(redditPostList.Data.Children))

	// Loop over posts and get the comments
	for index, child := range redditPostList.Data.Children {
		fmt.Printf("%d, Title: %s, ID: %s\n", index, child.Data.Title, child.Data.ID)
		//GetJsonCommentList(child.Data.ID)
	}
	fmt.Println("exit", uri)
}

func GetJsonCommentList(ID string) (redditCommentList *RedditJsonCommentList, err error) {

	// Get data from url
	uri := fmt.Sprintf("https://www.reddit.com/r/golang/comments/%s.json", ID)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)

	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()

		// capture all bytes from the response body
		buf, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 { // 200 = OK
			httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", uri, resp.StatusCode, resp.Status)
			err = errors.New(httperr)
			return nil, err
		}
		var rpl RedditJsonCommentList
		err = json.Unmarshal(buf, &rpl)
		fmt.Println(string(buf))
		return &rpl, err

	} else {
		err = errors.New("Response from " + uri + " is nil")
		return nil, err
	}

	return nil, errors.New("Uncatched error in GetJsonPostList")
}

func GetJsonPostList(uri string) (redditPostList *RedditJsonPostList, err error) {

	// Get data from url
	//resp, err := http.Get(url)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)

	if err != nil {
		err = errors.New("Failed to http.Get from " + uri + ": " + err.Error())
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()

		// capture all bytes from the response body
		buf, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 { // 200 = OK
			httperr := fmt.Sprintf("Failed to http.Get from %s: Http Status code: %d: Msg: %s", uri, resp.StatusCode, resp.Status)
			err = errors.New(httperr)
			return nil, err
		}
		// Fix Reddit API type mismatch error
		// edited is reported as false(bool) if it really should be 0(float)
		// if the post was not edited after creation
		buf = bytes.Replace(buf, []byte(`"edited": false`), []byte(`"edited": 0`), -1)

		var rpl RedditJsonPostList
		fmt.Println(string(buf))
		err = json.Unmarshal(buf, &rpl)
		return &rpl, err

	} else {
		err = errors.New("Response from " + uri + " is nil")
		return nil, err
	}

	return nil, errors.New("Uncatched error in GetJsonPostList")
}
