package cmdmenu

import (
	"unsafe"
)

type Runnable interface {
	getName() string
	getPrintUsage() func(msg string)
	getUserValue() unsafe.Pointer
	setOrGetParent(set Runnable) Runnable
	// TODO: place parent and user value here instead
	Run(args []string)
}

func getUserValue[T any](r Runnable) *T {
	value := r.getUserValue()
	if value != nil {
		return (*T)(value)
	}

	for value == nil {
		r = r.setOrGetParent(nil)
		if r == nil {
			break
		}

		value = r.getUserValue()
		if value != nil {
			return (*T)(value)
		}
	}

	return nil
}

func printUsage(r Runnable, usage string) {
	handle := r.getPrintUsage()
	if handle != nil {
		handle(usage)
		return
	}

	for handle == nil {
		r = r.setOrGetParent(nil)
		if r == nil {
			break
		}

		handle = r.getPrintUsage()
		if handle != nil {
			handle(usage)
			return
		}
	}

	panic("failed to get print handle")
}

func getCallStack(r Runnable) (names string) {
	for r != nil {
		if names == "" {
			names = r.getName()
		} else {
			names = r.getName() + " " + names
		}
		r = r.setOrGetParent(nil)
	}
	return
}
