package report

import (
	"errors"
	"regexp"
	"strconv"
)

// TaskDuration is a stuct holding information about the time worked on a task on a given day.
type TaskDuration struct {
	Hours int
	Task  string
}

// ParseDuration parses a duration-string, eg. 5h, 3d, and returns the corresponding hours
func ParseDuration(duration string) (hours int, err error) {
	durationMatcher := regexp.MustCompile("^(\\d{1,2})([hdwm]{1})$")
	matched := durationMatcher.FindStringSubmatch(duration)

	if matched != nil {
		value, err := strconv.Atoi(matched[1])

		if err == nil && matched[2] == "h" {
			hours = value
		} else if matched[2] == "d" {
			hours = value * 24
		} else if matched[2] == "w" {
			hours = value * 24 * 7
		} else if matched[2] == "m" {
			hours = value * 24 * 7 * 30
		} else {
			value = 0
		}
	} else {
		err = errors.New("Invalid duration: " + duration)
	}

	return hours, err
}
