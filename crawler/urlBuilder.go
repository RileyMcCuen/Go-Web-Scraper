package crawler

import (
	"fmt"
	"net/url"
)

// nextURL gets the next url based on the given absolute and relative urls
func nextURL(abs string, rel string) (*url.URL, error) {
	fmt.Printf("%s %s\n", abs, rel)
	absURL, err := url.Parse(abs)
	if !(err == nil && absURL.IsAbs()) {
		return nil, err
	}
	relURL, err := url.Parse(rel)
	if err != nil {
		return nil, err
	}
	conc := absURL.ResolveReference(relURL)
	if conc.IsAbs() {
		return conc, nil
	}
	return nil, url.InvalidHostError("not absolute")
}
