package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/perhenrik/timesheet-txt/model"

	"github.com/perhenrik/timesheet-txt/util"

	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/report"
)

var commandName = ""
var timesheetFilename = ""

func main() {
	commandName = path.Base(os.Args[0])

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
	case "tidy", "t":
		tidy()
	case "delete", "d", "del":
		delete(flag.Args()[1:])
	case "help", "h":
		help()
	default:
		fmt.Println("action provided but not defined: " + flag.Arg(0))
		usageWithHelp()
	}
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

func add(arguments []string) {
	s := strings.Join(arguments, " ")
	workTime, err := model.CreateWorkFromString(s)
	util.Check(err)

	file := file.TimesheetFile{Name: timesheetFilename}
	file.AppendToFile(workTime)
	fmt.Printf("Added: %s\n", workTime)
}

func list() {
	file := file.TimesheetFile{Name: timesheetFilename}
	workItems := file.ReadFile()
	for _, workItem := range workItems {
		fmt.Printf("%4d: %s\n", workItem.Index, workItem.String())
	}
}

func delete(arguments []string) {
	arguments = util.MakeSureArrayHasEnoughElements(arguments, 1)
	index, err := strconv.Atoi(arguments[0])
	util.Check(err)

	index--
	file := file.TimesheetFile{Name: timesheetFilename}
	workList := file.ReadFile()
	newWorkList, deletedWorkItem, err := util.DeleteFromArray(workList, index)
	util.Check(err)

	file.WriteFile(newWorkList)
	fmt.Printf("Deleted: %s\n", deletedWorkItem)
}

func tidy() {
	file := file.TimesheetFile{Name: timesheetFilename}
	workItems := file.ReadFile()
	file.WriteFile(workItems)
}

func createReport(arguments []string) {
	s := strings.Join(arguments, " ")
	workTime, err := model.CreateWorkFromString(s)
	util.Check(err)

	reportName := workTime.Task;
	if(reportName == "") {
		reportName = "simple"
	}
	file := file.TimesheetFile{Name: timesheetFilename}
	workItems := file.ReadFile()
	reportItems := report.Create(workItems, workTime.Date, workTime.Hours)

	var theReport = ""
	switch workTime.Task {
	case "summary":
		theReport = report.Summary(reportItems)
	default:
		theReport = report.Simple(reportItems)
	}
	//theReport := report.Simple(reportItems)

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

	add|a [date] duration task
	    Description:
	        Appends a new task with the given duration to the timesheet file
		Arguments:
            duration: <number>[h|m], examples: 1h (one hour), 30m (30 minutes)
            task: free text string
	
	list|ls|l
	    Description:
	        Lists all work registered, more or less a cat of the timesheet file.
            All lines are prepended with a number wich can be used in other action, ie. delete.
	
	delete|del|d [number]
	    Description:
		    Deletes the work identified by number. This number can found using the list action.
		Arguments:
		    number: the work item to delere

	tidy|t
	    Description:
		    Cleans up the timesheet file. Note: this action will overwrite your timesheetfile.
		
	report|r [date] [period] [type]
		Description:
			Prints a time report. All tasks on the same date are summarized.
		Arguments
			date:   the date wich is the end of the report period, defaults to now.
			period: the duration of the report counting backwords from date. Defaults to 5 days (5d)
			type:	The report type (simple|summary). Defaults to 'simple'
  `)
}
