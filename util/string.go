package util

// PadRight pads a string to the right
func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// PadLeft pads a string to the left
func PadLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[(len(str) - lenght):]
		}
	}
}

// ClipString cuts of a string at length-3 and adds ... at the end
func ClipString(str string, length int) string {
	clipped := str
	if len(str) > length+1 {
		clipped = str[:length-3] + "..."
	}
	return clipped
}
