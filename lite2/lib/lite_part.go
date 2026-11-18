package lib

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type PartName string

type PartClass string

type PartID int

////////////////////////////////////////////////////////////////////////////////

type PartInfo struct {
	Class PartClass
	Name  PartName
	ID    PartID
	Label string
	Part  any
}

////////////////////////////////////////////////////////////////////////////////

type PartProvider interface {
	ListParts() []*PartInfo
}

type PartRegistry interface {
	RegisterAny(o any) error

	Register(info *PartInfo) error

	RegisterProvider(p PartProvider) error
}

type PartManager interface {
	GetRegistry() PartRegistry

	GetPart(id PartID) (any, error)

	GetInfo(id PartID) (*PartInfo, error)

	GetInfoByName(name PartName) (*PartInfo, error)

	ListNames() []PartName
}

////////////////////////////////////////////////////////////////////////////////

func NewPartManager() PartManager {
	return &innerPartManager{
		byID:   make(map[PartID]*PartInfo),
		byName: make(map[PartName]*PartInfo),
	}
}

////////////////////////////////////////////////////////////////////////////////

type innerPartManager struct {
	mu sync.RWMutex

	nextID PartID
	byID   map[PartID]*PartInfo
	byName map[PartName]*PartInfo
}

const firstAutoPartID PartID = 0x123456789

// RegisterAny implements [PartRegistry].
func (inst *innerPartManager) RegisterAny(o any) error {
	if inst == nil {
		return fmt.Errorf("part manager is nil")
	}
	if o == nil {
		return fmt.Errorf("part is nil")
	}

	switch value := o.(type) {
	case *PartInfo:
		return inst.Register(value)
	case PartProvider:
		return inst.RegisterProvider(value)
	default:
		// return fmt.Errorf("unsupported part type %T", o)
		return inst.innerRegisterPart(o)
	}
}

func (inst *innerPartManager) innerRegisterPart(part any) error {
	info := &PartInfo{
		ID:    0,
		Class: "",
		Name:  "",
		Label: "",
		Part:  part,
	}
	return inst.Register(info)
}

// Register implements [PartRegistry].
func (inst *innerPartManager) Register(info *PartInfo) error {
	if inst == nil {
		return fmt.Errorf("part manager is nil")
	}
	if info == nil {
		return fmt.Errorf("part info is nil")
	}
	if innerIsNil(info.Part) {
		return fmt.Errorf("part %q has no instance", info.Name)
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.initLocked()

	if info.ID == 0 {
		info.ID = inst.innerMakeNextIDLocked()
		if info.ID == 0 {
			return fmt.Errorf("part id space is exhausted")
		}
	}
	if info.Name == "" {
		info.Name = PartName(fmt.Sprintf("unnamed_part_%d", info.ID))
	}

	if existing := inst.byID[info.ID]; existing != nil {
		return fmt.Errorf("part id %d is already registered as %q", info.ID, existing.Name)
	}
	if existing := inst.byName[info.Name]; existing != nil {
		return fmt.Errorf("part name %q is already registered with id %d", info.Name, existing.ID)
	}

	inst.byID[info.ID] = info
	inst.byName[info.Name] = info
	return nil
}

// RegisterProvider implements [PartRegistry].
func (inst *innerPartManager) RegisterProvider(provider PartProvider) error {
	if inst == nil {
		return fmt.Errorf("part manager is nil")
	}
	if innerIsNil(provider) {
		return fmt.Errorf("part provider is nil")
	}

	infos := provider.ListParts()
	for _, info := range infos {
		if err := inst.Register(info); err != nil {
			return err
		}
	}
	return nil
}

// GetInfo implements [PartManager].
func (inst *innerPartManager) GetInfo(id PartID) (*PartInfo, error) {
	if inst == nil {
		return nil, fmt.Errorf("part manager is nil")
	}

	inst.mu.RLock()
	info := inst.byID[id]
	inst.mu.RUnlock()
	if info == nil {
		return nil, fmt.Errorf("part with id %d is not registered", id)
	}
	return info, nil
}

// GetInfoByName implements [PartManager].
func (inst *innerPartManager) GetInfoByName(name PartName) (*PartInfo, error) {
	if inst == nil {
		return nil, fmt.Errorf("part manager is nil")
	}

	inst.mu.RLock()
	info := inst.byName[name]
	inst.mu.RUnlock()
	if info == nil {
		return nil, fmt.Errorf("part with name %q is not registered", name)
	}
	return info, nil
}

// GetPart implements [PartManager].
func (inst *innerPartManager) GetPart(id PartID) (any, error) {
	info, err := inst.GetInfo(id)
	if err != nil {
		return nil, err
	}
	if innerIsNil(info.Part) {
		return nil, fmt.Errorf("part %q has no instance", info.Name)
	}
	return info.Part, nil
}

// GetRegistry implements [PartManager].
func (inst *innerPartManager) GetRegistry() PartRegistry {
	if inst == nil {
		return nil
	}
	return inst
}

// ListNames implements [PartManager].
func (inst *innerPartManager) ListNames() []PartName {
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

func (inst *innerPartManager) innerMakeNextIDLocked() PartID {
	if inst.nextID == 0 {
		inst.nextID = firstAutoPartID
	}

	for id := inst.nextID; id != 0; id++ {
		if inst.byID[id] == nil {
			inst.nextID = id + 1
			return id
		}
	}

	// Registration of every possible PartID is not practically reachable. If
	// it is reached, there is no ID left to allocate.
	return 0
}

func (inst *innerPartManager) initLocked() {
	if inst.byID == nil {
		inst.byID = make(map[PartID]*PartInfo)
	}
	if inst.byName == nil {
		inst.byName = make(map[PartName]*PartInfo)
	}
}

func innerIsNil(o any) bool {
	if o == nil {
		return true
	}
	value := reflect.ValueOf(o)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////
// EOF
