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

func main() {

	if len(os.Args) < 2 {
		usage(os.Args[0])
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
			usage(os.Args[0])
			return
		}
		delete(index)
	} else {
		usage(os.Args[0])
	}
}

func usage(name string) {
	fmt.Println("Usage: " + path.Base(name) + " add [date] duration task")
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
