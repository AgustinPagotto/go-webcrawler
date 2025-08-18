package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	"github.com/AgustinPagotto/go-webcrawler/internal/db"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
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

func main() {
	urlToCrawl, depthCrawl := receiveFlags()
	if err := validate.ValidateFlags(urlToCrawl, depthCrawl); err != nil {
		log.Fatal(err)
	}
	DB, err := db.OpenConToDB()
	if err != nil {
		log.Fatalf("Error openning connection to the db: %s\n", err)
	}
	defer DB.Close()
	res, err := db.IsUrlOnDb(DB, urlToCrawl)
	if res == nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal(err)
	} else if err != nil {
		fmt.Println("the page is not in our DB: ", err)
	}
	if res == nil {
		res, err = crawl.CrawlPage(urlToCrawl)
		if err != nil {
			log.Fatalf("Error crawling page: %s\n", err)
		}
		err = db.InitiateDB(DB)
		if err != nil {
			log.Fatalf("Error adding tables to db: %s\n", err)
		}
		err = db.EnterNewUrl(DB, res)
		if err != nil {
			log.Fatalf("Error inserting new url into db: %s\n", err)
		}
		if res.TextAndLinks != nil {
			db.EnterNewChilds(DB, res)
		}
	}
	fmt.Println("page was crawled successfuly", res.String())
}
