// This example illustrates how C++ classes can be used from Go using SWIG.

package main

import (
	"fmt"

	"example"
)

func main() {
	// ----- Object creation -----

	// Exception test
	l := example.NewLine(5)
	fmt.Printf("area of line: %f\n", l.Area())
}
