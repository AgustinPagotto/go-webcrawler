package main

import (
	"flag"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	"log"
	"net/url"
)

func main() {
	var urlFromCli string
	var depthCrawl int
	flag.StringVar(&urlFromCli, "url", "http://google.com", "Url to be Crawled")
	flag.StringVar(&urlFromCli, "u", "http://google.com", "Url to be Crawled")
	flag.IntVar(&depthCrawl, "depth", 1, "Depth of the crawl")
	flag.IntVar(&depthCrawl, "d", 1, "Depth of the crawl")
	flag.Parse()
	if depthCrawl < 1 {
		log.Fatalf("The depth of crawl can't be less than 1")
	}
	if urlFromCli == "" {
		log.Fatalf("Please provide a non empty url")
	}
	parsedUrl, err := url.ParseRequestURI(urlFromCli)
	if err != nil {
		log.Fatalf("url inserted is invalid: %s", err)
	}
	page, err := crawl.CrawlPage(parsedUrl.String())
	if err != nil {
		log.Fatalf("Error crawling page: %s\n", err)
	}
	fmt.Println(page)
}
