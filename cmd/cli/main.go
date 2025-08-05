package main

import (
	"flag"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	"log"
)

func receiveFlags() (string, int) {
	var urlFromCli string
	var depthCrawl int
	flag.StringVar(&urlFromCli, "url", "http://google.com", "Url to be Crawled")
	flag.StringVar(&urlFromCli, "u", "http://google.com", "Url to be Crawled")
	flag.IntVar(&depthCrawl, "depth", 1, "Depth of the crawl")
	flag.IntVar(&depthCrawl, "d", 1, "Depth of the crawl")
	flag.Parse()
	return urlFromCli, depthCrawl
}

func validateFlags(url string, depth int) error {
	if depth < 1 {
		return fmt.Errorf("The depth of crawl can't be less than 1")
	}
	if url == "" {
		return fmt.Errorf("Please provide a non-empty url")
	}
	return nil
}

func main() {
	urlToCrawl, depthCrawl := receiveFlags()
	if err := validateFlags(urlToCrawl, depthCrawl); err != nil {
		log.Fatal(err)
	}
	page, err := crawl.CrawlPage(urlToCrawl)
	if err != nil {
		log.Fatalf("Error crawling page: %s\n", err)
	}
	fmt.Println(page)
}
