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

//////////////////////////////////////////////////////////////

/************************/
/* SimpleCell           */
/************************/

type SimpleCell struct {
	Cell

	StepRef  *Step
	BankRef  *Bank
	TrackRef *Track

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

	track := inst.TrackRef
	if track == nil {
		return fmt.Errorf("SimpleCell.track_ref is nil")
	}
	in := track.Instrument
	if in == nil {
		return fmt.Errorf("track.Instrument is nil")
	}

	msg := new(MidiMessage)
	noteOn := new(MidiNote)
	noteOff := new(MidiNote)
	chl := track.Channel
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

//////////////////////////////////////////////////////////////

/************************/
/* ComplexCell          */
/************************/

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

//////////////////////////////////////////////////////////////

/************************/
/* CellCache            */
/************************/

type CellCache struct {
	list []*SimpleCell
}

func (inst *CellCache) sort() {

}

/************************/
/* CellManager    */
/************************/

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
// step (配件)

// 物理下标
type StepIndex uint

// 逻辑下标
type StepLocation uint

/**********************/
/* Step (position)    */
/**********************/

type StepPosition struct {

	// Position Seconds
	// At       *Timeline

	RelativePosition

	Index    StepIndex
	Location StepLocation
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

func (inst *StepList) Init(size, capacity uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(capacity)
		inst.SetSize(size)
	}
}

func (inst *StepList) Items() []*Step {

	if inst == nil {
		return nil
	}

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

		item.Index = StepIndex(have)
		item.Weight = 60

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// step  (主体结构)

type Step struct {

	// Head

	Index StepIndex

	Weight byte

	// Body

	Context *SequencerContext

	Cells CellCache
}

// OnStep implements [StepTickHandler].
func (inst *Step) OnStep(stc *StepTickContext) error {

	if inst == nil {
		return fmt.Errorf("Step-instance is nil")
	}

	var err error
	clist := inst.Cells.list

	for _, cell := range clist {
		err = cell.OnStep(stc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Step) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// track

type TrackIndex int

type TrackName string

type Track struct {

	// Head

	Index   TrackIndex
	Name    TrackName
	Label   string
	Weight  byte
	Channel byte // midi-channel

	// Body

	Context *SequencerContext

	Instrument Instrument

	Cells CellCache
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

func (inst *TrackList) Init(size, capacity uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(capacity)
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
		nstr := strconv.Itoa(have)

		item.Index = TrackIndex(have)
		item.Name = TrackName("track" + nstr)
		item.Label = "Track-" + nstr

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// bank (配件)

type BankIndex uint

type BankName string

/**********************/
/* BankHolder           */
/**********************/

type BankHolder interface {
	GetBank() (*Bank, error)
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

func (inst *BankList) Init(size, capacity uint) {
	if inst.capacity == 0 {
		inst.SetCapacity(capacity)
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
		nstr := strconv.Itoa(have)

		item.Index = BankIndex(have)
		item.Name = BankName("bank" + nstr)
		item.Label = "Bank-" + nstr

		inst.list = append(inst.list, item)
	}

}

////////////////////////////////////////////////////////////////////////////////
// bank (存储区, 主体结构)

type Bank struct {

	// head

	Index BankIndex
	Name  BankName
	Label string

	// body

	Context *SequencerContext

	CellManager *CellManager
	Steps       *StepList
	Tracks      *TrackList
}

// OnStep implements [StepTickHandler].
func (inst *Bank) OnStep(stc *StepTickContext) error {

	if inst == nil {
		return fmt.Errorf("bank-instance is nil")
	}

	if stc == nil {
		return fmt.Errorf("param:STC(StepTickContext) is nil")
	}

	stp := &stc.Current
	idx := stp.Index
	step := inst.Steps.GetStep(idx)

	return step.OnStep(stc)
}

func (inst *Bank) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// Sequencer

type SequencerPart interface {
	Update(ctx *SequencerContext) error
}

type Sequencer interface {
	StepTickHandler

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

	CurrentBank       *Bank
	CurrentBankIndex  BankIndex
	CurrentBankHolder BankHolder

	// sets

	CellManager *CellManager
	Banks       BankList
	Tracks      TrackList
	Steps       StepList

	SecondsPerStep Seconds // 每一步的时长, 由 bpm 和 tempo 计算所得
	SecondsPerBeat Seconds // 每一拍的时长, 由 bpm 和 tempo 计算所得
	StepsPerBeat   uint8   // 每一拍包含的步数

	STC *StepTickContext

	Updating bool // 1 表示有内容需要更新
}

func (inst *SequencerContext) NewBank(index BankIndex) *Bank {

	list_s := &inst.Steps
	list_t := &inst.Tracks

	b := &Bank{
		Index:       index,
		Context:     inst,
		CellManager: inst.CellManager,
		Tracks:      list_t,
		Steps:       list_s,
	}

	return b
}

func (inst *SequencerContext) NewStep(index StepIndex) *Step {

	s := &Step{
		Index:   index,
		Context: inst,
	}

	return s
}

func (inst *SequencerContext) NewTrack(index TrackIndex) *Track {

	t := &Track{
		Index:   index,
		Context: inst,
	}

	return t
}

////////////////////////////////////////////////////////////////////////////////
// EOF
