package lib

type Instrument interface{}

type InstrumentInfo struct {
	Name   string
	Label  string
	Preset string
}

type InstrumentContext struct{}

type InstrumentParams struct {
	InstrumentInfo

	Context *InstrumentContext
}

type InstrumentFactory interface {
	CreateInstrument(ip *InstrumentParams) (Instrument, error)
}

type InstrumentProvider interface {
	ListRegistrations() []*InstrumentRegistration
}

type InstrumentRegistration struct {
	InstrumentInfo

	Factory InstrumentFactory

	Provider InstrumentProvider
}

type InstrumentManager interface {
	FindRegistration(info *InstrumentInfo) (*InstrumentRegistration, error)
}

type InstrumentRegistry interface {
	Register(ip *InstrumentProvider) error
}
