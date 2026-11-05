package lite

import (
	"fmt"
	"sort"
)

////////////////////////////////////////////////////////////////////////////////
// cell

type Cell struct {
	BankIndex  BankIndex
	StepIndex  StepIndex
	TrackIndex TrackIndex
}

////////////////////////////////////////////////////////////////////////////////
// SimpleCell

type SimpleCell struct {
	Cell

	cube *Cube

	ownerStep  *Step
	ownerBank  *Bank
	ownerTrack *Track

	Enabled bool

	Note *Note

	Complex *ComplexCell
}

// OnStep implements [StepTickHandler].
func (inst *SimpleCell) OnStep(stc *StepTickContext) error {

	if inst == nil {
		return fmt.Errorf("SimpleCell-instance is nil")
	}

	if !inst.Enabled {
		return nil
	}

	var err error
	complex := inst.Complex
	note := inst.Note

	if complex == nil {
		// as a simple-cell
		if note != nil {
			err = inst.innerDoSimpleCellNote(stc, note)
		}
	} else {
		// as a complex-cell
		if complex.First == inst {
			err = complex.OnStep(stc)
		}
	}

	return err
}

func (inst *SimpleCell) innerDoSimpleCellNote(stc *StepTickContext, note *Note) error {

	track := inst.ownerTrack
	if track == nil {
		return fmt.Errorf("SimpleCell.track_ref is nil")
	}

	meta := track.Meta
	if meta == nil {
		return fmt.Errorf("SimpleCell.track_meta is nil")
	}

	in := meta.Instrument
	if in == nil {
		return fmt.Errorf("track.meta.Instrument is nil")
	}

	msg := new(MidiMessage)
	noteOn := new(MidiNote)
	noteOff := new(MidiNote)
	chl := meta.Channel
	time := stc.Current
	steplen := stc.StepLength

	noteOn.Channel = chl
	noteOn.Key = note.Key
	noteOn.Velocity = note.Velocity
	noteOn.Position = time.Position.Position
	noteOn.At = time.Position.At

	noteOff.Channel = chl
	noteOff.Key = note.Key
	noteOff.Velocity = 0
	noteOff.Position = time.Position.Position + steplen
	noteOff.At = time.Position.At

	msg.NoteOn = noteOn
	msg.NoteOff = noteOff

	return in.Dispatch(msg)
}

func (inst *SimpleCell) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// ComplexCell

// ComplexCell : 每个 ComplexCell 占用一个或多个 SimpleCell 的坑位
type ComplexCell struct {
	Cell

	First       *SimpleCell
	SimpleCells []*SimpleCell

	Notes []*Note // 每个 ComplexCell 可以平均拆分成一个或多个 Note
}

// OnStep implements [StepTickHandler].
func (inst *ComplexCell) OnStep(cc *StepTickContext) error {
	panic("unimplemented")
}

func (inst *ComplexCell) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// CellCache

type CellCache struct {
	list []*SimpleCell
}

func (inst *CellCache) sort(fn func(a, b *SimpleCell) bool) {

	if inst == nil || fn == nil {
		return
	}

	all := inst.list
	sort.Slice(all, func(i, j int) bool {
		a := all[i]
		b := all[j]
		if a == nil || b == nil {
			return false
		}
		return fn(a, b)
	})
}

func (inst *CellCache) add(cell *SimpleCell) {
	if inst == nil || cell == nil {
		return
	}
	inst.list = append(inst.list, cell)
}

func (inst *CellCache) reset() {
	if inst == nil {
		return
	}
	inst.list = make([]*SimpleCell, 0)
}

////////////////////////////////////////////////////////////////////////////////
// CellManager

type CellManager struct {
	table map[string]*SimpleCell
}

func (inst *CellManager) OpenCell(want *Cell, autoMake bool) (*SimpleCell, error) {

	key := inst.innerGetKeyForCell(want)
	tab := inst.innerGetTable()
	have := tab[key]

	if have == nil {
		if autoMake {
			have = new(SimpleCell)
			have.BankIndex = want.BankIndex
			have.StepIndex = want.StepIndex
			have.TrackIndex = want.TrackIndex
			tab[key] = have
		} else {
			const f = "no item:[cell bank:%d step:%d track:%d] in manager"
			err := fmt.Errorf(f, want.BankIndex, want.StepIndex, want.TrackIndex)
			return nil, err
		}
	}

	return have, nil
}

func (inst *CellManager) innerGetKeyForCell(want *Cell) string {
	if want == nil {
		return ""
	}
	return fmt.Sprintf("b%d_s%d_t%d", want.BankIndex, want.StepIndex, want.TrackIndex)
}

func (inst *CellManager) innerGetTable() map[string]*SimpleCell {
	t := inst.table
	if t == nil {
		t = make(map[string]*SimpleCell)
		inst.table = t
	}
	return t
}

////////////////////////////////////////////////////////////////////////////////
// EOF
