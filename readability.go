// Package readability is a Go package that find the main readable
// content from a HTML page. It works by removing clutter like buttons,
// ads, background images, script, etc.
//
// This package is based from Readability.js by Mozilla, and written line
// by line to make sure it looks and works as similar as possible. This
// way, hopefully all web page that can be parsed by Readability.js
// are parse-able by go-readability as well.
package readability

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// FromReader parses an `io.Reader` and returns the readable content. It's the wrapper
// or `Parser.Parse()` and useful if you only want to use the default parser.
func FromReader(input io.Reader, pageURL *nurl.URL) (Article, error) {
	parser := NewParser()
	return parser.Parse(input, pageURL)
}

// FromDocument parses an document and returns the readable content. It's the wrapper
// or `Parser.ParseDocument()` and useful if you only want to use the default parser.
func FromDocument(doc *html.Node, pageURL *nurl.URL) (Article, error) {
	parser := NewParser()
	return parser.ParseDocument(doc, pageURL)
}

// FromURL fetch the web page from specified url then parses the response to find
// the readable content.
func FromURL(pageURL string, timeout time.Duration) (Article, error) {
	// Make sure URL is valid
	parsedURL, err := nurl.ParseRequestURI(pageURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Fetch page from URL
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return Article{}, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Accept-Encoding header to indicate support for gzip
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return Article{}, fmt.Errorf("failed to fetch the page: %v", err)
	}
	defer resp.Body.Close()

	// Check if the content is encoded with gzip
	var reader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		// If encoded with gzip, use a gzip reader
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return Article{}, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.(*gzip.Reader).Close()
	default:
		// If not encoded, use the response body as is
		reader = resp.Body
	}

	// Make sure content type is HTML
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "text/html") {
		return Article{}, fmt.Errorf("URL is not a HTML document")
	}

	// Parse content
	parser := NewParser()
	return parser.Parse(reader, parsedURL)
}

// Check checks whether the input is readable without parsing the whole thing. It's the
// wrapper for `Parser.Check()` and useful if you only use the default parser.
func Check(input io.Reader) bool {
	parser := NewParser()
	return parser.Check(input)
}

// CheckDocument checks whether the document is readable without parsing the whole thing.
// It's the wrapper for `Parser.CheckDocument()` and useful if you only use the default
// parser.
func CheckDocument(doc *html.Node) bool {
	parser := NewParser()
	return parser.CheckDocument(doc)
}
