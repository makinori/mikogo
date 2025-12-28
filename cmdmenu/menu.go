package cmdmenu

import (
	"strings"
	"unsafe"
)

type Menu[T any] struct {
	Name     string
	Commands []Runnable
	// can be nil, will take from parent
	UserValue *T
	// can be nil, will take from parent
	PrintUsage func(msg string)

	currentParent Runnable
}

func (m *Menu[T]) getName() string {
	return m.Name
}

func (m *Menu[T]) getPrintUsage() func(msg string) {
	return m.PrintUsage
}

func (m *Menu[T]) getUserValue() unsafe.Pointer {
	return unsafe.Pointer(m.UserValue)
}

func (m *Menu[T]) setOrGetParent(r Runnable) Runnable {
	if r != nil {
		m.currentParent = r
		return nil
	}
	return m.currentParent
}

func (m *Menu[T]) Run(args []string) {
	if len(args) > 0 {
		name := strings.ToLower(args[0])
		for i := range m.Commands {
			if name == m.Commands[i].getName() {
				m.Commands[i].setOrGetParent(m)
				m.Commands[i].Run(args[1:])
				return
			}
		}
	}

	names := make([]string, len(m.Commands))
	for i := range m.Commands {
		names[i] = m.Commands[i].getName()
	}

	printUsage(m, getCallStack(m)+" <subcommand>\n  "+
		strings.Join(names, ", "),
	)
}
