package search

import (
	"fmt"
	"time"

	"github.com/alexey-petrov/go-server/db"
)

func RunEngine() {
	fmt.Println("Running search engine")
	defer fmt.Println("Search Engine stopped")

	settings := &db.SearchSettings{}

	searchSettings, err := settings.GetSearchSettings()

	if err != nil {
		fmt.Println("Error getting search settings:", err)
		return
	}

	if !searchSettings.SearchOn {
		fmt.Println("Search is turned off")
		return
	}

	crawl := &db.CrawledUrl{}
	nextUrls, err := crawl.GetNextCrawlUrls(int(settings.Amount))

	if err != nil {
		fmt.Println("Error getting next crawl urls:", err)
		return
	}

	newUrls := []db.CrawledUrl{}
	testedTime := time.Now()

	for _, next := range nextUrls {
		result := runCrawl(next.Url)

		if !result.Success {
			err := next.UpdatedUrl(db.CrawledUrl{
				ID:              next.ID,
				Url:             next.Url,
				Success:         false,
				CrawlDuration:   result.CrawlData.CrawlTime,
				ResponseCode:    result.ResponseCode,
				PageTitle:       result.CrawlData.PageTitle,
				PageDescription: result.CrawlData.PageDescription,
				Heading:         result.CrawlData.Headings,
				LastTested:      &testedTime,
			})

			if err != nil {
				fmt.Println("Error updating url:", err)
			}

			continue
		} else {
			err := next.UpdatedUrl(db.CrawledUrl{
				ID:              next.ID,
				Url:             next.Url,
				Success:         result.Success,
				CrawlDuration:   result.CrawlData.CrawlTime,
				ResponseCode:    result.ResponseCode,
				PageTitle:       result.CrawlData.PageTitle,
				PageDescription: result.CrawlData.PageDescription,
				Heading:         result.CrawlData.Headings,
				LastTested:      &testedTime,
				Indexed:         true,
			})

			if err != nil {
				fmt.Println("Error updating a successfull url:", err)
				fmt.Println(next.Url)
			}

			for _, link := range result.CrawlData.Links.Internal {
				newUrls = append(newUrls, db.CrawledUrl{
					Url: link,
				})
			}
		}

		if !settings.AddNew {
			return
		}

		for _, newUrl := range newUrls {
			err := newUrl.Save()

			if err != nil {
				fmt.Println("Error saving new url:", err)
			}
		}

		fmt.Printf("Crawled & added %d urls\n", len(nextUrls))

	}

}

func RunIndex() {
	fmt.Println("Running search indexing")
	defer fmt.Println("Stopped search indexing")

	crawled := &db.CrawledUrl{}

	notIndexed, err := crawled.GetNotIndex()

	if err != nil {
		return
	}

	index := make(Index)
	index.Add(notIndexed)

	searchIndex := &db.SearchIndex{}

	if err = searchIndex.Save(index, notIndexed); err != nil {
		return
	}

	if err = crawled.SetIndexedTrue(notIndexed); err != nil {
		return
	}
}
