package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var visited = make(map[string]bool)

var depth = 0

func main() {
	flag.Parse()

	args := flag.Args()
	fmt.Println(args)
	if len(args) < 1 {
		fmt.Println("Please specify start page")
		os.Exit(1)
	}

	queue := make(chan string)

	go func() { queue <- args[0] }()

	for uri := range queue {
		enqueue(uri, queue)
	}
}

func enqueue(uri string, queue chan string) {
	fmt.Println("fetching", uri)

	depth++

	fmt.Println("depth", depth)

	visited[uri] = true

	defer func() { depth-- }()

	if depth > 1 {
		return
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	links := collectlinks.All(resp.Body)

	for _, link := range links {
		absolute := fixUrl(link, uri)
		//fmt.Println("absolute", absolute)
		if (uri != "") && (strings.HasPrefix(absolute, uri)) {
			if !visited[absolute] {
				go func() {
					queue <- absolute
					fmt.Println("added", absolute)
				}()
			}
		}
	}
	fmt.Println("exit", uri)
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}
