package main

import (
	"github.com/ak47less/step1go/lite"
	"github.com/ak47less/step1go/lite/drivers/malgo4lite"
	"github.com/ak47less/step1go/lite/instruments/simplesynth"
)

func main() {

	var err error

	// config

	cfg := lite.DefaultConfig()

	// driver

	driver := new(malgo4lite.MalgoDriver)
	cfg.Driver = driver

	// monitor

	cfg.Monitor = &lite.MonitorConfig{
		URL: "tcp://localhost:33309/",
	}

	// Instrument

	ssfactory := new(simplesynth.SimpleSynthFactory)
	insfr := new(lite.InstrumentFactoryRegistry)
	insfr.Register(ssfactory)
	insfr.SetDefault(ssfactory)

	cfg.InstrumentFactory = insfr

	// run

	err = lite.Run(cfg)
	if err != nil {
		panic(err)
	}
}
