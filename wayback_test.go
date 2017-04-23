package wayback

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIALinkNowDate(t *testing.T) {
	archiveurl, _ := GetPotentialURLLatest("http://www.bbc.co.uk/news")
	fmt.Println(archiveurl)
}

func TestIALinkEarliestDate(t *testing.T) {
	archiveurl, _ := GetPotentialURLEarliest("http://www.bbc.co.uk/news")
	fmt.Println(archiveurl)
}

func TestSavedURL(t *testing.T) {
	u, _ := GetSavedURL(generateInternetArchiveSaveMock())
	fmt.Println(u)
}

func generateInternetArchiveSaveMock() http.Response {

	var r = http.Response{}

	r.Status = "200 OK"
	r.StatusCode = 200
	r.Proto = "HTTP/1.0" //probably not needed

	var h = http.Header{}
	h.Add("Content-Location", "/web/20170314100523/http://www.bbc.co.uk/news")
	h.Add("X-Archive-Orig-Vary", "X-CDN,X-BBC-Edge-Cache,Accept-Encoding")
	h.Add("Content-Type", "text/html;charset=utf-8")
	h.Add("X-Archive-Orig-X-News-Data-Centre", "cwwtf")
	h.Add("X-Page-Cache", "MISS")
	h.Add("X-Archive-Orig-X-Pal-Host", "pal029.back.live.cwwtf.local:80")
	h.Add("Server", "Tengine/2.1.0")

	r.Header = h

	return r
}
