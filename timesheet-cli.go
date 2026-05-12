package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JamesClonk/go-todotxt"
	"github.com/perhenrik/timesheet-txt/model"

	"github.com/perhenrik/timesheet-txt/util"

	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/report"
)

var commandName = ""
var timesheetFilename = ""

func main() {
	commandName = filepath.Base(os.Args[0])

	flag.Usage = func() {
		usageWithHelp()
	}

	flag.StringVar(&timesheetFilename, "f", file.DefaultFileName(), "the timesheet filename")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("no action provided")
		usageWithHelp()
		return
	}

	switch flag.Arg(0) {
	case "add", "a":
		add(flag.Args()[1:])
	case "report", "r":
		createReport(flag.Args()[1:])
	case "list", "l", "ls":
		list()
	case "delete", "d", "del":
		deleteEntry(flag.Args()[1:])
	case "help", "h":
		help()
	default:
		fmt.Println("action provided but not defined: " + flag.Arg(0))
		usageWithHelp()
	}
}

func add(arguments []string) {
	s := strings.Join(arguments, " ")
	task, err := todotxt.ParseTask(s)
	util.Check(err)
	task.Complete()

	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()
	tasklist.AddTask(task)
	file.WriteFile(tasklist)

	fmt.Printf("Added: %s\n", task)
}

func list() {
	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()
	for _, task := range tasklist {
		fmt.Printf("%d: %s\n", task.Id, task)
	}
}

func deleteEntry(arguments []string) {
	arguments = util.MakeSureArrayHasEnoughElements(arguments, 1)
	index, err := strconv.Atoi(arguments[0])
	util.Check(err)

	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()
	err = tasklist.RemoveTaskById(index)
	util.Check(err)

	file.WriteFile(tasklist)
}

func createReport(arguments []string) {
	s := strings.Join(arguments, " ")
	workTime, err := model.CreateWorkFromString(s)
	util.Check(err)

	reportType := workTime.Task
	if reportType == "" {
		reportType = "simple"
	}

	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()
	reportItems := report.Create(tasklist, workTime.Date, workTime.Hours)

	var theReport string
	switch reportType {
	case "summary":
		theReport = report.Summary(reportItems)
	default:
		theReport = report.Simple(reportItems)
	}

	fmt.Print(theReport)
}

func usage() {
	fmt.Println("Usage: " + commandName + " [-f filename] action [parameters]")
}

func usageWithHelp() {
	usage()
	fmt.Println("Try '" + commandName + " help' for more information.")
}

func help() {
	usage()
	fmt.Println()
	fmt.Println(`Actions:

	add|a [date] +<project> [task:<taskname>] hours:<number>
	    Description:
	        Appends a new project/task with the given duration to the timesheet file
		Arguments:
			date: the date the work was performed. Defaults to today.
            project: name of the project
			taskname: free text string
			number: the hours worked (float)
	
	list|ls|l
	    Description:
	        Lists all registered work items, effectively a cat of the timesheet file.
            All lines are prepended with an ID that can be used in other actions, e.g. delete.
	
	delete|del|d [number]
	    Description:
		    Deletes the work identified by number. This number can be found using the list action.
		Arguments:
		    number: the work item to delete
		
	report|r [date] [period] [type]
		Description:
			Prints a time report. All tasks on the same date are summarized.
		Arguments
			date:   the date that is the end of the report period, defaults to now.
			period: the duration of the report counting backwards from date. Defaults to 5 days (5d)
			type:	The report type (simple|summary). Defaults to 'simple'
  `)
}
