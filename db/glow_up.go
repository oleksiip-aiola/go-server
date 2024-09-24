package db

import (
	"time"
)

type MoodScore struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	UserId    string    `json:"userId"`
	Year      int32     `json:"year"`
	Month     int32     `json:"month"`
	Day       int32     `json:"day"`
	MoodId    int32     `json:"moodId"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type Day struct {
	Day int
}

type Month struct {
	Day Day
}

type Year struct {
	Month Month
}

type MoodDto struct {
	Year Year
}

func CreateMoodScore(userId string, year int32, month int32, day int32, moodId int32) (MoodScore, error) {
	moodScore := MoodScore{
		UserId: userId,
		Year:   year,
		Month:  month,
		Day:    day,
		MoodId: moodId,
	}

	if err := DBConn.Create(&moodScore).Error; err != nil {
		return MoodScore{}, err
	}

	return moodScore, nil
}

func UpdateMoodScore(id string, moodId int32) error {

	if err := DBConn.Model(&MoodScore{}).Where("id = ?", id).Update("mood_id", moodId).Error; err != nil {
		return err
	}

	return nil
}

func GetMoodScores(userId string, year int, month int) (map[int32]map[int32]map[int32]MoodScore, error) {
	var moodScores []MoodScore
	result := make(map[int32]map[int32]map[int32]MoodScore)

	if err := DBConn.Table(`mood_scores`).Where("user_id = ? AND (year = ? AND month = ?) OR (year = ? AND month = ?) OR (year = ? AND month = ?)", userId, year, month, year, month+1, year, month-1).Find(&moodScores).Error; err != nil {
		return nil, err
	}

	// for _, moodScore := range moodScores {
	// 	if _, ok := result[moodScore.Year]; !ok {
	// 		result[moodScore.Year] = make(map[int32]map[int32][]MoodScore)
	// 	}
	// 	if _, ok := result[moodScore.Year][moodScore.Month]; !ok {
	// 		result[moodScore.Year] = make(map[int32]map[int32][]MoodScore)
	// 	}
	// 	result[moodScore.Year][moodScore.Month][moodScore.Day] = append(result[moodScore.Year][moodScore.Month][moodScore.Day], moodScore)
	// }

	for _, moodScore := range moodScores {
		if _, ok := result[moodScore.Year]; !ok {
			result[moodScore.Year] = make(map[int32]map[int32]MoodScore)
		}
		if _, ok := result[moodScore.Year][moodScore.Month]; !ok {
			result[moodScore.Year][moodScore.Month] = make(map[int32]MoodScore)
		}
		result[moodScore.Year][moodScore.Month][moodScore.Day] = moodScore
	}

	return result, nil
}
