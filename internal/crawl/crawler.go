package crawl

import (
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

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

func CrawlPage(urlToParse string) (*models.PageData, error) {
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
	var linksDepthOne []string
	for _, v := range pageBeingCrawled.TextAndLinks {
		linksDepthOne = append(linksDepthOne, v)
	}
	concurrentCrawl(linksDepthOne)
	return pageBeingCrawled, nil
}

func concurrentCrawl(links []string) {
	var wg sync.WaitGroup
	pagesTitlesStream := make(chan map[string]string)
	concurrencyErrors := make(chan error)
	for i, link := range links {
		wg.Add(1)
		go func(i int, link string) {
			defer wg.Done()
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
				concurrencyErrors <- err
				return
			}
			defer resp.Body.Close()
			tokenizer := html.NewTokenizer(resp.Body)
			textAndLinks, err := retrieveUrlData(baseUrl, tokenizer)
			if err != nil {
				concurrencyErrors <- err
				return
			}
			pagesTitlesStream <- textAndLinks
		}(i, link)
	}
	go func() {
		wg.Wait()
		close(pagesTitlesStream)
		close(concurrencyErrors)
	}()
	var amountOfErrors, amountOfLinks int
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
					amountOfLinks++
				}
			}
		}
		if concurrencyErrors == nil && pagesTitlesStream == nil {
			fmt.Println("amount of errors: ", amountOfErrors, " amount of links: ", amountOfLinks)
			break
		}
	}
}
