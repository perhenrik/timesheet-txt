package file

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/perhenrik/timesheet-txt/model"
	"github.com/perhenrik/timesheet-txt/util"
)

func timesheetFile() string {
	filename := ".timesheet.txt"
	homeDirectory, err := homedir.Dir()
	util.Check(err)
	return filepath.Join(homeDirectory, filename)
}

func writeLine(file *os.File, workAmount model.Work) {
	_, err := file.WriteString(workAmount.String() + "\n")
	util.Check(err)
}

// AppendToFile append a work time item to the timesheet file
func AppendToFile(workAmount model.Work) {
	file, err := os.OpenFile(timesheetFile(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	util.Check(err)

	defer func() {
		cerr := file.Close()
		util.Check(cerr)
	}()

	writeLine(file, workAmount)
}

// ReadFile reads in and parses the default timesheet file
func ReadFile() (workAmounts []model.Work) {
	file, err := os.OpenFile(timesheetFile(), os.O_RDONLY|os.O_CREATE, 0600)
	util.Check(err)

	defer func() {
		cerr := file.Close()
		util.Check(cerr)
	}()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			index++
			workAmount, err := model.CreateWorkAmountFromString(scanner.Text())
			if err == nil {
				workAmount.Index = index
				workAmounts = append(workAmounts, workAmount)
			}
		}
	}
	util.Check(scanner.Err())

	sort.Slice(workAmounts[:], func(i, j int) bool {
		return workAmounts[i].String() < workAmounts[j].String()
	})

	return workAmounts
}

// WriteFile overwrites the default timesheet file with the values in the supplied array
func WriteFile(lines []model.Work) {
	file, err := os.OpenFile(timesheetFile(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	util.Check(err)

	defer func() {
		cerr := file.Close()
		util.Check(cerr)
	}()

	for _, line := range lines {
		writeLine(file, line)
	}
}
