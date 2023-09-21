package urlDecode

import (
	"fmt"
	"log"
)
//
func UrlDecode() {
	// decode URL by url.Parse
	parsedURL, err := parse("https://example.com/foo+bar%21?query=ab%2Bc&query2=de%24f")
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("scheme: %s\n", parsedURL.Scheme)
	fmt.Printf("host: %s\n", parsedURL.Host)
	fmt.Printf("path: %s\n", parsedURL.Path)
	fmt.Println("query args:")
	for key, values := range parsedURL.Query() {
		fmt.Printf("  %s = %s\n", key, values[0])
	}
}
