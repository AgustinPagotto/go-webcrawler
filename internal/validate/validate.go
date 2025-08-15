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
