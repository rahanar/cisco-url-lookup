package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	gourl "net/url"
	"os"
	"strings"

	"github.com/rahanar/cisco-url-lookup/url"
)

var handlerPattern = "/urlinfo/1/"

var localDB = make(map[string]url.URL)

type URLsJsonFile struct {
	Hostnames []url.URL `json:"hostnames"`
}

func buildLocalDB(urls URLsJsonFile) {
	for _, u := range urls.Hostnames {
		urlObj := url.NewURL()
		urlObj.SetHostname(u.Hostname)
		urlObj.SetMalicious(u.Malicious)
		localDB[urlObj.Hostname] = *urlObj
	}
}

func main() {
	setupLocalDB()
	log.Fatal(http.ListenAndServe(":8000", http.HandlerFunc(wrapperMuxHandler)))
}

func setupLocalDB() {
	jsfile, err := os.Open("./url-database.json")
	if err != nil {
		panic(err)
	}
	defer jsfile.Close()

	byteJSFile, err := ioutil.ReadAll(jsfile)
	if err != nil {
		panic(err)
	}

	var urlsJSONFile URLsJsonFile
	json.Unmarshal(byteJSFile, &urlsJSONFile)
	buildLocalDB(urlsJSONFile)
}

// This wrapper intercepts the incoming requests and checks RequestURI
// to decide whether to call Handler from the DefaultServeMux or a custom one.
// This is necessary because the default Handler does a series of sanitazations
// before invoking registered handlers.
// If a URI contains multiple forward slashes (//) or . and .., the clean up process
// would eleminate them and return a redirect response.
func wrapperMuxHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.URL.RequestURI(), handlerPattern) || (strings.HasPrefix(r.URL.RequestURI(),
		handlerPattern) && r.Method != http.MethodGet) {
		// Everything that's not /urlinfo/1/ will be ignored.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Resource not found","status": "NotFound"}`))
		return
	}

	isURLMaliciousHandler(w, r)
}

func isURLMaliciousHandler(w http.ResponseWriter, r *http.Request) {
	requestURL := r.URL.RequestURI()[len(handlerPattern):]
	hostname, err := extractHostname(requestURL)
	if err != nil {
		log.Printf("error parsing request URI: %s", requestURL)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var msgByte []byte
	isMalicious := isURLMalicious(hostname)

	// Have to set Content-Type before WriteHeader
	w.Header().Set("Content-Type", "application/json")
	if isMalicious {
		w.WriteHeader(http.StatusForbidden)
		msg := fmt.Sprintf(`{"message":"URL is malicious: %s","status":"Forbidden"}`, requestURL)
		msgByte = []byte(msg)
	} else {
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf(`{"message":"URL is not malicious: %s","status":"OK"}`, requestURL)
		msgByte = []byte(msg)
	}
	w.Write(msgByte)
}

func extractHostname(incomingURL string) (string, error) {
	var extractedURI *gourl.URL
	extractedURI, err := gourl.Parse(incomingURL)
	if err != nil {
		return "", err
	}

	// if the Host field is empty, it means the passed in URL doesn't have scheme set
	// adding https:// as a scheme to get a complete URL structure
	if extractedURI.Hostname() == "" {
		extractedURI, err = gourl.ParseRequestURI("https://" + incomingURL)
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
