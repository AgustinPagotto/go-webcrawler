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
	var needsRecrawl bool
	const recrawlAfter = 1 * 24 * time.Hour
	urlToCrawl, depthCrawl := receiveFlags()
	if err := validate.ValidateFlags(urlToCrawl, depthCrawl); err != nil {
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
		res, err = crawl.CrawlPage(urlToCrawl)
		if err != nil {
			log.Fatalf("Error crawling page: %s\n", err)
		}
		if needsRecrawl {
			fmt.Print("updating recrawl")
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
			db.EnterNewChilds(dbConn, res)
		}
	}
	log.Println("Page was crawled successfuly", res.String())
}
