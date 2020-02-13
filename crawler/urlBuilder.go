package crawler

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

// Match[1] will be the html attribute that has the link, Match[2] will be the url
var htmlURLReg = regexp.MustCompile("(href|src|action)=\"([a-zA-Z0-9:/\\-._~]+)\"")

// nextURL gets the next url based on the given absolute and relative urls
func nextURL(abs string, rel string) (*url.URL, error) {
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

func findAll(data []byte, parentURL string) []*url.URL {
	all := htmlURLReg.FindAllSubmatch(data, -1)
	ret := make([]*url.URL, 0, len(all))
	for _, match := range all {
		next, err := nextURL(parentURL, string(match[2]))
		if err == nil {
			ret = append(ret, next)
		}
	}
	return ret
}

func getURLsFromResponse(resp *http.Response) ([]*url.URL, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ret := findAll(data, resp.Request.URL.String())
	return ret, nil
}
