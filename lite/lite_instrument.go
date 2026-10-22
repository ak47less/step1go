package lite

type Instrument interface {
	Dispatch(msg *MidiMessage) error

	GetInfo() *InstrumentInfo
}

type InstrumentFactory interface {
	CreateInstrument(ctx *Context, info *InstrumentInfo) Instrument
}

type InstrumentInfo struct {
	Name   string
	Label  string
	Preset string
}
