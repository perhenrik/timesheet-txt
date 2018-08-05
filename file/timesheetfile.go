package file

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func writeWorkTime(file *os.File, workTime model.WorkTime) {
	if _, err := file.WriteString(workTime.Date + " " + workTime.Duration + " " + workTime.Task + "\n"); err != nil {
		panic(err)
	}
}

// AppendToFile append a work time item to the timesheet file
func AppendToFile(workTime model.WorkTime) {
	file, err := os.OpenFile(timesheetFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	writeWorkTime(file, workTime)
}

// ReadFile reads in and parses the default timesheet file
func ReadFile() (workTimes []model.WorkTime) {
	file, err := os.OpenFile(timesheetFile(), os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			index++
			workTime := model.CreateFromString(scanner.Text())
			workTime.Index = index
			workTimes = append(workTimes, workTime)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return workTimes
}

// WriteFile overwrites the default timesheet file with the values in the supplied array
func WriteFile(workTimes []model.WorkTime) {
	file, err := os.OpenFile(timesheetFile(), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, workTime := range workTimes {
		writeWorkTime(file, workTime)
	}
}
