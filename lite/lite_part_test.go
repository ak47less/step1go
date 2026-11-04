package lite

import (
	"errors"
	"reflect"
	"testing"
)

type partTestPart struct {
	info *PartInfo
}

func (inst *partTestPart) GetInfo() *PartInfo {
	return inst.info
}

type partTestFactory struct {
	registration *PartRegistration
	create       func(*PartInfo) (Part, error)
	calls        int
}

func (inst *partTestFactory) CreatePart(info *PartInfo) (Part, error) {
	inst.calls++
	if inst.create != nil {
		return inst.create(info)
	}
	return &partTestPart{info: info}, nil
}

func (inst *partTestFactory) GetRegistration() *PartRegistration {
	return inst.registration
}

func newPartTestRegistration(id PartID, name PartName, classes ...PartClass) (*PartRegistration, *partTestFactory) {
	factory := new(partTestFactory)
	registration := &PartRegistration{PartMeta: PartMeta{
		ID:      id,
		Name:    name,
		Classes: classes,
		Factory: factory,
	}}
	factory.registration = registration
	return registration, factory
}

func TestNewPartContext(t *testing.T) {
	ec := new(EngineContext)
	pc := NewPartContext(ec)

	if pc == nil {
		t.Fatal("NewPartContext() returned nil")
	}
	if pc.EC != ec {
		t.Errorf("EC = %p, want %p", pc.EC, ec)
	}
	if pc.Registry == nil {
		t.Fatal("Registry is nil")
	}
	if pc.Manager == nil {
		t.Fatal("Manager is nil")
	}

	if names := pc.Registry.ListNames(); len(names) != 0 {
		t.Errorf("Registry.ListNames() = %v, want empty", names)
	}
	if names := pc.Manager.ListNames(); len(names) != 0 {
		t.Errorf("Manager.ListNames() = %v, want empty", names)
	}
	if _, err := pc.Registry.FindByID(1); err == nil {
		t.Error("Registry.FindByID() returned nil error for an unknown part")
	}
	if _, err := pc.Manager.FindPartByName("unknown"); err == nil {
		t.Error("Manager.FindPartByName() returned nil error for an unknown part")
	}
}

func TestPartRegistryRegisterAndFind(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	beta, _ := newPartTestRegistration(2, "beta", "ui", "audio")
	alpha, _ := newPartTestRegistration(1, "alpha", "audio")

	if err := pc.Registry.Register(beta); err != nil {
		t.Fatalf("Register(beta) returned error: %v", err)
	}
	if err := pc.Registry.Register(alpha); err != nil {
		t.Fatalf("Register(alpha) returned error: %v", err)
	}

	byID, err := pc.Registry.FindByID(1)
	if err != nil {
		t.Fatalf("FindByID() returned error: %v", err)
	}
	if byID != alpha {
		t.Error("FindByID() returned a different registration")
	}

	byName, err := pc.Registry.FindByName("beta")
	if err != nil {
		t.Fatalf("FindByName() returned error: %v", err)
	}
	if byName != beta {
		t.Error("FindByName() returned a different registration")
	}
	if alpha.Context != pc {
		t.Errorf("registration context = %p, want %p", alpha.Context, pc)
	}

	names := pc.Registry.ListNames()
	wantNames := []PartName{"alpha", "beta"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("ListNames() = %v, want %v", names, wantNames)
	}

	audio, err := pc.Registry.ListWithClass("audio")
	if err != nil {
		t.Fatalf("ListWithClass() returned error: %v", err)
	}
	if want := []*PartRegistration{alpha, beta}; !reflect.DeepEqual(audio, want) {
		t.Errorf("ListWithClass() = %v, want %v", audio, want)
	}

	ui, err := pc.Registry.ListWithClass("ui")
	if err != nil {
		t.Fatalf("ListWithClass() returned error: %v", err)
	}
	if want := []*PartRegistration{beta}; !reflect.DeepEqual(ui, want) {
		t.Errorf("ListWithClass() = %v, want %v", ui, want)
	}
}

func TestPartRegistryRejectsInvalidAndDuplicateRegistrations(t *testing.T) {
	pc := NewPartContext(new(EngineContext))

	if err := pc.Registry.Register(nil); err == nil {
		t.Error("Register(nil) returned nil error")
	}
	if err := pc.Registry.Register(&PartRegistration{}); err == nil {
		t.Error("Register() accepted an empty name")
	}
	if err := pc.Registry.Register(&PartRegistration{PartMeta: PartMeta{Name: "no-factory"}}); err == nil {
		t.Error("Register() accepted a nil factory")
	}

	first, _ := newPartTestRegistration(1, "first")
	if err := pc.Registry.Register(first); err != nil {
		t.Fatalf("Register(first) returned error: %v", err)
	}

	duplicateID, _ := newPartTestRegistration(1, "other")
	if err := pc.Registry.Register(duplicateID); err == nil {
		t.Error("Register() accepted a duplicate ID")
	}

	duplicateName, _ := newPartTestRegistration(2, "first")
	if err := pc.Registry.Register(duplicateName); err == nil {
		t.Error("Register() accepted a duplicate name")
	}
}

func TestPartRegistryRegisterFactory(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	registration, factory := newPartTestRegistration(1, "factory-part", "audio")

	if err := pc.Registry.RegisterFactory(factory); err != nil {
		t.Fatalf("RegisterFactory() returned error: %v", err)
	}

	found, err := pc.Registry.FindByID(1)
	if err != nil {
		t.Fatalf("FindByID() returned error: %v", err)
	}
	if found != registration {
		t.Error("RegisterFactory() registered a different registration")
	}
	if found.Factory != factory {
		t.Error("RegisterFactory() did not bind the factory")
	}

	if err := pc.Registry.RegisterFactory(nil); err == nil {
		t.Error("RegisterFactory(nil) returned nil error")
	}
}

func TestPartRegistryRegisterPart(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	info := &PartInfo{PartMeta: PartMeta{
		ID:      7,
		Name:    "ready-part",
		Classes: []PartClass{"ui"},
	}}
	part := &partTestPart{info: info}

	if err := pc.Registry.RegisterPart(part); err != nil {
		t.Fatalf("RegisterPart() returned error: %v", err)
	}

	found, err := pc.Registry.FindByName("ready-part")
	if err != nil {
		t.Fatalf("FindByName() returned error: %v", err)
	}
	if found.Part != part {
		t.Error("RegisterPart() did not bind the part")
	}
	if found.Context != pc {
		t.Errorf("registration context = %p, want %p", found.Context, pc)
	}

	got, err := pc.Manager.FindPartByID(7)
	if err != nil {
		t.Fatalf("FindPartByID() returned error: %v", err)
	}
	if got != part {
		t.Error("Manager did not return the registered part")
	}

	if err := pc.Registry.RegisterPart(nil); err == nil {
		t.Error("RegisterPart(nil) returned nil error")
	}
	if err := pc.Registry.RegisterPart(&partTestPart{}); err == nil {
		t.Error("RegisterPart() accepted a part with nil info")
	}
}

func TestPartManagerCreatesAndCachesParts(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	beta, betaFactory := newPartTestRegistration(2, "beta", "audio")
	alpha, alphaFactory := newPartTestRegistration(1, "alpha", "audio")

	for _, registration := range []*PartRegistration{beta, alpha} {
		if err := pc.Registry.Register(registration); err != nil {
			t.Fatalf("Register(%q) returned error: %v", registration.Name, err)
		}
	}

	alphaPart, err := pc.Manager.FindPartByID(1)
	if err != nil {
		t.Fatalf("FindPartByID() returned error: %v", err)
	}
	if alphaPart == nil {
		t.Fatal("FindPartByID() returned nil")
	}
	if alphaFactory.calls != 1 {
		t.Errorf("alpha factory calls = %d, want 1", alphaFactory.calls)
	}
	if alpha.Part != alphaPart {
		t.Error("created part was not stored in its registration")
	}
	if alphaPart.GetInfo() == nil || alphaPart.GetInfo().Context != pc {
		t.Error("created part did not receive the part context")
	}

	again, err := pc.Manager.FindPartByName("alpha")
	if err != nil {
		t.Fatalf("FindPartByName() returned error: %v", err)
	}
	if again != alphaPart {
		t.Error("FindPartByName() did not return the cached part")
	}
	if alphaFactory.calls != 1 {
		t.Errorf("alpha factory calls after cached lookup = %d, want 1", alphaFactory.calls)
	}

	betaPart, err := pc.Manager.FindPartByID(2)
	if err != nil {
		t.Fatalf("FindPartByID(beta) returned error: %v", err)
	}
	if betaFactory.calls != 1 {
		t.Errorf("beta factory calls = %d, want 1", betaFactory.calls)
	}

	parts, err := pc.Manager.ListPartsWithClass("audio")
	if err != nil {
		t.Fatalf("ListPartsWithClass() returned error: %v", err)
	}
	if want := []Part{alphaPart, betaPart}; !reflect.DeepEqual(parts, want) {
		t.Errorf("ListPartsWithClass() = %v, want %v", parts, want)
	}

	names := pc.Manager.ListNames()
	wantNames := []PartName{"alpha", "beta"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("ListNames() = %v, want %v", names, wantNames)
	}
}

func TestPartManagerUsesRegisteredPart(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	registration, _ := newPartTestRegistration(1, "ready")
	existing := &partTestPart{info: &PartInfo{PartMeta: registration.PartMeta}}
	registration.Factory = nil
	registration.Part = existing

	if err := pc.Registry.Register(registration); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	part, err := pc.Manager.FindPartByID(1)
	if err != nil {
		t.Fatalf("FindPartByID() returned error: %v", err)
	}
	if part != existing {
		t.Error("FindPartByID() did not return the registered part")
	}
}

func TestPartManagerPropagatesFactoryErrors(t *testing.T) {
	pc := NewPartContext(new(EngineContext))
	registration, factory := newPartTestRegistration(1, "broken")
	wantErr := errors.New("factory failed")
	factory.create = func(*PartInfo) (Part, error) {
		return nil, wantErr
	}
	if err := pc.Registry.Register(registration); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	if _, err := pc.Manager.FindPartByID(1); !errors.Is(err, wantErr) {
		t.Errorf("FindPartByID() error = %v, want %v", err, wantErr)
	}

	factory.create = func(*PartInfo) (Part, error) {
		return nil, nil
	}
	if _, err := pc.Manager.FindPartByID(1); err == nil {
		t.Error("FindPartByID() accepted a nil part from the factory")
	}
}
