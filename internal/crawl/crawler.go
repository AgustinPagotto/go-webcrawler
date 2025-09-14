package crawl

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
	"golang.org/x/net/html"
)

func CrawlPage(urlToParse string, depth int) (*models.PageData, error) {
	fmt.Println("Crawling the url: ", urlToParse, "with a depth of: ", depth)
	baseUrl, err := validate.ValidateAndParseUrl(urlToParse)
	if err != nil {
		return nil, fmt.Errorf("error validating the url: %s", err)
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	pageBeingCrawled := models.NewPageData(urlToParse, 0, time.Now())
	resp, err := client.Get(urlToParse)
	if err != nil {
		return nil, fmt.Errorf("there was an error trying to perform a get on the baseUrl %s", err)
	}
	defer resp.Body.Close()
	pageBeingCrawled.Status = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		return pageBeingCrawled, nil
	}
	tokenizer := html.NewTokenizer(resp.Body)
	if tokenizer == nil {
		return pageBeingCrawled, nil
	}
	textAndLinks, err := retrieveUrlData(baseUrl, tokenizer)
	pageBeingCrawled.TextAndLinks = textAndLinks
	for range depth {
		var linksNextDepth []string
		for _, v := range pageBeingCrawled.TextAndLinks {
			linksNextDepth = append(linksNextDepth, v)
		}
		concurrentResult := ConcurrentCrawlNew(linksNextDepth)
		maps.Copy(pageBeingCrawled.TextAndLinks, concurrentResult)
	}
	fmt.Println("amount of links retrieved in total", len(pageBeingCrawled.TextAndLinks))
	return pageBeingCrawled, nil
}

func retrieveUrlData(baseUrl *url.URL, z *html.Tokenizer) (map[string]string, error) {
	textAndLinks := make(map[string]string)
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
							fmt.Printf("skipping malformed href: %s\n", err)
							continue
						}
						link = baseUrl.ResolveReference(anchorUrl).String()
					}
				}
				nextToken := z.Next()
				if nextToken == html.TextToken && link != "" {
					trimmedText := strings.TrimSpace(string(z.Text()))
					if trimmedText != "" {
						textAndLinks[trimmedText] = link
					}
				}
			}
		}
	}
	return textAndLinks, nil
}

func ConcurrentCrawl(links []string) map[string]string {
	var wg sync.WaitGroup
	pagesTitlesStream := make(chan map[string]string)
	concurrencyErrors := make(chan error)
	crawlerWorker := func(link string, wgf *sync.WaitGroup, errChan chan<- error, titleChan chan<- map[string]string) {
		defer wgf.Done()
		baseUrl, err := validate.ValidateAndParseUrl(link)
		if err != nil {
			concurrencyErrors <- err
			return
		}
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Get(link)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()
		tokenizer := html.NewTokenizer(resp.Body)
		textAndLinks, err := retrieveUrlData(baseUrl, tokenizer)
		if err != nil {
			errChan <- err
			return
		}
		titleChan <- textAndLinks
	}
	for _, link := range links {
		wg.Add(1)
		go crawlerWorker(link, &wg, concurrencyErrors, pagesTitlesStream)
	}
	go func() {
		wg.Wait()
		close(pagesTitlesStream)
		close(concurrencyErrors)
	}()
	var amountOfErrors int
	textAndLinksCrawled := make(map[string]string)
	for {
		select {
		case _, ok := <-concurrencyErrors:
			if !ok {
				concurrencyErrors = nil
			} else {
				amountOfErrors++
			}
		case pages, ok := <-pagesTitlesStream:
			if !ok {
				pagesTitlesStream = nil
			} else {
				for range pages {
					maps.Copy(textAndLinksCrawled, pages)
				}
			}
		}
		if concurrencyErrors == nil && pagesTitlesStream == nil {
			return textAndLinksCrawled
		}
	}
}

type Result struct {
	Error       error
	InfoCrawled map[string]string
}

func crawlLink(link string) Result {
	baseUrl, err := validate.ValidateAndParseUrl(link)
	if err != nil {
		return Result{Error: err, InfoCrawled: nil}
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(link)
	if err != nil {
		return Result{Error: err, InfoCrawled: nil}
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	textAndLinks, err := retrieveUrlData(baseUrl, tokenizer)
	if err != nil {
		return Result{Error: err, InfoCrawled: nil}
	}
	return Result{Error: err, InfoCrawled: textAndLinks}
}

func ConcurrentCrawlNew(links []string) map[string]string {
	var wg sync.WaitGroup
	crawledInfo := make(chan Result)
	linkStream := make(chan string)
	done := make(chan any)
	crawlerWorker := func(done <-chan any, linkChannel <-chan string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case link, ok := <-linkChannel:
					if !ok {
						return
					}
					crawledInfo <- crawlLink(link)
				case <-done:
					return
				}
			}
		}()
	}
	for range 5 {
		crawlerWorker(done, linkStream)
	}
	go func() {
		defer close(linkStream)
		for _, link := range links {
			linkStream <- link
		}
	}()
	go func() {
		wg.Wait()
		close(crawledInfo)
	}()
	textAndLinksCrawled := make(map[string]string)
	for {
		workerInfo, ok := <-crawledInfo
		if !ok {
			break
		}
		maps.Copy(textAndLinksCrawled, workerInfo.InfoCrawled)
	}
	return textAndLinksCrawled
}
