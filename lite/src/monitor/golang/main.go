package main

import (
	"github.com/ak47less/step1go/lite"
	"github.com/ak47less/step1go/lite/instruments/simplesynth"
)

func main() {

	var err error
	cfg := lite.DefaultConfig()
	cfg.Monitor = &lite.MonitorConfig{
		URL: "tcp://localhost:33309/",
	}

	ssfactory := new(simplesynth.SimpleSynthFactory)
	insfr := new(lite.InstrumentFactoryRegistry)
	insfr.Register(ssfactory)
	insfr.SetDefault(ssfactory)

	cfg.InstrumentFactory = insfr

	err = lite.Run(cfg)
	if err != nil {
		panic(err)
	}
}
