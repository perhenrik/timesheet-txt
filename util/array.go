package util

import "github.com/perhenrik/timesheet-txt/model"

// MakeSureArrayHasEnoughElements adds empty elements to an array unil it has the expected number of elements.
func MakeSureArrayHasEnoughElements(in []string, length int) (out []string) {
	out = make([]string, length)
	copy(out, in)
	return
}

// DeleteFromArray deletes an item from an array and returns the item removed.
func DeleteFromArray(array []model.Work, index int) (deletedItem model.Work) {
	if index < 0 || index > len(array)-1 {
		return
	}
	deletedItem = array[index]
	array = append(array[:index], array[index+1:]...)
	return
}
