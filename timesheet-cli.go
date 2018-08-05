package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/model"

	"github.com/fatih/color"
)

type timeSlice struct {
	date       string
	offsetDays string
}

func main() {

	if len(os.Args) < 2 {
		usage(os.Args[0])
		return
	}

	if os.Args[1] == "add" {
		add(os.Args[2:])
	} else if os.Args[1] == "list" {
		list()
	} else {
		usage(os.Args[0])
	}
}

func usage(name string) {
	fmt.Println("Usage: " + path.Base(name) + " add [date] duration task")
}

func list() {
	workItems := file.ReadFile()
	for _, workItem := range workItems {
		fmt.Printf("%d: %s %s %s\n", workItem.Index, workItem.Date, workItem.Duration, workItem.Task)
	}
}

func add(arguments []string) {
	s := strings.Join(arguments, " ")
	workTime := model.CreateFromString(s)
	file.AppendToFile(workTime)
	color.Yellow("Adding %s", workTime)
}
