package urlDecode

import (
	"net/url"
)
//
func parse(url string) (string, err) {
	// decode URL by url.Parse
	return url.Parse("https://example.com/foo+bar%21?query=ab%2Bc&query2=de%24f")
}
