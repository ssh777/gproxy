package main

import (
	"net/http"
	"strings"
)

func copyHeaders(w http.ResponseWriter, r *http.Response) {
	wh := w.Header()
	for key, value := range r.Header {
		wh[key] = value
	}
}

func isDecompressionRequired(request *http.Request, response *http.Response) bool {
	requestCompressed := false
	responceCompressed := false
	/**
	* check request header
	*/
	if h := getHeaderValue(request.Header, "accept-encoding"); h != nil {
		for _, h := range h {
			h = strings.ToLower(h)
			if strings.Contains(h, "gzip") {
				requestCompressed = true
				break
			}
		}
	}
	/**
	* check response header
	 */
	if h := getHeaderValue(response.Header, "content-encoding"); h != nil {
		for _, h := range h {
			h = strings.ToLower(h)
			if strings.Contains(h, "gzip") {
				responceCompressed = true
				break
			}
		}
	}
	if !responceCompressed {
		if h := getHeaderValue(response.Header, "transfer-encoding"); h != nil {
			for _, h := range h {
				h = strings.ToLower(h)
				if strings.Contains(h, "gzip") {
					responceCompressed = true
					break
				}
			}
		}
	}
	return requestCompressed == false && responceCompressed == true
}

func getHeaderValue(h http.Header, key string) []string {
	for k, v := range h {
		if key == strings.ToLower(k) {
			return v
		}
	}
	return nil
}
