package gormdb

import "time"

type SearchSettings struct {
	ID uint `gorm:"primaryKey" json:"id"`
	SearchOn bool `json:"searchOn"`
	AddNew bool `json:"addNew"`
	Amount int32 `json:"amount"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *SearchSettings) UpdateSearchSettings(searchOn bool, addNew bool, amount int32) error {
	s.SearchOn = searchOn
	s.AddNew = addNew
	s.Amount = amount
	s.ID = 1;

	if err := DBConn.Save(&s).Error; err != nil {
		return err
	}

	return nil
}

func (s *SearchSettings) GetSearchSettings() (*SearchSettings, error) {
	var searchSettings SearchSettings

	if err := DBConn.First(&searchSettings).Error; err != nil {
		return nil, err
	}

	return &searchSettings, nil
}