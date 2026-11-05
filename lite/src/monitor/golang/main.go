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

	// parts

	lcsl := new(lite.LifeCycleStageLogger)
	aec := new(lite.AutoEngineController)

	cfg.Parts = append(cfg.Parts, insfr)
	cfg.Parts = append(cfg.Parts, lcsl)
	cfg.Parts = append(cfg.Parts, aec)

	// run

	err = lite.Run(cfg)
	if err != nil {
		panic(err)
	}
}
