package util

// MakeSureArrayHasEnoughElements adds empty elements to an array unil it has the expected number of elements
func MakeSureArrayHasEnoughElements(array []string, numberOfElements int) []string {
	if len(array) < numberOfElements {
		array = append(array, "")
	}
	return array
}
