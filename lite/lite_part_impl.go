package lite

import (
	"fmt"
	"sort"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

// NewPartContext 创建并初始化零部件上下文、注册表和管理器。
func NewPartContext(ec *EngineContext) *PartContext {

	pc := new(PartContext)
	registry := &innerPartRegistry{pc: pc}
	manager := &innerPartManager{pc: pc}

	pc.EC = ec
	pc.Manager = manager
	pc.Registry = registry

	registry.init()
	manager.init()

	return pc
}

////////////////////////////////////////////////////////////////////////////////

type innerPartManager struct {
	pc *PartContext

	mu     sync.Mutex
	byID   map[PartID]Part
	byName map[PartName]Part
}

func (inst *innerPartManager) init() {
	if inst == nil {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.initLocked()
}

// FindPartByID implements [PartManager].
func (inst *innerPartManager) FindPartByID(id PartID) (Part, error) {
	registry, err := inst.innerGetRegistry()
	if err != nil {
		return nil, err
	}

	registration, err := registry.FindByID(id)
	if err != nil {
		return nil, err
	}
	return inst.innerPartFromRegistration(registration)
}

// FindPartByName implements [PartManager].
func (inst *innerPartManager) FindPartByName(name PartName) (Part, error) {
	registry, err := inst.innerGetRegistry()
	if err != nil {
		return nil, err
	}

	registration, err := registry.FindByName(name)
	if err != nil {
		return nil, err
	}
	return inst.innerPartFromRegistration(registration)
}

// ListNames implements [PartManager].
func (inst *innerPartManager) ListNames() []PartName {
	registry, err := inst.innerGetRegistry()
	if err != nil {
		return nil
	}
	return registry.ListNames()
}

// ListPartsWithClass implements [PartManager].
func (inst *innerPartManager) ListPartsWithClass(cl PartClass) ([]Part, error) {
	registry, err := inst.innerGetRegistry()
	if err != nil {
		return nil, err
	}

	registrations, err := registry.ListWithClass(cl)
	if err != nil {
		return nil, err
	}

	parts := make([]Part, 0, len(registrations))
	for _, registration := range registrations {
		part, err := inst.innerPartFromRegistration(registration)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func (inst *innerPartManager) innerGetRegistry() (PartRegistry, error) {
	if inst == nil || inst.pc == nil {
		return nil, fmt.Errorf("part manager context is nil")
	}
	if inst.pc.Registry == nil {
		return nil, fmt.Errorf("part registry is nil")
	}
	return inst.pc.Registry, nil
}

func (inst *innerPartManager) innerPartFromRegistration(registration *PartRegistration) (Part, error) {
	if inst == nil {
		return nil, fmt.Errorf("part manager is nil")
	}
	if registration == nil {
		return nil, fmt.Errorf("part registration is nil")
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.initLocked()

	if part := inst.byID[registration.ID]; part != nil {
		return part, nil
	}
	if part := inst.byName[registration.Name]; part != nil {
		return part, nil
	}
	if registration.Part != nil {
		inst.innerCachePartLocked(registration, registration.Part)
		return registration.Part, nil
	}
	if registration.Factory == nil {
		return nil, fmt.Errorf("part %q has no factory", registration.Name)
	}

	info := &PartInfo{PartMeta: registration.PartMeta}
	info.Context = inst.pc
	part, err := registration.Factory.CreatePart(info)
	if err != nil {
		return nil, fmt.Errorf("create part %q: %w", registration.Name, err)
	}
	if part == nil {
		return nil, fmt.Errorf("create part %q: factory returned nil", registration.Name)
	}

	registration.Part = part
	inst.innerCachePartLocked(registration, part)
	return part, nil
}

func (inst *innerPartManager) innerCachePartLocked(registration *PartRegistration, part Part) {
	inst.byID[registration.ID] = part
	inst.byName[registration.Name] = part
}

func (inst *innerPartManager) initLocked() {
	if inst.byID == nil {
		inst.byID = make(map[PartID]Part)
	}
	if inst.byName == nil {
		inst.byName = make(map[PartName]Part)
	}
}

func (inst *innerPartManager) _impl() PartManager {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerPartRegistry struct {
	pc *PartContext

	autoIdGenerator int

	mu     sync.RWMutex
	byID   map[PartID]*PartRegistration
	byName map[PartName]*PartRegistration
}

// RegisterFactory implements [PartRegistry].
func (inst *innerPartRegistry) RegisterFactory(factory PartFactory) error {
	if inst == nil {
		return fmt.Errorf("part registry is nil")
	}
	if factory == nil {
		return fmt.Errorf("part factory is nil")
	}

	registration := factory.GetRegistration()
	if registration == nil {
		return fmt.Errorf("part factory registration is nil")
	}
	registration.Factory = factory
	registration.Context = inst.pc
	return inst.Register(registration)
}

// RegisterPart implements [PartRegistry].
func (inst *innerPartRegistry) RegisterPart(part Part) error {
	if inst == nil {
		return fmt.Errorf("part registry is nil")
	}
	if part == nil {
		return fmt.Errorf("part is nil")
	}

	info := part.GetInfo()
	if info == nil {
		return fmt.Errorf("part info is nil")
	}

	registration := &PartRegistration{PartMeta: info.PartMeta}
	registration.Context = inst.pc
	registration.Part = part
	return inst.Register(registration)
}

func (inst *innerPartRegistry) init() {
	if inst == nil {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.initLocked()
}

// FindByID implements [PartRegistry].
func (inst *innerPartRegistry) FindByID(id PartID) (*PartRegistration, error) {
	if inst == nil {
		return nil, fmt.Errorf("part registry is nil")
	}

	inst.mu.RLock()
	registration := inst.byID[id]
	inst.mu.RUnlock()
	if registration == nil {
		return nil, fmt.Errorf("part with id %d is not registered", id)
	}
	return registration, nil
}

// FindByName implements [PartRegistry].
func (inst *innerPartRegistry) FindByName(name PartName) (*PartRegistration, error) {
	if inst == nil {
		return nil, fmt.Errorf("part registry is nil")
	}

	inst.mu.RLock()
	registration := inst.byName[name]
	inst.mu.RUnlock()
	if registration == nil {
		return nil, fmt.Errorf("part with name %q is not registered", name)
	}
	return registration, nil
}

// ListNames implements [PartRegistry].
func (inst *innerPartRegistry) ListNames() []PartName {
	if inst == nil {
		return nil
	}

	inst.mu.RLock()
	names := make([]PartName, 0, len(inst.byName))
	for name := range inst.byName {
		names = append(names, name)
	}
	inst.mu.RUnlock()

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}

// ListWithClass implements [PartRegistry].
func (inst *innerPartRegistry) ListWithClass(cl PartClass) ([]*PartRegistration, error) {
	if inst == nil {
		return nil, fmt.Errorf("part registry is nil")
	}

	inst.mu.RLock()
	registrations := make([]*PartRegistration, 0)
	for _, registration := range inst.byID {
		if registration == nil {
			continue
		}
		if registrationHasClass(registration, cl) {
			registrations = append(registrations, registration)
		}
	}
	inst.mu.RUnlock()

	sort.Slice(registrations, func(i, j int) bool {
		if registrations[i].ID != registrations[j].ID {
			return registrations[i].ID < registrations[j].ID
		}
		return registrations[i].Name < registrations[j].Name
	})
	return registrations, nil
}

func (inst *innerPartRegistry) innerMakeNextID() PartID {
	if inst.autoIdGenerator == 0 {
		inst.autoIdGenerator = 0x123456789
	}
	inst.autoIdGenerator++
	return PartID(inst.autoIdGenerator)
}

func (inst *innerPartRegistry) innerMakePartName(pr *PartRegistration) PartName {
	id := int(pr.ID)
	str := fmt.Sprintf("unnamed_part_%d", id)
	return PartName(str)
}

// Register implements [PartRegistry].
func (inst *innerPartRegistry) Register(pr *PartRegistration) error {
	if inst == nil {
		return fmt.Errorf("part registry is nil")
	}
	if pr == nil {
		return fmt.Errorf("part registration is nil")
	}
	if pr.ID == 0 {
		pr.ID = inst.innerMakeNextID()
	}
	if pr.Name == "" {
		// return fmt.Errorf("part name is empty")
		pr.Name = inst.innerMakePartName(pr)
	}
	if pr.Factory == nil && pr.Part == nil {
		return fmt.Errorf("part %q has neither a factory nor an instance", pr.Name)
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.initLocked()

	if existing := inst.byID[pr.ID]; existing != nil {
		return fmt.Errorf("part id %d is already registered as %q", pr.ID, existing.Name)
	}
	if existing := inst.byName[pr.Name]; existing != nil {
		return fmt.Errorf("part name %q is already registered with id %d", pr.Name, existing.ID)
	}
	if pr.Context == nil {
		pr.Context = inst.pc
	}

	inst.byID[pr.ID] = pr
	inst.byName[pr.Name] = pr
	return nil
}

func (inst *innerPartRegistry) initLocked() {
	if inst.byID == nil {
		inst.byID = make(map[PartID]*PartRegistration)
	}
	if inst.byName == nil {
		inst.byName = make(map[PartName]*PartRegistration)
	}
}

func registrationHasClass(registration *PartRegistration, cl PartClass) bool {
	if registration == nil {
		return false
	}
	for _, item := range registration.Classes {
		if item == cl {
			return true
		}
	}
	return false
}

func (inst *innerPartRegistry) _impl() PartRegistry {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
