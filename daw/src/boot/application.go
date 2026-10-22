package boot

func Run() error {
	runner := new(Runner)
	return runner.Run()
}
