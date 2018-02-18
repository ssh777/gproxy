package main

import (
	"net/http"
	"strings"
	"log"
	"compress/gzip"
	"io"
)

type HttpRoute struct {
	path 		string
	location 	LocationConf
	proxy 		*GProxy
}

func (route *HttpRoute) httpHandler(w http.ResponseWriter, r *http.Request) {
	requestUrl := route.getDestinationUrl(r)
	if requestUrl == "" {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}
	request, err := http.NewRequest(r.Method, requestUrl, r.Body)
	if err != nil {
		log.Println(err)
	}
	request.Header = r.Header
	resp, err := route.proxy.client.Do(request);
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	respReader := resp.Body
	if route.isGziped(resp) {
		respReader, _ = gzip.NewReader(resp.Body)
	}
	_, err = io.Copy(w, respReader)
	if err != nil {
		log.Println(err)
	}
}

func (route *HttpRoute) getDestinationUrl(r *http.Request) string {
	url := route.location.Destination
	if path := r.URL.Path[len(route.path):]; path != "" {
		url += "/" + path
	}
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery;
	}
	return url;
}

func (route *HttpRoute) isGziped(r *http.Response) bool {
	header, ok := r.Header["Content-Encoding"]
	if !ok {
		header, ok = r.Header["Transfer-Encoding"]
		if !ok {
			return false
		}
	}
	for _, h := range header {
		if strings.Contains(h, "gzip") {
			return true
		}
	}
	return false
}
