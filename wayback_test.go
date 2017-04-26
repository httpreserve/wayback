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

func TestHumanDate(t *testing.T) {

	var d1 = "http://web.archive.org/web/20170413225815/http://www.nationalarchives.gov.uk/"
	var d2 = "http://web.archive.org/web/20030415174607/http://www.nationalarchives.gov.uk:80/"
	var d3 = "http://web.archive.org/web/19961221203254/http://www0.bbc.co.uk:80/"
	var d4 = "http://web.archive.org/web/19961221203254/http://www0.bbc.co.uk:80/"

	var e1 = "13 April 2017"
	var e2 = "15 April 2003"
	var e3 = "21 December 1996"
	var e4 = "21 December 1996"

	r1 := GetHumanDate(d1)
	if r1 != e1 {
		t.Errorf("Unexpected response '%s', expected: '%s'", r1, e1)
	}

	r2 := GetHumanDate(d2)
	if r2 != e2 {
		t.Errorf("Unexpected response '%s', expected: '%s'", r2, e2)
	}

	r3 := GetHumanDate(d3)
	if r3 != e3 {
		t.Errorf("Unexpected response '%s', expected: '%s'", r3, e3)
	}

	r4 := GetHumanDate(d4)
	if r4 != e4 {
		t.Errorf("Unexpected response '%s', expected: '%s'", r4, e4)
	}

}

// Mock a response from the internet archive...
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

func TestSubmitToInternetArchive(t *testing.T) {

	// Not best practice to create a network connection during a unit test
	// maintain tests as a method of reverse engineering the wayback response
	// we can remove this once we've a better track record at handling the return

	var err error
	resp, err := SubmitToInternetArchive("http://www.bbc.co.uk/news", Version())
	if err != nil {
		t.Errorf("Unexpected response '%s', expected: '%s'", err.Error(), "nil")
	}

	resp, err = SubmitToInternetArchive("http://www.jezebel.com", Version())
	if err != nil {
		if err.Error() != SaveForbidden {
			t.Errorf("Unexpected response '%s', '%s', expected: '%s'", err.Error(), resp.StatusText, SaveForbidden)
		}
	}

	resp, err = SubmitToInternetArchive("http://xrssrssx.com", Version())
	if err != nil {
		if err.Error() != SaveGone {
			t.Errorf("Unexpected response '%s', '%s', expected: '%s'", err.Error(), resp.StatusText, SaveGone)
		}
	}
}
