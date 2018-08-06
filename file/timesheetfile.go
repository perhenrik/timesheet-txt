package file

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

func timesheetFile() string {
	filename := ".timesheet.txt"
	homeDirectory, err := homedir.Dir()
	check(err)
	return filepath.Join(homeDirectory, filename)
}

func writeLine(file *os.File, line Line) {
	_, err := file.WriteString(line.Date + " " + line.Duration + " " + line.Task + "\n")
	check(err)
}

// AppendToFile append a work time item to the timesheet file
func AppendToFile(line Line) {
	file, err := os.OpenFile(timesheetFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	check(err)

	defer file.Close()

	writeLine(file, line)
}

// ReadFile reads in and parses the default timesheet file
func ReadFile() (lines []Line) {
	file, err := os.OpenFile(timesheetFile(), os.O_RDONLY|os.O_CREATE, 0600)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			index++
			line, err := CreateLineFromString(scanner.Text())
			if err == nil {
				line.Index = index
				lines = append(lines, line)
			}
		}
	}
	check(scanner.Err())

	return lines
}

// WriteFile overwrites the default timesheet file with the values in the supplied array
func WriteFile(lines []Line) {
	file, err := os.OpenFile(timesheetFile(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	check(err)
	defer file.Close()

	for _, line := range lines {
		writeLine(file, line)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
