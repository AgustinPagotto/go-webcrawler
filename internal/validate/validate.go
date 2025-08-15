package validate

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateAndParseUrl(urlToValidate string) (*url.URL, error) {
	res, err := url.ParseRequestURI(urlToValidate)
	if err != nil {
		return nil, err
	}
	if res.Scheme != "https" {
		return nil, fmt.Errorf("The scheme of the url is not https, any other scheme is blocked")
	}
	res.Scheme = strings.ToLower(res.Scheme)
	res.Host = strings.ToLower(res.Host)
	return res, nil
}

func ValidateFlags(url string, depth int) error {
	if depth < 1 {
		return fmt.Errorf("The depth of crawl can't be less than 1")
	}
	if url == "" {
		return fmt.Errorf("Please provide a non-empty url")
	}
	return nil
}
