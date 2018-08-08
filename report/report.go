package report

import (
	"sort"
	"strings"
	"time"

	"github.com/perhenrik/timesheet-txt/file"
)

//Create returns an array of sorted report items
func Create(workItems []file.Line, endTime time.Time, taskDuration string) (items []Item) {
	hours, err := ParseDuration(taskDuration)
	if err != nil {
		return
	}

	itemMap := make(map[string]int)
	startTime := endTime.Add(time.Hour * -time.Duration(hours))
	for _, line := range workItems {
		if dateInRange(line.Time, startTime, endTime) {
			hours, _ := ParseDuration(line.Duration)
			itemMap[line.Time.Format("2006-01-02")+"^"+line.Task] += hours
		}
	}

	keys := make([]string, len(itemMap))
	i := 0
	for k := range itemMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, key := range keys {
		itemTime, itemTask := splitKey(key)
		r := Item{Date: itemTime, Hours: itemMap[key], Task: itemTask}
		items = append(items, r)
	}
	return items
}

func dateInRange(current time.Time, start time.Time, end time.Time) (result bool) {
	// make sure we ignore time and only check dates and remove on day from start and add a day to end
	current = stripTime(current)
	start = stripTime(start.Add(time.Hour * -24))
	end = stripTime(end.Add(time.Hour * 24))

	if current.After(start) && current.Before(end) {
		return true
	}
	return false
}

func stripTime(t time.Time) (newTime time.Time) {
	newTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return
}

func splitKey(key string) (keyTime time.Time, keyTask string) {
	items := strings.Split(key, "^")
	if len(items) == 2 {
		keyTime, _ = time.Parse("2006-01-02", items[0])
		keyTask = items[1]
	}
	return
}
