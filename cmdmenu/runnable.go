package cmdmenu

type Runnable interface {
	getName() string
	// sets if r != nil, otherwise gets
	setOrGetParent(r Runnable) Runnable
	getPrintUsageHandle() func(msg string)
	Run(args []string)
}

func printUsage(r Runnable, usage string) {
	handle := r.getPrintUsageHandle()
	if handle != nil {
		handle(usage)
		return
	}

	for handle == nil {
		parent := r.setOrGetParent(nil)
		if parent == nil {
			break
		}

		handle = parent.getPrintUsageHandle()
		if handle != nil {
			handle(usage)
			return
		}

		r = parent
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
