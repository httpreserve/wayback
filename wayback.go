package wayback

import (
	"github.com/httpreserve/simplerequest"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const iaRoot = "http://web.archive.org"
const iaBeta = "http://web-beta.archive.org"

const iaSRoot = "https://web.archive.org"
const iaSBeta = "https://web-beta.archive.org"

const iaSave = "/save/" //e.g. https://web.archive.org/save/http://www.bbc.com/news
const iaWeb = "/web/"   //e.g. http://web.archive.org/web/20161104020243/http://exponentialdecayxxxx.co.uk/#

// IsWayback checks a URL (string) and returns whether or not we expect it
// to be an internet archive link or not...
func IsWayback(link string) bool {
	if strings.Contains(link, iaRoot) || strings.Contains(link, iaBeta) ||
		strings.Contains(link, iaSRoot) || strings.Contains(link, iaSBeta) {
		return true
	}
	return false
}

// Data stores information we're going to need for analysing what's
// in the internet archive. We need to follow a heuristic to use it most
// effectively. E.g. Use AlreadyWayback first to see what data we might
// have, NotWayback second to see if anything is already there...
type Data struct {
	AlreadyWayback  error  // Flags the URL as a Wayback URL already
	NotInWayback    bool   // Flags the URL as having zero entries in Wayback
	EarliestWayback string // String denoting the Earliest Wayback URL
	LatestWayback   string // String denoting the Latest available Wayback URL
	WaybackSaveURL  string // String to handle saving of the link in Wayback
	ResponseCode    int    // Response code from the Internet Archive
	ResponseText    string // Human readable response text from the Internet Archive
}

// ErrorNoIALink enables us to check for the non-existence of a record
var ErrorNoIALink = errors.New("no internet archive record")

// ErrorIAExists so that we can identify links we do not need to process
// a second time, or send to IA
var ErrorIAExists = errors.New("already an internet archive record")

// GetWaybackData returns some wayback information for the calling code in an
// appropriate struct... groups external functions conveniently, when calling
// externally, users can set their own agent string as required...
func GetWaybackData(link string, agent string) (Data, error) {

	var wb Data

	if !IsWayback(link) {

		earliest, err := GetPotentialURLEarliest(link)
		if err != nil {
			return wb, errors.Wrap(err, "IA url creation failed")
		}

		// We don't NotWaybackhave to be concerned with error here is URL is already
		// previously Parsed correctly, which we do so dilligently under iafunctions.go
		sr, err := simplerequest.Create(simplerequest.HEAD, earliest.String())

		sr.Accept("*/*")

		// Custom user agent...
		if agent == "" {
			sr.Agent(Version())
		} else {
			sr.Agent(agent)
		}

		sr.NoRedirect(true)

		//set some values for the simplerequest...
		sr.Timeout(10)

		resp, err := sr.Do()
		if err != nil {
			return wb, errors.Wrap(err, "IA http request failed")
		}

		wb.ResponseCode = resp.StatusCode
		wb.ResponseText = resp.StatusText

		// First test for existence of an internet archive copy
		if wb.ResponseCode == http.StatusNotFound {
			if resp.Header.Get("Location") == "" {
				wb.NotInWayback = true
				return wb, nil
			}
		}

		// Else, continue to retrieve IA links
		// Try and get the latest link available in the archive...
		wb.EarliestWayback = resp.Header.Get("Location")

		// Reuse our previous SimpleRequest struct to redo the work... 
		sr.URL, _ = GetPotentialURLLatest(link)
		resp, err = sr.Do()
		if err != nil {
			return wb, errors.Wrap(err, "IA http request failed")
		}

		// Add to our wayback structure...
		wb.LatestWayback = resp.Header.Get("Location")

	} else {
		wb.AlreadyWayback = ErrorIAExists
	}

	wb.WaybackSaveURL = SaveURL(link)

	return wb, nil
}

//Explanation: https://andrey.nering.com.br/2015/how-to-format-date-and-time-with-go-lang/
//Golang Date Formatter: http://fuckinggodateformat.com/
const datelayout = "20060102150405"
const humandate = "02 January 2006"

// GetPotentialURLLatest is used to create a URL that we can test for a 404
// error or 200 OK. The URL if it works can be used to display to
// the user for QA. The URL if it fails, can be used to prompt the
// user to save the URL as it is found today. A motivation, even if
// there is no saved IA record, to save copy today, even if it is a 404
// is that the earliest date we can pin on a broken link the
// better we can satisfy outselves in future that we did all we can.
// Example URI we need to create looks like this:
// web.archive.org/web/{date}/url-to-lookup
// {date} == "20161104020243" == "YYYYMMDDHHMMSS" == %Y%m%d%k%M%S
func GetPotentialURLLatest(archiveurl string) (*url.URL, error) {
	latestDate := time.Now().Format(datelayout)
	return constructURL(latestDate, archiveurl)
}

// GetPotentialURLEarliest is used to returning the
// earliest possible record available in the internet archive. We
// can make it easier by using this function here.
// Example URI we need to create looks like this:
// web.archive.org/web/{date}/url-to-lookup
func GetPotentialURLEarliest(archiveurl string) (*url.URL, error) {
	oldestDate := time.Date(1900, time.August, 31, 23, 13, 0, 0, time.Local).Format(datelayout)
	return constructURL(oldestDate, archiveurl)
}

const split1 = iaRoot + "/web/"
const split2 = iaBeta + "/web/"
const split3 = iaSRoot + "/web/"
const split4 = iaSBeta + "/web/"

var iasplits = []string{split1, split2, split3, split4}

// GetHumanDate returns a human readable date from an Internet Archive link
// rudimentary code for now. Can improve once we've got other pieces working.
func GetHumanDate(link string) string {
	var dateslug string
	for i := range iasplits {
		if strings.Contains(link, iasplits[i]) {
			r := strings.Split(link, iasplits[i])
			if len(r) == 2 {
				s := strings.Split(r[1], "/")
				dateslug = s[0]
			}
		}
	}

	if dateslug != "" {
		//latestDate := time.Now().Format(datelayout)
		date, err := time.Parse(datelayout, dateslug)
		if err != nil {
			return ""
		}
		return date.Format(humandate)
	}
	return ""
}

// Construct the url to return to either the IA earliest or latest
// IA get functions and return...
func constructURL(iadate string, archiveurl string) (*url.URL, error) {
	newurl, err := url.Parse(iaRoot + iaWeb + iadate + "/" + archiveurl)
	if err != nil {
		return newurl, errors.Wrap(err, "internet archive url creation failed")
	}
	return newurl, nil
}

// SaveURL is used to create a URL that will enable us to
// submit it to the Internet Archive SaveNow function
func SaveURL(link string) string {
	//e.g. https://web.archive.org/save/http://www.bbc.com/news
	return iaRoot + iaSave + link
}

// SubmitToInternetArchive will handle the request and response to
// and from the Internet Archive for a URL that we wish to save as
// part of this initiative.
func SubmitToInternetArchive() {

}

// GetSavedURL will help us to retrieve the URL returned by the
// Internet Archive when we've sent a request to the SaveNow function.
// We've constructed the URL to save ours in the Internet Archive
// We've submitted the URL via the IA REST API and we've receieved
// a 200 OK. In the response will be a partial SLUG that takes us
// to our newly archived record.
func GetSavedURL(resp http.Response) (*url.URL, error) {
	loc := resp.Header["Content-Location"]
	u, err := url.Parse(iaRoot + strings.Join(loc, ""))
	if err != nil {
		return &url.URL{}, errors.Wrap(err, "creation of URL from http response failed.")
	}
	return u, nil
}

// Retrieve the IA www link that we've been passing about
// from the IA response header sent to us previously.
func getWaybackfromRel(lnk string) string {
	lnksplit := strings.Split(lnk, "; ")
	for _, www := range lnksplit {
		if strings.Contains(www, iaRoot) {
			return www
		}
	}
	return ""
}

var version = "httpreserve-wayback-0.0.1"

// Version retrieves the version text for the httpreserve/wayback agent
func Version() string {
	return version
}
