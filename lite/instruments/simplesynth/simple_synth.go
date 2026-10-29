package simplesynth

import "github.com/ak47less/step1go/lite"

type SimpleSynthFactory struct {
}

// CreateInstrument implements [lite.InstrumentFactory].
func (inst *SimpleSynthFactory) CreateInstrument(ctx *lite.Context, info *lite.InstrumentInfo) lite.Instrument {
	panic("unimplemented")
}

func (inst *SimpleSynthFactory) _impl() lite.InstrumentFactory {
	return inst
}
