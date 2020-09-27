package main

import (
	"fmt"
)

type TestReturn struct {
}

type TestArg struct {
}

type TestReceiver struct {
}

// TODO: check pointer
func (t TestReceiver) exception(ta TestArg) bool {
	fmt.Println("next: exception")
	i := 5
	if i == 5 {
		panic("titanic")
	}
	fmt.Println("prev: exception")

	return true
}

type ReceiverInterface interface {
	exception(ta TestArg) bool
}

func main() {
	fmt.Println("main")
	var tr ReceiverInterface
	tr = TestReceiver{}
	ta := TestArg{}
	tr.exception(ta)
}
