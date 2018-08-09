package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/report"
)

var commandName = ""

func main() {
	commandName = path.Base(os.Args[0])

	if len(os.Args) < 2 {
		usageWithHelp()
		return
	}

	if os.Args[1] == "add" {
		add(os.Args[2:])
	} else if os.Args[1] == "report" {
		createReport()
	} else if os.Args[1] == "list" {
		list()
	} else if os.Args[1] == "tidy" {
		tidy()
	} else if os.Args[1] == "delete" {
		index, err := strconv.Atoi(os.Args[2])
		if err != nil {
			usageWithHelp()
			return
		}
		delete(index)
	} else if os.Args[1] == "help" {
		help()
	} else {
		usageWithHelp()
	}
}

func add(arguments []string) {
	s := strings.Join(arguments, " ")
	workTime, err := file.CreateLineFromString(s)
	if err != nil {
		color.Red(err.Error())
	} else {
		file.AppendToFile(workTime)
		color.Yellow("Adding %s", workTime)
	}
}

func list() {
	workItems := file.ReadFile()
	for _, workItem := range workItems {
		fmt.Printf("%4d: %s %s %s\n", workItem.Index, workItem.Time.Format("2006-01-02"), workItem.Duration, workItem.Task)
	}
}

func delete(index int) {
	index--
	workTimes := file.ReadFile()
	if index < 0 || index > len(workTimes)-1 {
		return
	}
	workTimes = append(workTimes[:index], workTimes[index+1:]...)
	file.WriteFile(workTimes)
}

func tidy() {
	workItems := file.ReadFile()
	file.WriteFile(workItems)
}

func createReport() {
	workItems := file.ReadFile()
	reportItems := report.Create(workItems, time.Now(), "5d")

	format := "%12s%31s%6s%7s%7s\n"
	previousDate := ""
	dailyTotal := 0.0
	total := 0.0
	for _, reportItem := range reportItems {
		currentDate := reportItem.Date.Format("2006-02-02")
		if previousDate != currentDate && dailyTotal != 0 {
			fmt.Printf(format, "", "", "", padLeft(fmt.Sprintf("%.1f", dailyTotal), ".", 7), "")
			previousDate = currentDate
			dailyTotal = 0
		}
		dailyTotal += reportItem.Hours
		total += reportItem.Hours
		task := padRight(clipString(reportItem.Task, 30), ".", 30)
		fmt.Printf(format, currentDate, task, fmt.Sprintf("%.1f", reportItem.Hours), "", "")
	}
	fmt.Printf(format, "", "", "", padLeft(fmt.Sprintf("%.1f", dailyTotal), ".", 7), "")
	fmt.Printf(format, "", "", "", "", padLeft(fmt.Sprintf("%.1f", total), ".", 7))
}

func padRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func padLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[(len(str) - lenght):]
		}
	}
}

func clipString(s string, length int) string {
	clipped := s
	if len(s) > length+1 {
		clipped = s[:length-3] + "..."
	}
	return clipped
}

func usage() {
	fmt.Println("Usage: " + commandName + " action [parameters]")
}

func usageWithHelp() {
	usage()
	fmt.Println("Try '" + commandName + " help' for more information.")
}

func help() {
	usage()
	fmt.Println()
	fmt.Println(`Actions:

	add [date] duration task
	    Description:
	        Appends a new task with the given duration to the timesheet file
		Arguments:
            duration: <number>[h|m], examples: 1h (one hour), 30m (30 minutes)
            task: free text string
	
	list
	    Description:
	        Lists all work registered, more or less a cat of the timesheet file.
            All lines are prepended with a number wich can be used in other action, ie. delete.
	
	delete [number]
	    Description:
		    Deletes the work identified by number. This number can found using the list action.
		Arguments:
		    number: the work item to delere

	tidy
	    Description:
		    Cleans up the timesheet file. Note: this action will overwrite your timesheetfile.
		
	report [date] [period]
		Description:
			Prints a summarized time report. All tasks on the same date are summarized.
		Arguments
			date:   the date wich is the end of the report period, defaults to now.
			period: the duration of the report counting backwords from date. Defaults to 5 days (5d)
  `)
}
