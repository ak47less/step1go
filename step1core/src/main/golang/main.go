package main

import (
	"github.com/ak47less/step1go/step1core"
	"github.com/ak47less/step1go/step1core/src/boot"
)

func main() {

	logger := step1core.Log()
	logger.SetGate(step1core.LogLevelTrace)

	err := boot.Run()
	if err != nil {
		logger.Error("%s", err.Error())
	}
}
