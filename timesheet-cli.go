package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

type TimeEntry struct {
	date     string
	duration string
	task     string
}

func main() {
	if len(os.Args) < 2 {
		usage(os.Args[0])
		return
	}

	if os.Args[1] == "add" {
		add(os.Args[2:])
	} else {
		usage(os.Args[0])
	}
}

func usage(name string) {
	fmt.Println("Usage: " + path.Base(name) + " add [date] duration task")
}

func add(arguments []string) {
	timeEntry := TimeEntry{
		date:     time.Now().Format("2006-01-02"),
		duration: "1h",
		task:     "foobar"}

	dateMatcher := regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}$")
	durationMatcher := regexp.MustCompile("^\\d{1,2}[hm]$")

	for _, argument := range arguments {
		if dateMatcher.MatchString(argument) {
			_, err := time.Parse("2006-01-02", argument)
			if err == nil {
				timeEntry.date = argument
			}

		} else if durationMatcher.MatchString(argument) {
			timeEntry.duration = argument

		} else {
			timeEntry.task = argument
		}
	}
	appendToFile(timesheetFile(), timeEntry)
	color.Yellow("Adding %s", timeEntry)
}

func timesheetFile() string {
	filename := ".timesheet.txt"
	homeDirectory, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDirectory, filename)
}

func appendToFile(filename string, timeEntry TimeEntry) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(timeEntry.date + " " + timeEntry.duration + " " + timeEntry.task + "\n"); err != nil {
		panic(err)
	}
}
