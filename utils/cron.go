package utils

import (
	"fmt"

	"github.com/alexey-petrov/go-server/search"
	"github.com/robfig/cron"
)

func StartCronJobs() {
	c := cron.New()
	c.AddFunc("0 * * * *", search.RunEngine)
	c.AddFunc("15 * * * *", search.RunIndex)

	cronCount := len(c.Entries())

	fmt.Printf("Cron job count: %d\n", cronCount)
}
