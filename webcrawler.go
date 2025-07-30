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
	baseUrl, _ := url.Parse(urlToCrawl)
	resp, err := http.Get(urlToCrawl)
	if err != nil {
		log.Fatalf("there was an error trying to get to the url %s", err)
	}
	defer resp.Body.Close()
	var textInsideATags []string
	z := html.NewTokenizer(resp.Body)
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			return
		}
		if tokenType == html.StartTagToken {
			token := z.Token()
			if token.Data == "a" {
				nextToken := z.Next()
				tokenAttributes := token.Attr
				for _, value := range tokenAttributes {
					if value.Key == "href" {
						//fmt.Println(value.Val)
						anchorUrl, _ := url.Parse(value.Val)
						fmt.Println(baseUrl.ResolveReference(anchorUrl))
					}
				}
				if nextToken == html.TextToken {
					trimmedText := strings.TrimSpace(string(z.Text()))
					textInsideATags = append(textInsideATags, trimmedText)
				}
			}
		}
		//fmt.Println(len(textInsideATags))
	}
}
