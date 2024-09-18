package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type CrawlData struct {
	Url          string
	Success      bool
	ResponseCode int
	CrawlData    ParsedBody
}

type ParsedBody struct {
	CrawlTime       time.Duration
	PageTitle       string
	PageDescription string
	Headings        string
	Links           Links
}

type Links struct {
	Internal []string
	External []string
}

func runCrawl(inputUrl string) CrawlData {
	resp, _ := http.Get(inputUrl)
	baseUrl, err := url.Parse(inputUrl)

	if err != nil {
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: 0, CrawlData: ParsedBody{}}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Failed to crawl", inputUrl)
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
	}

	contentType := resp.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "text/html") {
		fmt.Println("Crawling", inputUrl)
		body, err := parseBody(resp.Body, baseUrl)

		if err != nil {
			fmt.Println("Error during parse", err)
			return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
		}

		return CrawlData{Url: inputUrl, Success: true, ResponseCode: resp.StatusCode, CrawlData: body}

	} else {
		fmt.Println("Not HTML, skipping", inputUrl)
		return CrawlData{Url: inputUrl, Success: false, ResponseCode: resp.StatusCode, CrawlData: ParsedBody{}}
	}
}

func parseBody(body io.Reader, baseUrl *url.URL) (ParsedBody, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return ParsedBody{}, err
	}

	start := time.Now()

	links := getLinks(doc, baseUrl)

	title, desc := getPageData(doc)

	headings := getPageHeadings(doc)

	end := time.Now()

	return ParsedBody{
		CrawlTime:       end.Sub(start),
		PageTitle:       title,
		PageDescription: desc,
		Headings:        headings,
		Links:           links,
	}, nil
}

func getLinks(node *html.Node, baseUrl *url.URL) Links {
	if node == nil {
		return Links{}
	}

	var links Links

	var findLinks func(*html.Node)
	findLinks = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					url, err := url.Parse(attr.Val)
					if err != nil ||
						strings.HasPrefix(url.String(), "#") ||
						strings.HasPrefix(url.String(), "tel") ||
						strings.HasPrefix(url.String(), "mail") ||
						strings.HasPrefix(url.String(), "javascript") ||
						strings.HasPrefix(url.String(), ".pdf") ||
						strings.HasPrefix(url.String(), ".md") {
						continue
					}

					if url.IsAbs() {
						if isSameHost(url.String(), baseUrl.String()) {
							links.Internal = append(links.Internal, url.String())
						} else {
							links.External = append(links.External, url.String())
						}

					} else {
						rel := baseUrl.ResolveReference(url)
						links.Internal = append(links.Internal, rel.String())
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findLinks(child)
		}
	}

	findLinks(node)

	return links
}

func isSameHost(absoluteUrl string, baseUrl string) bool {
	absUrl, err := url.Parse(absoluteUrl)
	if err != nil {
		return false
	}

	base, err := url.Parse(baseUrl)
	if err != nil {
		return false
	}

	return absUrl.Host == base.Host
}

func getPageData(node *html.Node) (string, string) {
	if node == nil {
		return "", ""
	}

	title, description := "", ""

	var findMetaAndTitle func(*html.Node)
	findMetaAndTitle = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "meta" {
			var name, content string
			for _, attr := range node.Attr {
				if attr.Key == "name" {
					name = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}

			if name == "description" {
				description = content
			}
		}

		if node.Type == html.ElementNode && node.Data == "title" {
			if node.FirstChild == nil {
				title = " "
			} else {
				title = node.FirstChild.Data
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findMetaAndTitle(child)
		}
	}

	findMetaAndTitle(node)

	return title, description

}

func getPageHeadings(node *html.Node) string {
	if node == nil {
		return ""
	}

	var headings strings.Builder

	var findH1 func(*html.Node)
	findH1 = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "h1" {
			headings.WriteString(node.FirstChild.Data)
			headings.WriteString(", ")
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			findH1(c)
		}
	}

	getPageHeadings(node)

	return strings.TrimSuffix(headings.String(), ",")
}
