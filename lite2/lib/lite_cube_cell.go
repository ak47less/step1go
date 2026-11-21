package lib

type CellName string

type CellCoordinate struct {
	BankIndex  BankIndex
	StepIndex  StepIndex
	TrackIndex TrackIndex
}

type SimpleCell struct {
	CellCoordinate

	OwnerCube  *Cube
	OwnerBank  *Bank
	OwnerStep  *Step
	OwnerTrack *Track

	Complex *ComplexCell

	Name CellName
}

type ComplexCell struct {
	CellCoordinate
}

type Cell = SimpleCell

type CellManager interface {
	CreateCells(cube *Cube, cfg *CubeConfig) error

	GetCell(coor CellCoordinate) (*Cell, error)

	GetCellByName(name CellName) (*Cell, error)

	ListAllCells() []*Cell
}
