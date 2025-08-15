package validate

import "testing"

func TestValidateUrl(t *testing.T) {
	ValidateUrl("https://www.google.com")
}
