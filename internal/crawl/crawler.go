package crawl

import (
	"context"
	"errors"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
	"golang.org/x/net/html"
	"maps"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Crawler struct {
	URL              string
	Depth            int
	Status           int
	TextLinksCrawled map[string]string
	LastTimeCrawled  time.Time
}

type Result struct {
	Error       error
	InfoCrawled map[string]string
}

func New(url string, depth int, status int, timeOfCrawl time.Time) *Crawler {
	p := Crawler{URL: url, Status: status, TextLinksCrawled: make(map[string]string), Depth: depth, LastTimeCrawled: timeOfCrawl}
	return &p
}

func (c *Crawler) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s \t %d \t %d \t %s", c.URL, c.Status, len(c.TextLinksCrawled), c.LastTimeCrawled.String()))
	return b.String()
}

func (c *Crawler) Crawl() error {
	crawlResult, validUrl, statusCode := crawlLink(c.URL)
	if crawlResult.Error != nil {
		return crawlResult.Error
	}
	c.URL = validUrl.String()
	c.Status = statusCode
	c.TextLinksCrawled = crawlResult.InfoCrawled
	c.LastTimeCrawled = time.Now()
	fmt.Print("status of crawl", statusCode, c.Status)
	return nil
}

func (c *Crawler) CrawlChildrenWithDepth() error {
	for range c.Depth {
		var linksNextDepth []string
		for _, v := range c.TextLinksCrawled {
			linksNextDepth = append(linksNextDepth, v)
		}
		concurrentResult, _ := concurrentCrawl(linksNextDepth)
		maps.Copy(c.TextLinksCrawled, concurrentResult)
	}
	return nil
}

func concurrentCrawl(links []string) (map[string]string, error) {
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
	fmt.Printf("\nWe are going to create %d number of goroutines to make the crawl concurrent", numCrawlers)
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
			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: timeout exceeded: %v ", err)
		} else if errors.Is(err, context.Canceled) {
			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: canceled by user: %v", err)
		} else {
			return textAndLinksCrawled, fmt.Errorf("Crawl stopped: %v", err)
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
						trimmedUrl := strings.TrimSpace(value.Val)
						anchorUrl, err := url.Parse(trimmedUrl)
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
