package cmdmenu

import (
	"unsafe"
)

type Runnable interface {
	getName() string
	Run(
		args []string, userValue unsafe.Pointer,
		printUsage func(msg string),
		parents ...Runnable,
	)
}

func getCallStack(name string, parents []Runnable) (names string) {
	for i := range parents {
		names += parents[i].getName() + " "
	}
	names += name
	return
}
