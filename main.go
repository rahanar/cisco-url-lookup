package main

import (
	"fmt"
	"log"
	"net/http"
	gourl "net/url"
	"strings"

	"github.com/rahanar/cisco-url-lookup/url"
)

var localDB = make(map[string]url.URL)

func buildLocalDB() {
	urls := []string{"test.com", "badwebsite.com"}
	for _, u := range urls {
		urlObj := url.NewURL()
		urlObj.SetHostname(u)
		urlObj.SetMalicious(true)
		localDB[urlObj.Hostname] = *urlObj
	}
}

func main() {
	buildLocalDB()
	log.Fatal(http.ListenAndServe(":8000", http.HandlerFunc(wrapperMuxHandler)))
}

// This wrapper intercepts the incoming requests and checks RequestURI
// to decide whether to call Handler from the DefaultServeMux or a custom one.
// This is necessary because the default Handler does a series of sanitazations
// before invoking registered handlers.
// If a URI contains multiple forward slashes (//) or . and .., the clean up process
// would eleminate them and return a redirect response.
func wrapperMuxHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.URL.RequestURI(), "/urlinfo/1/") {
		// TODO: figure out a bette way to handle this
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	isURLMaliciousHandler(w, r)
}

func isURLMaliciousHandler(w http.ResponseWriter, r *http.Request) {
	hostname, err := extractHostname(r.URL)
	if err != nil {
		log.Printf("error parsing request URI: %s", r.URL.RequestURI())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var msg string
	isMalicious := isURLMalicious(hostname)
	if isMalicious {
		msg = fmt.Sprintf("The following url is malicious %s", hostname)
	} else {
		msg = fmt.Sprintf("The following url is not malicious %s", hostname)
	}
	fmt.Fprintln(w, msg)
}

func extractHostname(u *gourl.URL) (string, error) {
	var extractedURI *gourl.URL
	inputURL := u.RequestURI()[len("/urlinfo/1/"):]
	extractedURI, err := gourl.Parse(inputURL)
	if err != nil {
		return "", err
	}

	// if the Host field is empty, it means the passed in URL doesn't have scheme set
	// adding https:// as a scheme to get a complete URL structure
	if extractedURI.Hostname() == "" {
		extractedURI, err = gourl.ParseRequestURI("https://" + inputURL)
		if err != nil {
			return "", err
		}
	}

	return extractedURI.Hostname(), nil
}

func isURLMalicious(hostname string) bool {
	urlVal, ok := localDB[hostname]
	if ok && urlVal.IsMalicious() {
		return true
	}
	return false
}
