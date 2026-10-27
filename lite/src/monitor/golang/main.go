package main

import (
	"github.com/ak47less/step1go/lite"
)

func main() {

	var err error

	// fd := int(os.Stdin.Fd())
	// oldState, err := term.MakeRaw(fd)
	// if err != nil {
	// 	panic(err)
	// }
	// defer term.Restore(fd, oldState) // 一定要恢复！

	///////////////////

	cfg := lite.DefaultConfig()
	cfg.Monitor = true
	err = lite.Run(cfg)
	if err != nil {
		panic(err)
	}
}
