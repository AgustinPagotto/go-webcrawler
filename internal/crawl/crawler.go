package crawl

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
	"golang.org/x/net/html"
)

type Result struct {
	Error       error
	InfoCrawled map[string]string
}

func CrawlPage(urlToCrawl string, depth int) (*models.PageData, error) {
	fmt.Println("Crawling the url: ", urlToCrawl, "with a depth of: ", depth)
	var pageCrawledInfo models.PageData
	crawlResult, validUrl, statusCode := crawlLink(urlToCrawl)
	if crawlResult.Error != nil {
		return nil, crawlResult.Error
	}
	pageCrawledInfo.URL = validUrl.String()
	pageCrawledInfo.Status = statusCode
	pageCrawledInfo.TextAndLinks = crawlResult.InfoCrawled
	for range depth {
		var linksNextDepth []string
		for _, v := range pageCrawledInfo.TextAndLinks {
			linksNextDepth = append(linksNextDepth, v)
		}
		concurrentResult := ConcurrentCrawl(linksNextDepth)
		maps.Copy(pageCrawledInfo.TextAndLinks, concurrentResult)
	}
	fmt.Println("amount of links retrieved in total", len(pageCrawledInfo.TextAndLinks))
	return &pageCrawledInfo, nil
}

func ConcurrentCrawl(links []string) map[string]string {
	done := make(chan any)
	defer close(done)
	linkChannelGenerator := func(done <-chan any, receivedLinks ...string) <-chan string {
		linkStream := make(chan string)
		go func() {
			defer close(linkStream)
			for _, v := range receivedLinks {
				select {
				case <-done:
					return
				case linkStream <- v:
				}
			}
		}()
		return linkStream
	}
	crawlerWorker := func(done <-chan any, linkChannel <-chan string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			defer fmt.Println("goroutine closed")
			for {
				select {
				case link, ok := <-linkChannel:
					if !ok {
						return
					}
					crawlResult, _, _ := crawlLink(link)
					results <- crawlResult
				case <-done:
					return
				}
			}
		}()
		return results
	}
	fanIn := func(done <-chan any, resultStream ...<-chan Result) <-chan Result {
		var wg sync.WaitGroup
		multiplexedStream := make(chan Result)
		multiplex := func(c <-chan Result) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}
		wg.Add(len(resultStream))
		for _, c := range resultStream {
			go multiplex(c)
		}
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()
		return multiplexedStream
	}
	linkChannel := linkChannelGenerator(done, links...)
	numCrawlers := runtime.NumCPU()
	crawlersChan := make([]<-chan Result, numCrawlers)
	for i := range numCrawlers {
		crawlersChan[i] = crawlerWorker(done, linkChannel)
	}

	textAndLinksCrawled := make(map[string]string)
	for c := range fanIn(done, crawlersChan...) {
		maps.Copy(textAndLinksCrawled, c.InfoCrawled)
	}
	return textAndLinksCrawled
}

func retrieveUrlData(baseUrl *url.URL, tz *html.Tokenizer) (map[string]string, error) {
	textAndLinks := make(map[string]string)
	for {
		tt := tz.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.StartTagToken {
			t := tz.Token()
			if t.Data == "a" {
				var link string
				tokenAttributes := t.Attr
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
				nextToken := tz.Next()
				if nextToken == html.TextToken && link != "" {
					trimmedText := strings.TrimSpace(string(tz.Text()))
					if trimmedText != "" {
						textAndLinks[trimmedText] = link
					}
				}
			}
		}
	}
	return textAndLinks, nil
}

func crawlLink(link string) (Result, *url.URL, int) {
	validatedUrl, err := validate.ValidateAndParseUrl(link)
	if err != nil {
		return Result{Error: err, InfoCrawled: nil}, nil, 0
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(validatedUrl.String())
	if err != nil {
		return Result{Error: fmt.Errorf("error trying to perform get to the url, %v", err), InfoCrawled: nil}, nil, resp.StatusCode
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	textAndLinks, err := retrieveUrlData(validatedUrl, tokenizer)
	if err != nil {
		return Result{Error: err, InfoCrawled: nil}, nil, resp.StatusCode
	}
	return Result{Error: err, InfoCrawled: textAndLinks}, validatedUrl, resp.StatusCode
}

func ConcurrentCrawlAlt(links []string) map[string]string {
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
func ConcurrentCrawlAlt1(links []string) map[string]string {
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
					crawlResult, _, _ := crawlLink(link)
					crawledInfo <- crawlResult
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
