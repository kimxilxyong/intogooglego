package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
	doc, err := goquery.NewDocument("http://www.reddit.com/r/golang")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".entry").Each(func(i int, s *goquery.Selection) {
		band := s.Find(".a .title").Text()
		title := s.Find("span").Text()
		fmt.Printf("Review %d: %s **** %s\n", i, band, title)
		//fmt.Println(s.String())
	})

	sel := doc.Find(".entry")
	for i := range sel.Nodes {
		single := sel.Eq(i)
		// use `single` as a selection of 1 node
		html, err := single.Html()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Lenght ---------------- ", single.Length())
		fmt.Println(html)
	}
}

func main() {
	ExampleScrape()
}
