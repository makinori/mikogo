package cmdmenu

import (
	"strings"
)

type Menu struct {
	Name     string
	Commands []Runnable
	// can be nil, will take from parent
	PrintUsage func(msg string)

	parent Runnable
}

func (c *Menu) getName() string {
	return c.Name
}

func (m *Menu) setOrGetParent(r Runnable) Runnable {
	if r != nil {
		m.parent = r
		return nil
	}
	return m.parent
}

func (c *Menu) getPrintUsageHandle() func(msg string) {
	return c.PrintUsage
}

func (c *Menu) Run(args []string) {
	if len(args) > 0 {
		name := strings.ToLower(args[0])
		for i := range c.Commands {
			if name == c.Commands[i].getName() {
				c.Commands[i].setOrGetParent(c)
				c.Commands[i].Run(args[1:])
				return
			}
		}
	}

	names := make([]string, len(c.Commands))
	for i := range c.Commands {
		names[i] = c.Commands[i].getName()
	}

	printUsage(c, getCallStack(c)+" <subcommand>\n  "+
		strings.Join(names, ", "),
	)
}
