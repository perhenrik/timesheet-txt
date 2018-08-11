package file

import (
	"regexp"
	"strings"
	"time"
)

// Line is a stuct holding information about the time worked.
type Line struct {
	Index    int
	Time     time.Time
	Duration string
	Task     string
}

func (line Line) String() string {
	return line.Time.Format("2006-01-02") + " " + line.Duration + " " + line.Task
}

// CreateLineFromString parses a string and returns a new WorkTime.
// The input string must be separated by spaces.
func CreateLineFromString(s string) (line Line, err error) {
	elements := strings.Split(s, " ")
	line = Line{
		Time:     time.Now(),
		Duration: "1h",
		Task:     "foobar"}

	dateMatcher := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	durationMatcher := regexp.MustCompile(`^\d{1,2}(\.\d){0,1}[hdwm]$`)

	for _, element := range elements {
		if dateMatcher.MatchString(element) {
			time, cerr := time.Parse("2006-01-02", element)
			if cerr == nil {
				line.Time = time
			}
		} else if durationMatcher.MatchString(element) {
			line.Duration = element

		} else {
			line.Task = element
		}
	}

	return line, err
}
