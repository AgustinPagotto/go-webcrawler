package main

import (
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
)

func main() {
	page, err := crawl.CrawlPage("http://github.com")
	if err != nil {
		fmt.Printf("Error crawling page: %s\n", err)
	}
	fmt.Println(page)
}
