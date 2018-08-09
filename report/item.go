package report

import "time"

// Item is a stuct holding information about the time worked on a task on a given day.
type Item struct {
	Date  time.Time
	Hours float64
	Task  string
}
