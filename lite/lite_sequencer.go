package lite

import (
	"fmt"
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////
// Note

type Note struct {
	Key      MidiNoteNumber
	Velocity MidiNoteVelocity
}

////////////////////////////////////////////////////////////////////////////////
// cell

type Cell struct {
	BankIndex  BankIndex
	StepIndex  StepIndex
	TrackIndex TrackIndex
}

type SimpleCell struct {
	Cell

	StepRef  *Step
	BankRef  *Bank
	TrackRef *Track

	Complex *ComplexCell
}

// ComplexCell : 每个 ComplexCell 占用一个或多个 SimpleCell 的坑位
type ComplexCell struct {
	Cell

	First       *SimpleCell
	SimpleCells []*SimpleCell

	Notes []*Note // 每个 ComplexCell 可以平均拆分成一个或多个 Note
}

/************************/
/* SimpleCellCache      */
/************************/

type SimpleCellCache struct {
	list []*SimpleCell
}

func (inst *SimpleCellCache) sort() {

}

/************************/
/* SimpleCellManager    */
/************************/

type SimpleCellManager struct {
	table map[string]*SimpleCell
}

func (inst *SimpleCellManager) OpenCell(want *Cell, autoMake bool) (*SimpleCell, error) {

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

func (inst *SimpleCellManager) innerGetKeyForCell(want *Cell) string {
	if want == nil {
		return ""
	}
	return fmt.Sprintf("b%d_s%d_t%d", want.BankIndex, want.StepIndex, want.TrackIndex)
}

func (inst *SimpleCellManager) innerGetTable() map[string]*SimpleCell {
	t := inst.table
	if t == nil {
		t = make(map[string]*SimpleCell)
		inst.table = t
	}
	return t
}

////////////////////////////////////////////////////////////////////////////////
// step

type StepIndex uint

type StepCount uint

/**********************/
/* Step (position)    */
/**********************/

type StepPosition struct {
	Index    StepIndex
	Count    StepCount
	Position Seconds
	AtLine   *Timeline
}

/**********************/
/* Step (head)        */
/**********************/

type StepHead struct {
	Index  StepIndex
	Weight byte
}

/**********************/
/* Step (body)        */
/**********************/

type StepBody struct {
	Head *StepHead

	Cells SimpleCellCache
}

/**********************/
/* Step (item)        */
/**********************/

type Step struct {
	Head *StepHead
	Body *StepBody
}

/**********************/
/* Step (List)        */
/**********************/

type StepList struct {
	list     []*Step // items
	size     uint    // 节拍(步奏)数量(有效值)
	capacity uint    // 节拍(步奏)数量(最大值)
}

func (inst *StepList) SetCapacity(cap uint) uint {
	const max = 16
	if cap > max {
		cap = max
	}
	inst.capacity = cap
	return cap
}

func (inst *StepList) GetSize() uint {
	return inst.size
}

func (inst *StepList) GetStep(index StepIndex) *Step {
	all := inst.list
	cnt := len(all)
	idx := int(index)
	if idx < cnt {
		return all[idx]
	}
	return nil
}

func (inst *StepList) SetSize(size uint) uint {
	max := inst.capacity
	if size > max {
		size = max
	}
	inst.size = size
	inst.updateList()
	return size
}

func (inst *StepList) Init(size uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(size)
		inst.SetSize(size)
	}
}

func (inst *StepList) Items() []*Step {
	all := inst.list
	want := int(inst.size)
	have := len(all)
	if want < have {
		return all[0:want]
	}
	return all
}

func (inst *StepList) updateList() {

	have := len(inst.list)
	want := int(inst.size)

	for ; have < want; have++ {

		item := new(Step)
		head := new(StepHead)

		item.Head = head

		head.Index = StepIndex(have)
		head.Weight = 60

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// track

type TrackIndex int

type TrackName string

type TrackHead struct {
	Index TrackIndex

	Name TrackName

	Label string

	Instrument Instrument
}

type TrackBody struct {
	Head *TrackHead

	Weight byte

	Cells SimpleCellCache
}

type Track struct {
	Head *TrackHead
	Body *TrackBody
}

/**********************/
/* TrackList          */
/**********************/

type TrackList struct {
	list     []*Track //items
	size     uint     // 轨道数量(有效值)
	capacity uint     // 轨道数量(最大值)
}

func (inst *TrackList) SetCapacity(cap uint) uint {
	const max = 16
	if cap > max {
		cap = max
	}
	inst.capacity = cap
	return cap
}

func (inst *TrackList) SetSize(size uint) uint {
	max := inst.capacity
	if size > max {
		size = max
	}
	inst.size = size
	inst.updateList()
	return size
}

func (inst *TrackList) GetSize() uint {
	return inst.size
}

func (inst *TrackList) GetTrack(index TrackIndex) *Track {
	all := inst.list
	cnt := len(all)
	idx := int(index)
	if idx < cnt {
		return all[idx]
	}
	return nil
}

func (inst *TrackList) Init(size uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(size)
		inst.SetSize(size)
	}
}

func (inst *TrackList) Items() []*Track {
	all := inst.list
	want := int(inst.size)
	have := len(all)
	if want < have {
		return all[0:want]
	}
	return all
}

func (inst *TrackList) updateList() {

	have := len(inst.list)
	want := int(inst.size)

	for ; have < want; have++ {

		item := new(Track)
		head := new(TrackHead)
		nstr := strconv.Itoa(have)

		item.Head = head

		head.Index = TrackIndex(have)
		head.Name = TrackName("track" + nstr)
		head.Label = "Track-" + nstr

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// bank

type BankIndex uint

type BankName string

// 存储区
type BankHead struct {
	Index BankIndex
	Name  BankName
	Label string
}

type BankBody struct {
	Head *BankHead

	Steps StepList

	CellManager SimpleCellManager
}

type Bank struct {
	Head *BankHead
	Body *BankBody
}

/**********************/
/* BankList           */
/**********************/

type BankList struct {
	list     []*Bank
	size     uint // 存储区数量(有效值)
	capacity uint // 存储区数量(最大值)
}

func (inst *BankList) SetCapacity(cap uint) uint {
	const max = 16
	if cap > max {
		cap = max
	}
	inst.capacity = cap
	return cap
}

func (inst *BankList) SetSize(size uint) uint {
	max := inst.capacity
	if size > max {
		size = max
	}
	inst.size = size
	inst.updateList()
	return size
}

func (inst *BankList) GetSize() uint {
	return inst.size
}

func (inst *BankList) GetBank(index BankIndex) *Bank {
	all := inst.list
	cnt := len(all)
	idx := int(index)
	if idx < cnt {
		return all[idx]
	}
	return nil
}

func (inst *BankList) Init(size uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(size)
		inst.SetSize(size)
	}
}

func (inst *BankList) Items() []*Bank {
	all := inst.list
	want := int(inst.size)
	have := len(all)
	if want < have {
		return all[0:want]
	}
	return all
}

func (inst *BankList) updateList() {

	have := len(inst.list)
	want := int(inst.size)

	for ; have < want; have++ {

		item := new(Bank)
		head := new(BankHead)
		nstr := strconv.Itoa(have)

		item.Head = head

		head.Index = BankIndex(have)
		head.Name = BankName("bank" + nstr)
		head.Label = "Bank-" + nstr

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// Sequencer

type SequencerPart interface {
	Update(ctx *SequencerContext) error
}

type Sequencer interface {
	ClockHandler

	SequencerPart

	GetContext() *SequencerContext

	GetBanks() []*Bank
	GetSteps() []*Step
	GetTracks() []*Track
}

type SequencerFactory interface {
	CreateSequencer(ctx *Context) (Sequencer, error)
}

type SequencerContext struct {
	Engine Engine

	EngineContext *EngineContext

	Sequencer Sequencer

	CurrentBank      *Bank
	CurrentBankIndex BankIndex

	Banks BankList

	Tracks TrackList

	Steps StepList

	SecondsPerStep Seconds // 每一步的时长, 由 bpm 和 tempo 计算所得
	SecondsPerBeat Seconds // 每一拍的时长, 由 bpm 和 tempo 计算所得
	StepsPerBeat   uint8   // 每一拍包含的步数

	TotalSteps StepCount //  总步数@work.timeline

	Updating bool // 1 表示有内容需要更新
}

////////////////////////////////////////////////////////////////////////////////
// EOF
