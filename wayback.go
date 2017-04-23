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
const iaRel = "rel="

// WAYEARLY helps us to retrieve the earliest link from memento
const WAYEARLY = "earliest"

// WAYLATE helps us to retrieve the latest link from memento
const WAYLATE = "latest"

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
	ResponseCode    int    // Reponse code from the Internet Archive
	ResponseText    string // Human readable response text from the Internet Archive
}

// ErrorNoIALink enables us to check for the non-existence of a record
var ErrorNoIALink = errors.New("no internet archive record")

// ErrorIAExists so that we can identify links we do not need to process
// a second time, or send to IA
var ErrorIAExists = errors.New("already an internet archive record")

// GetWaybackData returns some wayback information for the calling code in an
// appropriate struct... groups external functions conveniently
func GetWaybackData(link string) (Data, error) {

	var wb Data

	if !IsWayback(link) {

		earliest, err := GetPotentialURLEarliest(link)
		if err != nil {
			return wb, errors.Wrap(err, "IA url creation failed")
		}

		// We don't NotWaybackhave to be concerned with error here is URL is already
		// previously Parsed correctly, which we do so dilligently under iafunctions.go
		sr, err := simplerequest.Create(simplerequest.HEAD, earliest.String())

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
			if resp.Header.Get("Link") == "" {
				wb.NotInWayback = true
				return wb, nil
			}
		}

		// Else, continue to retrieve IA links
		iaLinkData := resp.Header.Get("Link")
		iaLinkInfo := strings.Split(iaLinkData, ", <")

		var mementomap = make(map[string]string)

		for _, lnk := range iaLinkInfo {
			trimmedlink := strings.Trim(lnk, " ")
			trimmedlink = strings.Replace(trimmedlink, ">;", ";", 1) // fix chevrons
			for _, rel := range iaRelList {
				if strings.Contains(trimmedlink, rel) {
					mementomap[rel] = trimmedlink
					break
				}
			}
		}

		// We've some internet archive links that we can use
		if len(mementomap) > 0 {
			links := GetWaybackLinkrange(mementomap)
			wb.EarliestWayback = links[WAYEARLY]
			wb.LatestWayback = links[WAYLATE]
		}

	} else {
		wb.AlreadyWayback = ErrorIAExists
	}

	wb.WaybackSaveURL = SaveURL(link)

	return wb, nil
}

//Explanation: https://andrey.nering.com.br/2015/how-to-format-date-and-time-with-go-lang/
//Golang Date Formatter: http://fuckinggodateformat.com/
const datelayout = "20060102150405"

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

// Memento returns various relationship attributes
// These are the ones observed so far in this work.
// Rather than separating the attributes, use whole string.
const relFirst = "rel=\"first memento\""          // syn: first, at least two
const relNext = "rel=\"next memento\""            // syn: at least three
const relLast = "rel=\"last memento\""            // syn: last, at least three
const relFirstLast = "rel=\"first last memento\"" // syn: only
const relNextLast = "rel=\"next last memento\""   // syn: second, and last
const relPrevLast = "rel=\"prev memento\""        // syn: at least three
const relPrevFirst = "rel=\"prev first memento\"" // syn: previous, and first, only two

// List of items to check against when parsing header attributes
var iaRelList = [...]string{relFirst, relNext, relLast, relFirstLast,
	relNextLast, relPrevLast, relPrevFirst}

// GetWaybackLinkrange will return the earliest and latest wayback links based on
// an understanding of the headers returned by the server.
func GetWaybackLinkrange(legacyCollection map[string]string) map[string]string {
	var links = make(map[string]string)
	for rel, lnk := range legacyCollection {
		switch rel {
		// first two cases give us the earliest IA link available
		case relFirst:
			fallthrough
		case relFirstLast:
			links[WAYEARLY] = getWaybackfromRel(lnk)
		// second two cases give us the latest IA link available
		case relLast:
			fallthrough
		case relNextLast:
			links[WAYLATE] = getWaybackfromRel(lnk)
		}
	}
	return links
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
