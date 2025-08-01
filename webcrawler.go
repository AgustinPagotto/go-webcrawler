package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	var urlToCrawl string = "https://books.toscrape.com"
	baseUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		log.Fatalf("there was an error trying to parse the baseUrl: %s", err)
	}
	resp, err := http.Get(urlToCrawl)
	if err != nil {
		log.Fatalf("there was an error trying to perform a get on the baseUrl %s", err)
	}
	defer resp.Body.Close()
	textAndLinksMap := make(map[string]string)
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
						textAndLinksMap[trimmedText] = link
					}
				}
			}
		}
	}
	fmt.Println(len(textAndLinksMap))
}
