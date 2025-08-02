package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type pageData struct {
	url          string
	textAndLinks map[string]string
}

func pageDataGenerator(url string) *pageData {
	p := pageData{url: url, textAndLinks: make(map[string]string)}
	return &p
}

func (p *pageData) printFormatedMap() {
	if len(p.textAndLinks) == 0 {
		fmt.Println("the data of links and texts is empty")
	}
	for k, v := range p.textAndLinks {
		fmt.Printf("%s \t \t %s\n", k, v)
	}
}

func main() {
	pageBeingCrawled := pageDataGenerator("https://books.toscrape.com")
	baseUrl, err := url.Parse(pageBeingCrawled.url)
	if err != nil {
		log.Fatalf("there was an error trying to parse the baseUrl: %s", err)
	}
	resp, err := http.Get(pageBeingCrawled.url)
	if err != nil {
		log.Fatalf("there was an error trying to perform a get on the baseUrl %s", err)
	}
	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken {
			token := z.Token()
			if token.Data == "a" {
				var link string
				tokenAttributes := token.Attr
				for _, value := range tokenAttributes {
					if value.Key == "href" {
						anchorUrl, err := url.Parse(value.Val)
						if err != nil {
							log.Printf("skipping malformed href: %s", err)
							continue
						}
						link = baseUrl.ResolveReference(anchorUrl).String()
					}
				}
				nextToken := z.Next()
				if nextToken == html.TextToken && link != "" {
					trimmedText := strings.TrimSpace(string(z.Text()))
					if trimmedText != "" {
						pageBeingCrawled.textAndLinks[trimmedText] = link
					}
				}
			}
		}
	}
	fmt.Println(len(pageBeingCrawled.textAndLinks))
	pageBeingCrawled.printFormatedMap()
}
