package file

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// Line is a stuct holding information about the time worked.
type Line struct {
	Index    int
	Date     string
	Duration string
	Task     string
}

// CreateLineFromString parses a string and returns a new WorkTime.
// The input string must be separated by spaces.
func CreateLineFromString(s string) (worktime Line, err error) {
	elements := strings.Split(s, " ")
	Line := Line{
		Date:     time.Now().Format("2006-01-02"),
		Duration: "1h",
		Task:     "foobar"}

	dateMatcher := regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}$")
	durationMatcher := regexp.MustCompile("^\\d{1,2}[hm]$")

	modified := false

	for _, element := range elements {
		if dateMatcher.MatchString(element) {
			_, err := time.Parse("2006-01-02", element)
			if err == nil {
				Line.Date = element
				modified = true
			}

		} else if durationMatcher.MatchString(element) {
			Line.Duration = element
			modified = true

		} else {
			Line.Task = element
		}
	}

	if !modified {
		err = errors.New("Could not parse \"" + s + "\"")
	}

	return Line, err
}
