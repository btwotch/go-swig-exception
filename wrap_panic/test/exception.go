package main

import (
	"fmt"
)

func exception() bool {
	fmt.Println("next: exception")
	i := 5
	if i == 5 {
		panic("titanic")
	}
	fmt.Println("prev: exception")

	return true
}

func main() {
	fmt.Println("main")
	exception()
}

type simplestruct struct {

}

func (s simplestruct) foo() bool {
	fmt.Println("foo")

	return false
}
