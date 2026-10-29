package main

import (
	"github.com/ak47less/step1go/lite"
)

func main() {

	var err error
	cfg := lite.DefaultConfig()
	cfg.Monitor = &lite.MonitorConfig{
		URL: "tcp://localhost:33309/",
	}

	cfg.InstrumentFactory = nil

	err = lite.Run(cfg)
	if err != nil {
		panic(err)
	}
}
