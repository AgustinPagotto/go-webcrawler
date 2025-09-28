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
	dbConn, err := db.OpenConToDB()
	if err != nil {
		log.Fatalf("Error openning connection to the db: %s\n", err)
	}
	err = db.InitiateDB(dbConn)
	if err != nil {
		log.Fatalf("Error adding tables to db: %s\n", err)
	}
	defer dbConn.Close()
	if searchBool {
		performSearch(dbConn, searchTerm)
	} else {
		performCrawl(dbConn, urlToCrawl, depthCrawl)
	}
}

func performCrawl(dbConn *sql.DB, urlToCrawl string, depthCrawl int) {
	var needsRecrawl bool
	const recrawlAfter = 1 * 24 * time.Hour
	res, err := db.IsUrlOnDb(dbConn, urlToCrawl)
	if res == nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal(err)
	} else if err != nil {
		log.Println("the page is not in our DB: ", err)
	} else if res != nil {
		if time.Since(res.LastCrawled) > recrawlAfter {
			fmt.Print("needs recrawl")
			needsRecrawl = true
		}
	}
	if res == nil || needsRecrawl {
		crawler := crawl.New(urlToCrawl, depthCrawl)
		err := crawler.Crawl()
		if err != nil {
			log.Fatalf("Error crawling page: %s\n", err)
		}
		err = crawler.CrawlChildrenWithDepth()
		if err != nil {
			log.Fatalf("Error crawling page childs: %s\n", err)
		}
		if needsRecrawl {
			fmt.Print("updating url with recrawl")
			err = db.UpdateLastCrawled(dbConn, res.URL)
			if err != nil {
				log.Fatalf("Error trying to update the date on the url %s\n", err)
			}
			err = db.FilterOldChilds(dbConn, res)
			if err != nil {
				log.Fatalf("Error trying to update the date on the url %s\n", err)
			}
		} else {
			err = db.EnterNewUrl(dbConn, res)
			if err != nil {
				log.Fatalf("Error inserting new url into db: %s\n", err)
			}
		}
		if res.TextAndLinks != nil {
			db.EnterNewChilds(dbConn, crawler)
		}
	}
	log.Println("Page was crawled successfuly", res.String())
}

func performSearch(dbConn *sql.DB, searchTerm string) {
	fmt.Println("Performing a search of urls in our database for the word: ", searchTerm)
	results, err := db.SearchTerm(dbConn, searchTerm)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range results {
		fmt.Println(v)
	}
}
