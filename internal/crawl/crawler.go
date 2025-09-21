package crawl

import (
	"context"
	"errors"
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
		concurrentResult, _ := ConcurrentCrawl(linksNextDepth)
		maps.Copy(pageCrawledInfo.TextAndLinks, concurrentResult)
	}
	fmt.Println("amount of links retrieved in total", len(pageCrawledInfo.TextAndLinks))
	return &pageCrawledInfo, nil
}

func ConcurrentCrawl(links []string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	linkChannelGenerator := func(ctx context.Context, receivedLinks ...string) <-chan string {
		linkStream := make(chan string)
		go func() {
			defer close(linkStream)
			for _, v := range receivedLinks {
				select {
				case <-ctx.Done():
					return
				case linkStream <- v:
				}
			}
		}()
		return linkStream
	}
	crawlerWorker := func(ctx context.Context, linkChannel <-chan string) <-chan Result {
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
				case <-ctx.Done():
					return
				}
			}
		}()
		return results
	}
	fanIn := func(ctx context.Context, resultStream ...<-chan Result) <-chan Result {
		var wg sync.WaitGroup
		multiplexedStream := make(chan Result)
		multiplex := func(c <-chan Result) {
			defer wg.Done()
			for i := range c {
				select {
				case <-ctx.Done():
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
	linkChannel := linkChannelGenerator(ctx, links...)
	numCrawlers := runtime.NumCPU()
	crawlersChan := make([]<-chan Result, numCrawlers)
	for i := range numCrawlers {
		crawlersChan[i] = crawlerWorker(ctx, linkChannel)
	}

	textAndLinksCrawled := make(map[string]string)
	for c := range fanIn(ctx, crawlersChan...) {
		maps.Copy(textAndLinksCrawled, c.InfoCrawled)
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: timeout exceeded", err)
		} else if errors.Is(err, context.Canceled) {

			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: canceled by user", err)
		} else {
			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: ", err)
		}
	}
	return textAndLinksCrawled, nil
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
