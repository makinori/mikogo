package cmdmenu

import (
	"unsafe"
)

type Command[T any] struct {
	Name   string
	Args   int
	Usage  string
	Handle func(user *T, args []string)
	// can be nil, will take from parent
	UserValue *T
	// can be nil, will take from parent
	PrintUsage func(msg string)

	currentParent Runnable
}

func (c *Command[T]) getName() string {
	return c.Name
}

func (c *Command[T]) getPrintUsage() func(msg string) {
	return c.PrintUsage
}

func (c *Command[T]) getUserValue() unsafe.Pointer {
	return unsafe.Pointer(c.UserValue)
}

func (c *Command[T]) setOrGetParent(r Runnable) Runnable {
	if r != nil {
		c.currentParent = r
		return nil
	}
	return c.currentParent
}

func (c *Command[T]) Run(args []string) {
	if c.Args == 0 || len(args) >= c.Args {
		c.Handle(getUserValue[T](c), args)
		return
	}
	printUsage(c, getCallStack(c)+" "+c.Usage)
}
