package cmdmenu

import (
	"strings"
	"unsafe"
)

type Menu[T any] struct {
	Name     string
	Commands []Runnable
}

func (m *Menu[T]) getName() string {
	return m.Name
}

func (m *Menu[T]) Run(
	args []string, userValue unsafe.Pointer,
	printUsage func(msg string),
	parents ...Runnable,
) {
	if len(args) > 0 {
		name := strings.ToLower(args[0])
		for i := range m.Commands {
			if name == m.Commands[i].getName() {
				m.Commands[i].Run(
					args[1:], userValue, printUsage, append(parents, m)...,
				)
				return
			}
		}
	}

	names := make([]string, len(m.Commands))
	for i := range m.Commands {
		names[i] = m.Commands[i].getName()
	}

	printUsage(
		getCallStack(m.Name, parents) + " <subcommand>\n  " +
			strings.Join(names, ", "),
	)
}
