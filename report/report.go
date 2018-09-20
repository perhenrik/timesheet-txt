package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/perhenrik/timesheet-txt/model"
	"github.com/perhenrik/timesheet-txt/util"
)

// Summary sums up all tasks, one per line
func Summary(reportItems []model.Work) string {
	format := "%30s: %7s\n"
	var report strings.Builder

	var sums = make(map[string]float64)

	for _, reportItem := range reportItems {
		task := util.PadRight(util.ClipString(reportItem.Task, 30), ".", 30)
		sums[task] += reportItem.Hours
	}

	for task, hours := range sums {
		_, err := fmt.Fprintf(&report, format, task, fmt.Sprintf("%.1f", hours))
		util.Check(err)
	}
	return report.String()
}

// Simple builds a simple report based an an array of model.Work
func Simple(reportItems []model.Work) string {
	format := "%12s%31s%6s%7s%7s\n"
	previousDate := ""
	dailyTotal := 0.0
	total := 0.0
	var report strings.Builder

	for i, reportItem := range reportItems {
		currentDate := reportItem.Date.Format("2006-01-02")
		if previousDate != currentDate && i != 0 {
			_, err := fmt.Fprintf(&report, format, "", "", "", util.PadLeft(fmt.Sprintf("%.1f", dailyTotal), ".", 7), "")
			util.Check(err)
			dailyTotal = 0
		}
		dailyTotal += reportItem.Hours
		total += reportItem.Hours
		task := util.PadRight(util.ClipString(reportItem.Task, 30), ".", 30)
		_, err := fmt.Fprintf(&report, format, currentDate, task, fmt.Sprintf("%.1f", reportItem.Hours), "", "")
		util.Check(err)
		previousDate = currentDate
	}
	_, err := fmt.Fprintf(&report, format, "", "", "", util.PadLeft(fmt.Sprintf("%.1f", dailyTotal), ".", 7), "")
	util.Check(err)
	_, err = fmt.Fprintf(&report, format, "", "", "", "", util.PadLeft(fmt.Sprintf("%.1f", total), ".", 7))
	util.Check(err)

	return report.String()
}

//Create returns an array of sorted report items
func Create(workAmounts []model.Work, endTime time.Time, taskDuration float64) (items []model.Work) {
	itemMap := make(map[string]float64)
	startTime := endTime.Add(time.Hour * -time.Duration(taskDuration))
	for _, workAmount := range workAmounts {
		if dateInRange(workAmount.Date, startTime, endTime) {
			itemMap[workAmount.Date.Format("2006-01-02")+"^"+workAmount.Task] += workAmount.Hours
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
		r := model.Work{Date: itemTime, Hours: itemMap[key], Task: itemTask}
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
		itemDate, err := time.Parse("2006-01-02", items[0])
		if err == nil {
			keyTime = itemDate
		}
		keyTask = items[1]
	}
	return
}
