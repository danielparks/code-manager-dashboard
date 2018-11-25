package main

import "fmt"

// Yes, No, Maybe
type Trinary int

const (
	No    Trinary = iota
	Maybe Trinary = iota
	Yes   Trinary = iota
)

func BoolToTrinary(v bool) Trinary {
	if v {
		return Yes
	} else {
		return No
	}
}

func (trinary Trinary) String() string {
	if trinary == No {
		return "No"
	} else if trinary == Yes {
		return "Yes"
	} else if trinary == Maybe {
		return "Maybe"
	} else {
		panic(fmt.Sprintf("Invalid Trinary: %d", trinary))
	}
}
