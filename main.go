package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	"github.com/AgustinPagotto/go-webcrawler/internal/db"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
)

func receiveFlags() (string, int, string) {
	var urlFromCli string
	var depthCrawl int
	var searchTerm string
	flag.StringVar(&urlFromCli, "url", "", "Url to be Crawled")
	flag.StringVar(&urlFromCli, "u", "", "Url to be Crawled")
	flag.IntVar(&depthCrawl, "depth", 1, "Depth of the crawl")
	flag.IntVar(&depthCrawl, "d", 1, "Depth of the crawl")
	flag.StringVar(&searchTerm, "search", "", "Word to Search")
	flag.StringVar(&searchTerm, "s", "", "Word to Search")
	flag.Parse()
	return urlFromCli, depthCrawl, searchTerm
}

func main() {
	urlToCrawl, depthCrawl, searchTerm := receiveFlags()
	searchBool, err := validate.ValidateFlags(urlToCrawl, depthCrawl, searchTerm)
	if err != nil {
		log.Fatal(err)
	}
	store, err := db.New()
	err = store.InitiateDB()
	if err != nil {
		log.Fatalf("Error adding tables to db: %s\n", err)
	}
	defer store.Close()
	if searchBool {
		performSearch(store, searchTerm)
	} else {
		performCrawl(store, urlToCrawl, depthCrawl)
	}
}

func performCrawl(store *db.Store, urlToCrawl string, depthCrawl int) {
	var needsRecrawl bool
	const recrawlAfter = 1 * 24 * time.Hour
	crawler, err := store.IsUrlOnDb(urlToCrawl)
	if crawler == nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal(err)
	} else if err != nil {
		log.Println("the page is not in our DB: ", err)
	} else if crawler != nil {
		if time.Since(crawler.LastTimeCrawled) > recrawlAfter {
			fmt.Print("needs recrawl")
			needsRecrawl = true
		} else {
			log.Println("Page already crawled successfuly", crawler.String())
		}
	}
	if crawler == nil || needsRecrawl {
		crawler := crawl.New(urlToCrawl, depthCrawl, time.Now())
		err := crawler.Crawl()
		if err != nil {
			log.Fatalf("Error crawling page: %s\n", err)
		}
		if depthCrawl > 1 {
			err = crawler.CrawlChildrenWithDepth()
			if err != nil {
				fmt.Printf("Error crawling page childs: %s\n", err)
			}
		}
		if needsRecrawl {
			fmt.Print("updating url with recrawl")
			err = store.UpdateLastCrawled(crawler.URL)
			if err != nil {
				log.Fatalf("Error trying to update the date on the url %s\n", err)
			}
			err = store.FilterOldChilds(crawler)
			if err != nil {
				log.Fatalf("Error trying to update the date on the url %s\n", err)
			}
		} else {
			err = store.EnterNewUrl(*crawler)
			if err != nil {
				log.Fatalf("Error inserting new url into db: %s\n", err)
			}
		}
		if crawler != nil {
			store.EnterNewChilds(*crawler)
		}
		log.Println("Page was crawled successfuly", crawler.String())
	}
}

func performSearch(store *db.Store, searchTerm string) {
	fmt.Println("Performing a search of urls in our database for the word: ", searchTerm)
	results, err := store.SearchTerm(searchTerm)
	if err != nil {
		log.Fatal(err)
	}
	if len(results) == 0 {
		log.Fatal("No results found in the db")
	}
	for _, v := range results {
		fmt.Println(v)
	}
}
