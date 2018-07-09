package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

const gcsUrl = "http://gcsweb.k8s.io/"

var (
	textPatternFlag = flag.String("pattern", "watchdog: BUG: soft lockup - CPU#", "Text pattern to search for in files.  Not a regex")
	fileNameFlag    = flag.String("file_name", "serial-1.log", "File name in which to search for pattern")
	urlFlag         = flag.String("url", "/gcs/kubernetes-jenkins/logs/", "base url from which to begin recursive search")
)

func main() {
	flag.Parse()
	findFiles(*urlFlag, *fileNameFlag, *textPatternFlag)
}

// findFiles prints out the files found in sub-urls of url which have the provided
// fileName, and contain textPattern within the text of the file.
func findFiles(url, fileName, textPattern string) {
	doc, err := getUrlHtml(gcsUrl + url)
	if err != nil {
		fmt.Printf("error getting URL html: %s\n", err)
		return
	}
	for _, file := range getFilesFromHtml(doc, fileName) {
		text, err := getUrlText(file)
		if err != nil {
			fmt.Printf("error getting URL text: %s\n", err)
		} else if strings.Contains(text, textPattern) {
			fmt.Println(file)
		}
		return
	}
	for _, subDir := range getSubURLsFromHtml(doc, url) {
		findFiles(subDir, fileName, textPattern)
	}
	return
}

// getUrlHtml returns the html.Node associated with the body of the http response from the provided URL,
// or an error if one occurrs.
func getUrlHtml(url string) (*html.Node, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating http request for %s: %s", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing http request %+v: %s", req, err)
	}
	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing http response for req %+v: %s", req, err)
	}
	return doc, nil
}

// getUrlHtml returns the text read from the body of the http response from the provided URL,
// or an error if one occurrs.
func getUrlText(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating http request for %s: %s", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing http request %+v: %s", req, err)
	}
	defer res.Body.Close()
	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing text response for req %+v: %s", req, err)
	}
	return string(output), nil
}

// getSubURLsFromHtml returns a list of all links in the provided document which contain the provided url
// e.g. Searching a document for '/a' could return ['/a/b', '/a/c'] if both of those links exist in the document
func getSubURLsFromHtml(document *html.Node, url string) []string {
	return getElementsFromHtml(document, func(a html.Attribute) bool {
		return a.Key == "href" && strings.Contains(a.Val, url) && !strings.Contains(a.Val, "?")
	})
}

// getFilesFromHtml returns a list of all files matching fileName in the provided document
func getFilesFromHtml(document *html.Node, fileName string) []string {
	return getElementsFromHtml(document, func(a html.Attribute) bool {
		return a.Key == "href" && strings.HasSuffix(a.Val, fileName)
	})
}

// htmlFilter is a function which returns true if the value of the html attribute should be included
type htmlFilter func(a html.Attribute) bool

// getElementsFromHtml recursively searches the DOM to find attributes for which filter returns true.
// It returns a list of the values of all matching attributes.
func getElementsFromHtml(n *html.Node, filter htmlFilter) []string {
	output := []string{}
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if filter(a) {
				output = append(output, a.Val)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		output = append(output, getElementsFromHtml(c, filter)...)
	}
	return output
}
