package model

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Work is a stuct holding information about the time worked.
type Work struct {
	Index int
	Date  time.Time
	Hours float64
	Task  string
}

var durationMatcher = regexp.MustCompile(`^(\d{1,4}(\.\d){0,1})([mhdw]{1})$`)
var dateMatcher = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (work Work) String() string {
	return work.Date.Format("2006-01-02") + " " + fmt.Sprintf("%.1f", work.Hours) + "h " + work.Task
}

// CreateWorkFromString parses a string and returns a new WorkTime.
// The input string must be separated by spaces.
func CreateWorkFromString(s string) (work Work, err error) {
	elements := strings.Split(s, " ")
	work = Work{
		Date:  time.Now(),
		Hours: 1,
		Task:  "foobar"}

	for _, element := range elements {
		if dateMatcher.MatchString(element) {
			time, cerr := time.Parse("2006-01-02", element)
			if cerr == nil {
				work.Date = time
			}
		} else if durationMatcher.MatchString(element) {
			hours, cerr := ParseDuration(element)
			if cerr == nil {
				work.Hours = hours
			}
		} else {
			work.Task = element
		}
	}
	return work, err
}

// ParseDuration parses a duration-string, eg. 5h, 3d, and returns the corresponding hours
func ParseDuration(duration string) (hours float64, err error) {

	matched := durationMatcher.FindStringSubmatch(duration)

	if matched != nil {
		value, cerr := strconv.ParseFloat(matched[1], 64)
		if cerr != nil {
			return 0, cerr
		}

		if matched[3] == "h" {
			hours = value
		} else if matched[3] == "d" {
			hours = value * 24
		} else if matched[3] == "w" {
			hours = value * 24 * 7
		} else if matched[3] == "m" {
			hours = value / 60
		}
	} else {
		err = errors.New("invalid duration: " + duration)
	}
	return hours, err
}
