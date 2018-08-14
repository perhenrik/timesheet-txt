package util

import (
	"errors"

	"github.com/perhenrik/timesheet-txt/model"
)

// MakeSureArrayHasEnoughElements adds empty elements to an array unil it has the expected number of elements.
func MakeSureArrayHasEnoughElements(in []string, length int) (out []string) {
	out = make([]string, length)
	copy(out, in)
	return
}

// DeleteFromArray deletes an item from an array and returns the item removed.
func DeleteFromArray(array []model.Work, index int) (newArray []model.Work, deletedWork model.Work, err error) {
	newArray = array
	if index >= 0 && index < len(array) {
		deletedWork = array[index]
		newArray = append(array[:index], array[index+1:]...)
	} else {
		err = errors.New("index out of bounds")
	}
	return
}
