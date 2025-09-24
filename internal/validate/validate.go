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

func ValidateFlags(url string, depth int, search string) (bool, error) {
	if url == "" && search != "" {
		if len(search) < 3 {
			return true, fmt.Errorf("word too small for a search")
		}
		return true, nil
	}
	if url != "" && search == "" {
		if depth < 1 {
			return false, fmt.Errorf("The depth of crawl can't be less than 1")
		}
		return false, nil
	}
	return false, fmt.Errorf("No requirement were satisfied, won't perform anything")
}
