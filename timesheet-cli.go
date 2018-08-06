package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/model"

	"github.com/fatih/color"
)

type timeSlice struct {
	date       string
	offsetDays string
}

var commandName = ""

func main() {
	commandName = path.Base(os.Args[0])

	if len(os.Args) < 2 {
		usageWithHelp()
		return
	}

	if os.Args[1] == "add" {
		add(os.Args[2:])
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
	workTime, err := model.CreateFromString(s)
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
		fmt.Printf("%4d: %s %s %s\n", workItem.Index, workItem.Date, workItem.Duration, workItem.Task)
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
			Prints a summarized time report.
		Arguments
			date:   the date wich is the end of the report period, defaults to now.
			period: the duration of the report, defaults to 5 days (5d)
  `)
}
