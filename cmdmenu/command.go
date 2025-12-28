package cmdmenu

import (
	"unsafe"
)

type Command[T any] struct {
	Name   string
	Args   int
	Usage  string
	Handle func(user *T, args []string)
}

func (c *Command[T]) getName() string {
	return c.Name
}

func (c *Command[T]) Run(
	args []string, userValue unsafe.Pointer,
	printUsage func(msg string),
	parents ...Runnable,
) {
	if c.Args == 0 || len(args) >= c.Args {
		c.Handle((*T)(userValue), args)
		return
	}
	printUsage(
		getCallStack(c.Name, parents) + " " + c.Usage,
	)
}
