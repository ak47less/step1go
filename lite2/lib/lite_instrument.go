package lib

import "github.com/ak47less/step1go/lite2/webaudioapi"

type Instrument interface {
	Dispatch(events ...*MidiEvent) error

	GetOutput() webaudioapi.AudioNode
}

type InstrumentHolder interface {
	GetInstrument() (Instrument, error)
}

type InstrumentLoader interface {
	LoadInstrument(info *InstrumentInfo) (Instrument, error)
}

type InstrumentInfo struct {
	Name   string
	Label  string
	Preset string
}

type InstrumentContext struct {
}

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

	ListInstruments() []*InstrumentInfo
}

type InstrumentRegistry interface {
	Register(provider InstrumentProvider) error
}

////////////////////////////////////////////////////////////////////////////////

type AutoInstrumentRegistry struct {
	context *ApplicationContext
}

// GetLifeCycle implements [Life].
func (inst *AutoInstrumentRegistry) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit:   inst.innerOnInit,
		OnCreate: inst.innerOnCreate,
	}
	return lc
}

func (inst *AutoInstrumentRegistry) innerOnInit(ac *ApplicationContext) error {
	inst.context = ac
	return nil
}

func (inst *AutoInstrumentRegistry) innerOnCreate() error {
	return inst.innerScanForInstruments()
}

func (inst *AutoInstrumentRegistry) innerScanForInstruments() error {

	src := inst.context.Parts
	namelist := src.ListNames()

	for _, name := range namelist {
		info, err := src.GetInfoByName(name)
		if err != nil {
			return err
		}
		part := info.Part
		provider, ok := part.(InstrumentProvider)
		if ok && provider != nil {
			err = inst.Register(provider)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Register implements [InstrumentRegistry].
func (inst *AutoInstrumentRegistry) Register(ip InstrumentProvider) error {
	panic("unimplemented")
}

func (inst *AutoInstrumentRegistry) _impl() (InstrumentRegistry, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
