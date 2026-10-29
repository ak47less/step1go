package lite

import (
	"fmt"
	"sort"
	"sync"
)

type Instrument interface {
	Dispatch(msg *MidiMessage) error

	GetInfo() *InstrumentInfo
}

type InstrumentFactory interface {
	ListInfo() []*InstrumentInfo

	CreateInstrument(ctx *Context, info *InstrumentInfo) (Instrument, error)
}

type InstrumentHolder interface {
	GetInstrument(ctx *Context) (Instrument, error)
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
	table          map[string]*InstrumentRegistration // map[name]ir
	factoryDefault InstrumentFactory
	mu             sync.RWMutex
}

// CreateInstrument implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) CreateInstrument(
	ctx *Context,
	info *InstrumentInfo,
) (Instrument, error) {
	if inst == nil {
		return nil, fmt.Errorf("instrument factory registry is nil")
	}
	if info == nil {
		return nil, fmt.Errorf("instrument info is nil")
	}

	inst.mu.RLock()
	ir := inst.table[info.Name]
	factory := inst.factoryDefault
	if ir != nil && ir.Factory != nil {
		factory = ir.Factory
	}
	inst.mu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("instrument %q is not registered and no default factory is set", info.Name)
	}

	return factory.CreateInstrument(ctx, info)
}

func (inst *InstrumentFactoryRegistry) SetDefault(f InstrumentFactory) {
	if inst == nil || f == nil {
		return
	}
	inst.mu.Lock()
	inst.factoryDefault = f
	inst.mu.Unlock()
}

// ListInfo implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) ListInfo() []*InstrumentInfo {
	if inst == nil {
		return nil
	}

	inst.mu.RLock()
	list := make([]*InstrumentInfo, 0, len(inst.table))
	for _, ir := range inst.table {
		if ir == nil {
			continue
		}
		item := ir.InstrumentInfo
		list = append(list, &item)
	}
	inst.mu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		if list[i].Label != list[j].Label {
			return list[i].Label < list[j].Label
		}
		return list[i].Preset < list[j].Preset
	})
	return list
}

func (inst *InstrumentFactoryRegistry) Register(f InstrumentFactory) {
	if inst == nil || f == nil {
		return
	}
	for _, info := range f.ListInfo() {
		if info == nil {
			continue
		}
		inst.add(&InstrumentRegistration{
			InstrumentInfo: *info,
			Factory:        f,
		})
	}
}

func (inst *InstrumentFactoryRegistry) add(ir *InstrumentRegistration) {
	if inst == nil || ir == nil || ir.Factory == nil || ir.Name == "" {
		return
	}

	inst.mu.Lock()
	inst.initLocked()
	inst.table[ir.Name] = ir
	inst.mu.Unlock()
}

func (inst *InstrumentFactoryRegistry) init() {
	if inst == nil {
		return
	}

	inst.mu.Lock()
	inst.initLocked()
	inst.mu.Unlock()
}

func (inst *InstrumentFactoryRegistry) initLocked() {
	if inst.table == nil {
		inst.table = make(map[string]*InstrumentRegistration)
	}
}

func (inst *InstrumentFactoryRegistry) _impl() InstrumentFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type DefaultInstrumentHolder struct {
}

// GetInstrument implements [InstrumentHolder].
func (inst *DefaultInstrumentHolder) GetInstrument(ctx *Context) (Instrument, error) {

	if ctx == nil {
		return nil, fmt.Errorf("lite.Context is nil")
	}

	i := ctx.CurrentInstrument
	if i != nil {
		return i, nil
	}

	factory := ctx.MainInstrumentFactory
	info := new(InstrumentInfo)

	if factory == nil {
		return nil, fmt.Errorf("context.MainInstrumentFactory is nil")
	}

	i, err := factory.CreateInstrument(ctx, info)
	if err != nil {
		return nil, err
	}

	ctx.CurrentInstrument = i
	return i, nil
}

func (inst *DefaultInstrumentHolder) _impl() InstrumentHolder {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
