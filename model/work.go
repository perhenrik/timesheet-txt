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

func (work Work) String() string {
	return work.Date.Format("2006-01-02") + " " + fmt.Sprintf("%.1f", work.Hours) + " " + work.Task
}

// CreateWorkAmountFromString parses a string and returns a new WorkTime.
// The input string must be separated by spaces.
func CreateWorkAmountFromString(s string) (workAmount Work, err error) {
	elements := strings.Split(s, " ")
	workAmount = Work{
		Date:  time.Now(),
		Hours: 1,
		Task:  "foobar"}

	dateMatcher := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	durationMatcher := regexp.MustCompile(`^\d{1,2}(\.\d){0,1}[hdwm]$`)

	for _, element := range elements {
		if dateMatcher.MatchString(element) {
			time, cerr := time.Parse("2006-01-02", element)
			if cerr == nil {
				workAmount.Date = time
			}
		} else if durationMatcher.MatchString(element) {
			hours, cerr := ParseDuration(element)
			if cerr == nil {
				workAmount.Hours = hours
			}
		} else {
			workAmount.Task = element
		}
	}
	return workAmount, err
}

// ParseDuration parses a duration-string, eg. 5h, 3d, and returns the corresponding hours
func ParseDuration(duration string) (hours float64, err error) {
	durationMatcher := regexp.MustCompile(`^(\d{1,2}(\.\d){0,1})([hdwm]{1})$`)
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
			hours = value * 24 * 7 * 30
		}
	} else {
		err = errors.New("invalid duration: " + duration)
	}

	return hours, err
}
