package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JamesClonk/go-todotxt"
	"github.com/perhenrik/timesheet-txt/model"

	"github.com/perhenrik/timesheet-txt/util"

	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/report"
)

var commandName = ""
var timesheetFilename = ""

const stopwatchTagKey = "clock"
const stopwatchTagValueRunning = "running"
const stopwatchStartTagKey = "start"

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
	case "start", "st":
		startStopwatch(flag.Args()[1:])
	case "stop", "sp":
		stopStopwatch()
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

func startStopwatch(arguments []string) {
	taskText := strings.TrimSpace(strings.Join(arguments, " "))
	if taskText == "" {
		fmt.Println("missing task text")
		usageWithHelp()
		return
	}

	now := time.Now().UTC()
	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()

	stoppedCount, _, err := stopRunningTasks(&tasklist, now)
	util.Check(err)

	task, err := todotxt.ParseTask(taskText)
	util.Check(err)
	markTaskAsRunning(task, now)
	tasklist.AddTask(task)
	file.WriteFile(tasklist)

	if stoppedCount > 0 {
		fmt.Printf("Stopped %d running stopwatch and started a new one.\n", stoppedCount)
	}
	fmt.Printf("Started stopwatch for: %s\n", task.Task())
}

func stopStopwatch() {
	now := time.Now().UTC()
	file := file.TimesheetFile{Name: timesheetFilename}
	tasklist := file.ReadFile()

	stoppedCount, totalHours, err := stopRunningTasks(&tasklist, now)
	util.Check(err)

	if stoppedCount == 0 {
		fmt.Println("No running stopwatch found.")
		return
	}

	file.WriteFile(tasklist)
	if stoppedCount == 1 {
		fmt.Printf("Stopped stopwatch (%.2f hours).\n", totalHours)
		return
	}

	fmt.Printf("Stopped %d running stopwatches (%.2f hours total).\n", stoppedCount, totalHours)
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

func stopRunningTasks(tasklist *todotxt.TaskList, stopAt time.Time) (stoppedCount int, totalHours float64, err error) {
	for i := range *tasklist {
		task := (*tasklist)[i]
		if !isRunningTask(task) {
			continue
		}

		hours, stopErr := markTaskAsStopped(&task, stopAt)
		if stopErr != nil {
			return stoppedCount, totalHours, stopErr
		}

		(*tasklist)[i] = task
		stoppedCount++
		totalHours += hours
	}

	return stoppedCount, totalHours, nil
}

func isRunningTask(task todotxt.Task) bool {
	return !task.Completed && task.AdditionalTags != nil && task.AdditionalTags[stopwatchTagKey] == stopwatchTagValueRunning
}

func markTaskAsRunning(task *todotxt.Task, startedAt time.Time) {
	if task.AdditionalTags == nil {
		task.AdditionalTags = make(map[string]string)
	}

	task.Completed = false
	task.CompletedDate = time.Time{}
	delete(task.AdditionalTags, "hours")
	task.AdditionalTags[stopwatchTagKey] = stopwatchTagValueRunning
	task.AdditionalTags[stopwatchStartTagKey] = strconv.FormatInt(startedAt.Unix(), 10)
}

func markTaskAsStopped(task *todotxt.Task, stoppedAt time.Time) (hours float64, err error) {
	if !isRunningTask(*task) {
		return 0, errors.New("task is not running")
	}

	startedAtRaw := task.AdditionalTags[stopwatchStartTagKey]
	if startedAtRaw == "" {
		return 0, errors.New("running task has no start timestamp")
	}

	startedAtUnix, parseErr := strconv.ParseInt(startedAtRaw, 10, 64)
	if parseErr != nil {
		return 0, parseErr
	}

	startedAt := time.Unix(startedAtUnix, 0).UTC()
	if stoppedAt.Before(startedAt) {
		return 0, errors.New("stop time is before start time")
	}

	hours = stoppedAt.Sub(startedAt).Hours()

	task.Complete()
	task.CompletedDate = stoppedAt
	delete(task.AdditionalTags, stopwatchTagKey)
	delete(task.AdditionalTags, stopwatchStartTagKey)
	task.AdditionalTags["hours"] = strconv.FormatFloat(hours, 'f', 4, 64)

	return hours, nil
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

	start|st +<project> [task:<taskname>]
	    Description:
	        Starts a stopwatch for a new task.
	        If another stopwatch is running, it is stopped automatically.

	stop|sp
	    Description:
	        Stops the running stopwatch and stores worked hours.
	
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
