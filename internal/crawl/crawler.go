package crawl

import (
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/validate"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PageData struct {
	URL          string
	TextAndLinks map[string]string
}

func newPageData(url string) *PageData {
	p := PageData{URL: url, TextAndLinks: make(map[string]string)}
	return &p
}

func (p *PageData) String() string {
	if len(p.TextAndLinks) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range p.TextAndLinks {
		b.WriteString(fmt.Sprintf("%s \t \t %s\n", k, v))
	}
	return b.String()
}

func CrawlPage(urlToParse string) (*PageData, error) {
	baseUrl, err := validate.ValidateAndParseUrl(urlToParse)
	if err != nil {
		return nil, fmt.Errorf("error validating the url: %s", err)
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	pageBeingCrawled := newPageData(urlToParse)
	resp, err := client.Get(pageBeingCrawled.URL)
	if err != nil {
		return nil, fmt.Errorf("there was an error trying to perform a get on the baseUrl %s", err)
	}
	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
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
							fmt.Printf("skipping malformed href: %s", err)
							continue
						}
						link = baseUrl.ResolveReference(anchorUrl).String()
					}
				}
				nextToken := z.Next()
				if nextToken == html.TextToken && link != "" {
					trimmedText := strings.TrimSpace(string(z.Text()))
					if trimmedText != "" {
						pageBeingCrawled.TextAndLinks[trimmedText] = link
					}
				}
			}
		}
	}
	return pageBeingCrawled, nil
}
