package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CrawledUrl struct {
	ID              string         `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Url             string         `json:"url" gorm:"unique;not null"`
	Success         bool           `json:"success" gorm:"not null"`
	CrawlDuration   time.Duration  `json:"crawlDuration"`
	ResponseCode    int            `json:"responseCode"`
	PageTitle       string         `json:"pageTitle"`
	PageDescription string         `json:"pageDescription"`
	Heading         string         `json:"heading"`
	LastTested      *time.Time     `json:"lastTested"`
	Indexed         bool           `json:"indexed" gorm:"default:false"`
	CreatedAt       *time.Time     `json:"autoCreatedAt"`
	UpdatedAt       time.Time      `json:"autoUpdatedAt"`
	DeletedAt       gorm.DeletedAt `json:"index"`
}

func (crawl *CrawledUrl) UpdatedUrl(input CrawledUrl) error {
	tx := DBConn.Select("url", "success", "crawl_duration", "response_code", "page_title", "page_description", "headings", "last_tested", "updated_at").Omit("created_at").Save(&input)

	if tx.Error != nil {
		fmt.Println(tx.Error)
		return tx.Error
	}

	return nil
}

func (crawl *CrawledUrl) GetNextCrawlUrls(limit int) ([]CrawledUrl, error) {
	var crawledUrls []CrawledUrl

	if err := DBConn.Where("last_tested IS NULL").Limit(limit).Find(&crawledUrls).Error; err != nil {
		return []CrawledUrl{}, err
	}

	return crawledUrls, nil
}

func (crawl *CrawledUrl) Save() error {
	tx := DBConn.Save(&crawl)

	if tx.Error != nil {
		fmt.Println(tx.Error)
		return tx.Error
	}

	return nil
}

func (crawl *CrawledUrl) GetNotIndex() ([]CrawledUrl, error) {
	var crawledUrls []CrawledUrl

	if err := DBConn.Where("indexed = ? AND last_tested IS NOT NULL", false).Find(&crawledUrls).Error; err != nil {
		return []CrawledUrl{}, err
	}

	return crawledUrls, nil
}

func (crawl *CrawledUrl) SetIndexedTrue(urls []CrawledUrl) error {
	for _, url := range urls {
		url.Indexed = true
		tx := DBConn.Save(&url)

		if tx.Error != nil {
			fmt.Println(tx.Error)
			return tx.Error
		}
	}

	return nil
}
