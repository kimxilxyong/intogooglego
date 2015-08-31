// Copyright 2015 Kim Il
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A HTML to markdown converter

package markdownify

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func HtmlToMarkdown(htmlInput string) (markdownResult string, err error) {

	// Get starttime for measuring how long this functions takes
	timeStart := time.Now()

	fmt.Printf("Start HtmlToMarkdown - ")

	serviceEndPoint := "http://heckyesmarkdown.com/go/"
	postParams := url.Values{}
	postParams.Set("html", htmlInput) // the html input string
	postParams.Set("read", "0")       // turn readability off, default is on
	postParams.Set("md", "1")         // Run Markdownify, default on

	timeout := time.Duration(30 * time.Second)
	client := &http.Client{}
	client.Timeout = timeout

	resp, err := client.PostForm(serviceEndPoint, postParams)
	if err != nil {
		requestDuration := (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))
		fmt.Printf("HtmlToMarkdown ERROR %s, duration %d\n", err.Error(), requestDuration)
		return
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	markdownResult = string(body)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", resp.Status)
	}

	requestDuration := (time.Since(timeStart).Nanoseconds() / int64(time.Millisecond))
	fmt.Printf("HtmlToMarkdown duration %d\n", requestDuration)
	return
}
