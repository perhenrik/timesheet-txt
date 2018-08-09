package report

import (
	"errors"
	"regexp"
	"strconv"
)

// TaskDuration is a stuct holding information about the time worked on a task on a given day.
type TaskDuration struct {
	Hours float64
	Task  string
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
		err = errors.New("Invalid duration: " + duration)
	}

	return hours, err
}
