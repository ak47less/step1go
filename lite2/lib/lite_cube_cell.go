package lib

import (
	"fmt"
	"sort"
)

////////////////////////////////////////////////////////////////////////////////

type CellName string

type Cell = SimpleCell

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////

type CellManager interface {
	GetCell(coor CellCoordinate) (*Cell, error)

	GetCellByName(name CellName) (*Cell, error)

	ListAllCells() []*Cell
}

////////////////////////////////////////////////////////////////////////////////

func NewCellManager(cube *Cube, cfg *CubeConfig) (CellManager, error) {
	cm := new(innerCellManager)
	err := cm.init(cube, cfg)
	return cm, err
}

////////////////////////////////////////////////////////////////////////////////

type innerCellManager struct {
	table map[CellName]*SimpleCell
}

// GetCell implements [CellManager].
func (inst *innerCellManager) GetCell(coor CellCoordinate) (*Cell, error) {
	name := inst.innerComputeCellName(coor)
	return inst.GetCellByName(name)
}

// GetCellByName implements [CellManager].
func (inst *innerCellManager) GetCellByName(name CellName) (*Cell, error) {

	if inst == nil {
		return nil, fmt.Errorf("innerCellManager instrance is nil")
	}

	tab := inst.table
	if tab == nil {
		return nil, fmt.Errorf("cell table is nil")
	}

	cell := tab[name]
	if cell == nil {
		return nil, fmt.Errorf("no cell with name '%s'", name)
	}
	return cell, nil
}

func (inst *innerCellManager) init(cube *Cube, cfg *CubeConfig) error {

	if inst == nil || cube == nil || cfg == nil {
		return fmt.Errorf("param(s) is nil")
	}

	count_bank := cfg.BankListCapacity
	count_step := cfg.StepListCapacity
	count_track := cfg.TrackListCapacity
	var coo CellCoordinate

	table := make(map[CellName]*Cell)

	for ibank := 0; ibank < int(count_bank); ibank++ {

		for istep := 0; istep < int(count_step); istep++ {

			for itrack := 0; itrack < int(count_track); itrack++ {

				coo.BankIndex = BankIndex(ibank)
				coo.StepIndex = StepIndex(istep)
				coo.TrackIndex = TrackIndex(itrack)
				name := inst.innerComputeCellName(coo)
				cell := &SimpleCell{
					Name:      name,
					OwnerCube: cube,
				}
				cell.CellCoordinate = coo
				table[name] = cell
			}
		}
	}

	inst.table = table
	return nil
}

// ListAllCells implements [CellManager].
func (inst *innerCellManager) ListAllCells() []*Cell {

	src := inst.table
	dst := make([]*Cell, 0)

	for _, cell := range src {
		dst = append(dst, cell)
	}

	sort.Slice(dst, func(i, j int) bool {

		o1 := dst[i]
		o2 := dst[j]
		var diff int

		diff = int(o1.BankIndex) - int(o2.BankIndex)
		if diff != 0 {
			return diff < 0
		}

		diff = int(o1.StepIndex) - int(o2.StepIndex)
		if diff != 0 {
			return diff < 0
		}

		diff = int(o1.TrackIndex) - int(o2.TrackIndex)
		return diff < 0
	})

	return dst
}

func (inst *innerCellManager) innerComputeCellName(coo CellCoordinate) CellName {
	const f = "b%d_s%d_t%d"
	str := fmt.Sprintf(f, coo.BankIndex, coo.StepIndex, coo.TrackIndex)
	return CellName(str)
}

func (inst *innerCellManager) _impl() CellManager {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
