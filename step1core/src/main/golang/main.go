package main

import (
	"github.com/ak47less/step1go/step1core"
	"github.com/ak47less/step1go/step1core/src/boot"
)

func main() {
	logger := new(step1core.Logger)
	err := boot.Run()
	if err != nil {
		logger.Error("%s", err.Error())
	}
}
