package cmdmenu

type Runnable[T any] interface {
	getName() string
	Run(
		args []string, userValue *T,
		printUsage func(msg string),
		parents ...Runnable[T],
	)
}

func getCallStack[T any](name string, parents []Runnable[T]) (names string) {
	for i := range parents {
		names += parents[i].getName() + " "
	}
	names += name
	return
}
