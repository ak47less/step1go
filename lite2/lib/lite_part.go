package lib

type PartName string

type PartClass string

type PartID int

type PartInfo struct {
	Class PartClass
	Name  PartName
	ID    PartID
	Label string
	Part  any
}

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
