package util

// MakeSureArrayHasEnoughElements adds empty elements to an array unil it has the expected number of elements
func MakeSureArrayHasEnoughElements(in []string, length int) (out []string) {
	out = make([]string, length)
	copy(out, in)
	return
}
