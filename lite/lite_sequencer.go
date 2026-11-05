package lite

////////////////////////////////////////////////////////////////////////////////
// Note

type Note struct {
	Key      MidiNoteNumber
	Velocity MidiNoteVelocity
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

	Cube Cube

	SecondsPerStep Seconds // 每一步的时长, 由 bpm 和 tempo 计算所得
	SecondsPerBeat Seconds // 每一拍的时长, 由 bpm 和 tempo 计算所得
	StepsPerBeat   uint8   // 每一拍包含的步数

	STC *StepTickContext

	Updating bool // 1 表示有内容需要更新
}

func (inst *SequencerContext) ComputeStepLocation(pos *RelativePosition) StepLocation {

	abs := pos.Absolute()
	wkline := inst.EngineContext.WorkTileline
	rp2 := abs.Relative(wkline, nil)

	const min_sps = 0.001
	sps := inst.SecondsPerStep
	if sps < min_sps {
		sps = min_sps
	}

	location := rp2.Position / sps
	return StepLocation(location)
}

func (inst *SequencerContext) ComputeStepIndex(pos *RelativePosition) StepIndex {

	location := inst.ComputeStepLocation(pos)
	cnt_step := inst.Cube.Steps.Length()

	if cnt_step < 1 {
		return 0
	}

	idx := uint(location) / cnt_step
	return StepIndex(idx)
}

func (inst *SequencerContext) NewBank(index BankIndex) *Bank {

	b := &Bank{}
	b.Init(&inst.Cube)
	b.Index = index

	return b
}

func (inst *SequencerContext) NewStep(index StepIndex) *Step {

	s := &Step{
		Index: index,
		cube:  &inst.Cube,
	}

	return s
}

func (inst *SequencerContext) NewTrack(index TrackIndex) *Track {

	t := &Track{
		Index: index,
		cube:  &inst.Cube,
	}

	return t
}

////////////////////////////////////////////////////////////////////////////////
// EOF
