package file

import (
	"bufio"
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/perhenrik/timesheet-txt/model"
)

func timesheetFile() string {
	filename := ".timesheet.txt"
	homeDirectory, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDirectory, filename)
}

// AppendToFile append a work time item to the timesheet file
func AppendToFile(workTime model.WorkTime) {
	f, err := os.OpenFile(timesheetFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(workTime.Date + " " + workTime.Duration + " " + workTime.Task + "\n"); err != nil {
		panic(err)
	}
}

// ReadFile reads in and parses the timesheet file
func ReadFile() (workTimes []model.WorkTime) {
	file, err := os.Open(timesheetFile())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		index++
		workTime := model.CreateFromString(scanner.Text())
		workTime.Index = index
		workTimes = append(workTimes, workTime)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return workTimes
}
