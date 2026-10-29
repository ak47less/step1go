package lite

type Instrument interface {
	Dispatch(msg *MidiMessage) error

	GetInfo() *InstrumentInfo
}

type InstrumentFactory interface {
	ListInfo() []*InstrumentInfo

	CreateInstrument(ctx *Context, info *InstrumentInfo) Instrument
}

type InstrumentInfo struct {
	Name   string
	Label  string
	Preset string
}

////////////////////////////////////////////////////////////////////////////////

type InstrumentRegistration struct {
	InstrumentInfo

	Factory InstrumentFactory
}

////////////////////////////////////////////////////////////////////////////////

type InstrumentFactoryRegistry struct {
	table map[string]*InstrumentRegistration // map[name]ir
}

// CreateInstrument implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) CreateInstrument(ctx *Context, info *InstrumentInfo) Instrument {

	//todo

	panic("unimplemented")
}

// ListInfo implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) ListInfo() []*InstrumentInfo {
	return make([]*InstrumentInfo, 0)
}

func (inst *InstrumentFactoryRegistry) Register(f InstrumentFactory) {

	//todo

}

func (inst *InstrumentFactoryRegistry) add(ir *InstrumentRegistration) {

	//todo

}

func (inst *InstrumentFactoryRegistry) init() {

	//todo

}

func (inst *InstrumentFactoryRegistry) _impl() InstrumentFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
