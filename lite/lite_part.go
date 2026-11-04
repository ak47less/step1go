package lite

////////////////////////////////////////////////////////////////////////////////

// PartID 是零部件的唯一标识。
type PartID int

// PartName 是零部件的名称。
type PartName string

// PartClass 表示零部件所属的分类，一个零部件可以属于多个分类。
type PartClass string

////////////////////////////////////////////////////////////////////////////////

// Part 表示可参与引擎运行并对外提供自身元数据的零部件。
type Part interface {
	GetInfo() *PartInfo
}

// PartFactory 根据零部件信息创建实例，并提供对应的注册信息。
type PartFactory interface {
	CreatePart(info *PartInfo) (Part, error)

	GetRegistration() *PartRegistration
}

// PartManager 负责按标识、名称或分类查询和列出零部件实例。
type PartManager interface {
	ListNames() []PartName

	FindPartByID(id PartID) (Part, error)

	FindPartByName(name PartName) (Part, error)

	ListPartsWithClass(cl PartClass) ([]Part, error)
}

// PartRegistry 管理零部件注册信息，并支持按标识、名称和分类查找。
type PartRegistry interface {
	Register(pr *PartRegistration) error

	ListNames() []PartName

	FindByID(id PartID) (*PartRegistration, error)

	FindByName(name PartName) (*PartRegistration, error)

	ListWithClass(cl PartClass) ([]*PartRegistration, error)
}

////////////////////////////////////////////////////////////////////////////////

// PartContext 汇集零部件运行和查找所需的引擎上下文、注册表和管理器。
type PartContext struct {
	EC *EngineContext

	Registry PartRegistry

	Manager PartManager
}

////////////////////////////////////////////////////////////////////////////////

// PartMeta 是零部件在创建、注册和运行阶段共享的基础元数据。
type PartMeta struct {
	ID      PartID
	Name    PartName
	Classes []PartClass

	Context *PartContext

	Factory PartFactory
	Part    Part
}

// PartInfo 描述提供给零部件工厂的元数据。
type PartInfo struct {
	PartMeta
}

// PartRegistration 保存零部件元数据以及负责创建该零部件的工厂。
type PartRegistration struct {
	PartMeta
}

////////////////////////////////////////////////////////////////////////////////
// EOF
