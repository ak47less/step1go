package main

import (
	"github.com/ak47less/step1go"
	"github.com/ak47less/step1go/src/boot"
)

func main() {
	logger := new(step1go.Logger)
	err := boot.Run()
	if err != nil {
		logger.Error("%s", err.Error())
	}
}
