package cmdmenu

type Command struct {
	Name   string
	Args   int
	Usage  string
	Handle func(args []string)
	// can be nil, will take from parent
	PrintUsage func(msg string)

	parent Runnable
}

func (c *Command) getName() string {
	return c.Name
}

func (c *Command) setOrGetParent(r Runnable) Runnable {
	if r != nil {
		c.parent = r
		return nil
	}
	return c.parent
}

func (c *Command) getPrintUsageHandle() func(msg string) {
	return c.PrintUsage
}

func (c *Command) Run(args []string) {
	if c.Args == 0 || len(args) >= c.Args {
		c.Handle(args)
		return
	}
	printUsage(c, getCallStack(c)+" "+c.Usage)
}
