package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JamesClonk/go-todotxt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/perhenrik/timesheet-txt/util"
)

// TimesheetFile handles all interactions with the file containing the Work
type TimesheetFile struct {
	Name string
}

func (f TimesheetFile) String() string {
	return fmt.Sprintf("%s", f.Name)
}

// DefaultFileName returns the default timesheet filename
func DefaultFileName() string {
	homeDirectory, err := homedir.Dir()
	util.Check(err)
	return filepath.Join(homeDirectory, ".timesheet.txt")
}

// ReadFile reads in and parses the default timesheet file
func (f TimesheetFile) ReadFile() (tasklist todotxt.TaskList) {
	file, err := os.OpenFile(f.Name, os.O_RDONLY|os.O_CREATE, 0600)
	util.Check(err)

	defer func() {
		cerr := file.Close()
		util.Check(cerr)
	}()

	tasklist, err = todotxt.LoadFromFile(file)
	util.Check(err)

	return tasklist
}

// WriteFile overwrites the default timesheet file with the values in the supplied array
func (f TimesheetFile) WriteFile(tasklist todotxt.TaskList) {
	file, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	util.Check(err)

	defer func() {
		cerr := file.Close()
		util.Check(cerr)
	}()

	todotxt.WriteToFile(&tasklist, file)
}
