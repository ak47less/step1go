package lib

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
	return new(innerPartManager)
}

////////////////////////////////////////////////////////////////////////////////

type innerPartManager struct {
}

// GetInfo implements [PartManager].
func (inst *innerPartManager) GetInfo(id PartID) (*PartInfo, error) {
	panic("unimplemented")
}

// GetInfoByName implements [PartManager].
func (inst *innerPartManager) GetInfoByName(name PartName) (*PartInfo, error) {
	panic("unimplemented")
}

// GetPart implements [PartManager].
func (inst *innerPartManager) GetPart(id PartID) (any, error) {
	panic("unimplemented")
}

// GetRegistry implements [PartManager].
func (inst *innerPartManager) GetRegistry() PartRegistry {
	panic("unimplemented")
}

// ListNames implements [PartManager].
func (inst *innerPartManager) ListNames() []PartName {
	panic("unimplemented")
}

////////////////////////////////////////////////////////////////////////////////
// EOF
