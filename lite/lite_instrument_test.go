package lite

import (
	"reflect"
	"testing"
)

type instrumentTestInstrument struct {
	info *InstrumentInfo
}

func (inst *instrumentTestInstrument) Dispatch(*MidiMessage) error {
	return nil
}

func (inst *instrumentTestInstrument) GetInfo() *InstrumentInfo {
	return inst.info
}

type instrumentTestFactory struct {
	infos      []*InstrumentInfo
	createErr  error
	createFunc func(*Context, *InstrumentInfo) (Instrument, error)
}

func (inst *instrumentTestFactory) ListInfo() []*InstrumentInfo {
	return inst.infos
}

func (inst *instrumentTestFactory) CreateInstrument(
	ctx *Context,
	info *InstrumentInfo,
) (Instrument, error) {
	if inst.createFunc != nil {
		return inst.createFunc(ctx, info)
	}
	if inst.createErr != nil {
		return nil, inst.createErr
	}
	return &instrumentTestInstrument{info: info}, nil
}

type instrumentTestProviderPart struct {
	info          *PartInfo
	registrations []*InstrumentRegistration
}

func (inst *instrumentTestProviderPart) GetInfo() *PartInfo {
	return inst.info
}

func (inst *instrumentTestProviderPart) ListRegistrations() []*InstrumentRegistration {
	return inst.registrations
}

type instrumentTestPlainPart struct {
	info *PartInfo
}

func (inst *instrumentTestPlainPart) GetInfo() *PartInfo {
	return inst.info
}

func newInstrumentTestPartInfo(name string) *PartInfo {
	return &PartInfo{PartMeta: PartMeta{Name: PartName(name)}}
}

func TestInstrumentRegistryRegisterInstrument(t *testing.T) {
	firstFactory := new(instrumentTestFactory)
	secondFactory := new(instrumentTestFactory)
	provider := &instrumentTestProviderPart{
		info: newInstrumentTestPartInfo("provider"),
		registrations: []*InstrumentRegistration{
			{
				InstrumentInfo: InstrumentInfo{Name: "piano", Label: "Piano"},
				Factory:        firstFactory,
			},
			{
				InstrumentInfo: InstrumentInfo{Name: "organ", Label: "Organ"},
				Factory:        secondFactory,
			},
		},
	}
	table := make(map[string]*InstrumentRegistration)
	registry := new(InstrumentRegistry)

	if err := registry.registerInstrument(provider, table); err != nil {
		t.Fatalf("registerInstrument() returned error: %v", err)
	}
	if len(table) != 2 {
		t.Fatalf("registration count = %d, want 2", len(table))
	}
	if table["piano"].Factory != firstFactory {
		t.Error("piano was registered with the wrong factory")
	}
	if table["organ"].Factory != secondFactory {
		t.Error("organ was registered with the wrong factory")
	}

	plain := &instrumentTestPlainPart{info: newInstrumentTestPartInfo("plain")}
	if err := registry.registerInstrument(plain, table); err != nil {
		t.Errorf("registerInstrument() returned error for a non-provider: %v", err)
	}
	if len(table) != 2 {
		t.Errorf("registration count after non-provider = %d, want 2", len(table))
	}
}

func TestInstrumentRegistryRegisterInstrumentRejectsInvalidInput(t *testing.T) {
	registry := new(InstrumentRegistry)
	if err := registry.registerInstrument(nil, make(map[string]*InstrumentRegistration)); err == nil {
		t.Error("registerInstrument(nil) returned nil error")
	}

	provider := &instrumentTestProviderPart{
		info: newInstrumentTestPartInfo("provider"),
		registrations: []*InstrumentRegistration{
			nil,
		},
	}
	if err := registry.registerInstrument(provider, make(map[string]*InstrumentRegistration)); err == nil {
		t.Error("registerInstrument() accepted a nil registration")
	}

	provider.registrations = []*InstrumentRegistration{{
		InstrumentInfo: InstrumentInfo{Name: "piano"},
	}}
	if err := registry.registerInstrument(provider, make(map[string]*InstrumentRegistration)); err == nil {
		t.Error("registerInstrument() accepted a registration without a factory")
	}

	provider.registrations = []*InstrumentRegistration{{
		Factory: new(instrumentTestFactory),
	}}
	if err := registry.registerInstrument(provider, make(map[string]*InstrumentRegistration)); err == nil {
		t.Error("registerInstrument() accepted an empty instrument name")
	}

	provider.registrations = []*InstrumentRegistration{{
		InstrumentInfo: InstrumentInfo{Name: "piano"},
		Factory:        new(instrumentTestFactory),
	}}
	if err := registry.registerInstrument(provider, nil); err == nil {
		t.Error("registerInstrument() accepted a nil table")
	}
}

func TestInstrumentRegistryRegisterAllInstruments(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	factory := new(instrumentTestFactory)
	provider := &instrumentTestProviderPart{
		info: newInstrumentTestPartInfo("provider"),
		registrations: []*InstrumentRegistration{{
			InstrumentInfo: InstrumentInfo{Name: "piano", Label: "Piano"},
			Factory:        factory,
		}},
	}
	plain := &instrumentTestPlainPart{info: newInstrumentTestPartInfo("plain")}

	if err := pc.Registry.RegisterPart(provider); err != nil {
		t.Fatalf("RegisterPart(provider) returned error: %v", err)
	}
	if err := pc.Registry.RegisterPart(plain); err != nil {
		t.Fatalf("RegisterPart(plain) returned error: %v", err)
	}

	registry := new(InstrumentRegistry)
	ctx := &Context{PartManager: pc.Manager}
	if err := registry.registerAllInstruments(ctx); err != nil {
		t.Fatalf("registerAllInstruments() returned error: %v", err)
	}
	registration := registry.table["piano"]
	if registration == nil {
		t.Fatal("piano was not registered")
	}
	if registration.Factory != factory {
		t.Error("piano was registered with the wrong factory")
	}
}

func TestInstrumentRegistryRegisterAllInstrumentsRejectsInvalidContext(t *testing.T) {
	registry := new(InstrumentRegistry)
	if err := registry.registerAllInstruments(nil); err == nil {
		t.Error("registerAllInstruments(nil) returned nil error")
	}
	if err := registry.registerAllInstruments(new(Context)); err == nil {
		t.Error("registerAllInstruments() accepted a nil part manager")
	}
	if err := (*InstrumentRegistry)(nil).registerAllInstruments(new(Context)); err == nil {
		t.Error("registerAllInstruments() on a nil registry returned nil error")
	}
}

func TestInnerInstrumentManagerGetFactory(t *testing.T) {
	pianoFactory := new(instrumentTestFactory)
	organFactory := new(instrumentTestFactory)
	manager := &innerInstrumentManagerImpl{
		ir: InstrumentRegistry{table: map[string]*InstrumentRegistration{
			"piano": {
				InstrumentInfo: InstrumentInfo{Name: "piano"},
				Factory:        pianoFactory,
			},
			"organ": {
				InstrumentInfo: InstrumentInfo{Name: "organ"},
				Factory:        organFactory,
			},
			"broken": {
				InstrumentInfo: InstrumentInfo{Name: "broken"},
			},
		}},
	}

	factory, err := manager.GetFactory(&InstrumentInfo{Name: "piano"})
	if err != nil {
		t.Fatalf("GetFactory(piano) returned error: %v", err)
	}
	if factory != pianoFactory {
		t.Error("GetFactory(piano) returned the wrong factory")
	}

	if _, err := manager.GetFactory(&InstrumentInfo{Name: "unknown"}); err == nil {
		t.Error("GetFactory(unknown) returned nil error")
	}
	if _, err := manager.GetFactory(&InstrumentInfo{Name: "broken"}); err == nil {
		t.Error("GetFactory(broken) returned nil error")
	}
	if _, err := manager.GetFactory(nil); err == nil {
		t.Error("GetFactory(nil) returned nil error")
	}
	if _, err := (*innerInstrumentManagerImpl)(nil).GetFactory(&InstrumentInfo{Name: "piano"}); err == nil {
		t.Error("GetFactory() on a nil manager returned nil error")
	}
}

func TestInnerInstrumentManagerListInfo(t *testing.T) {
	manager := &innerInstrumentManagerImpl{
		ir: InstrumentRegistry{table: map[string]*InstrumentRegistration{
			"beta": {
				InstrumentInfo: InstrumentInfo{Name: "beta", Label: "Beta"},
				Factory:        new(instrumentTestFactory),
			},
			"alpha-2": {
				InstrumentInfo: InstrumentInfo{Name: "alpha", Label: "Alpha 2"},
				Factory:        new(instrumentTestFactory),
			},
			"alpha-1": {
				InstrumentInfo: InstrumentInfo{Name: "alpha", Label: "Alpha 1"},
				Factory:        new(instrumentTestFactory),
			},
		}},
	}

	list := manager.ListInfo()
	want := []*InstrumentInfo{
		{Name: "alpha", Label: "Alpha 1"},
		{Name: "alpha", Label: "Alpha 2"},
		{Name: "beta", Label: "Beta"},
	}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("ListInfo() = %v, want %v", list, want)
	}

	list[0].Label = "changed"
	if manager.ir.table["alpha-1"].Label != "Alpha 1" {
		t.Error("ListInfo() returned data that aliases the registry")
	}

	if got := (*innerInstrumentManagerImpl)(nil).ListInfo(); got != nil {
		t.Errorf("ListInfo() on a nil manager = %v, want nil", got)
	}
}

func TestInstrumentRegistryOnInitRegistersAvailableFactories(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	providedFactory := new(instrumentTestFactory)
	provider := &instrumentTestProviderPart{
		info: newInstrumentTestPartInfo("provider"),
		registrations: []*InstrumentRegistration{{
			InstrumentInfo: InstrumentInfo{Name: "piano"},
			Factory:        providedFactory,
		}},
	}
	if err := pc.Registry.RegisterPart(provider); err != nil {
		t.Fatalf("RegisterPart() returned error: %v", err)
	}

	manager := NewInstrumentManager()
	life := manager.(Life)
	lifeCycle := life.GetLifeCycle()
	if err := lifeCycle.OnInit(&Context{PartManager: pc.Manager}); err != nil {
		t.Fatalf("OnInit() returned error: %v", err)
	}

	factory, err := manager.GetFactory(&InstrumentInfo{Name: "piano"})
	if err != nil {
		t.Fatalf("GetFactory(piano) returned error: %v", err)
	}
	if factory != providedFactory {
		t.Error("GetFactory(piano) returned the wrong factory")
	}
}
