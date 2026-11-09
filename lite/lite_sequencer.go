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

////////////////////////////////////////////////////////////////////////////////

type SequencerInfo struct {
	Profile MusicProfile

	SecondsPerStep Seconds // 每一步的时长, 由 bpm 和 tempo 计算所得
	SecondsPerBeat Seconds // 每一拍的时长, 由 bpm 和 tempo 计算所得
	StepsPerBeat   uint8   // 每一拍包含的步数
	StepsPerCycle  uint    // 每一个循环包含的步数
	PreloadSeconds Seconds // 需要提前加载的时长

}

func (inst *SequencerInfo) Normalize(rewrite bool) *SequencerInfo {

	var src, dst *SequencerInfo
	src = inst

	if src == nil {
		src = new(SequencerInfo)
	}

	if rewrite {
		dst = src
	} else {
		dst = new(SequencerInfo)
	}

	sec_p_step := src.SecondsPerStep
	sec_p_beat := src.SecondsPerBeat
	step_p_beat := src.StepsPerBeat
	step_p_cycl := src.StepsPerCycle
	pre_sec := src.PreloadSeconds
	profile := src.Profile

	const (
		sec_min = 0.001
	)

	if profile == nil {
		sp := &SimpleProfile{}
		profile = sp.Init()
	}

	if pre_sec < 1.0 {
		pre_sec = 1.0
	}

	if step_p_cycl < 4 {
		step_p_cycl = 16
	}

	if step_p_beat < 1 {
		step_p_beat = 1
	}

	if sec_p_beat < sec_min {
		sec_p_beat = sec_min
	}

	if sec_p_step < sec_min {
		sec_p_step = sec_min
	}

	dst.PreloadSeconds = pre_sec
	dst.SecondsPerBeat = sec_p_beat
	dst.SecondsPerStep = sec_p_step
	dst.StepsPerBeat = step_p_beat
	dst.StepsPerCycle = step_p_cycl
	dst.Profile = profile

	return dst
}

////////////////////////////////////////////////////////////////////////////////

type SequencerContext struct {
	Engine Engine

	EngineContext *EngineContext

	Sequencer Sequencer

	CurrentBank       *Bank
	CurrentBankIndex  BankIndex
	CurrentBankHolder BankHolder

	Cube Cube

	Info SequencerInfo

	STC *StepTickContext

	Updating bool // 1 表示有内容需要更新
}

func (inst *SequencerContext) NormalizeSTP(stp *StepTickPos, align bool) {

	location := inst.ComputeStepLocation(&stp.Position)
	index := inst.ComputeStepIndex(&stp.Position)

	if align {
		wkline := inst.EngineContext.WorkTileline
		sps := inst.Info.SecondsPerStep
		stp.Position.Pos = sps * Seconds(location)
		stp.Position.Ref = wkline
	}

	stp.Index = index
	stp.Location = location

}

func (inst *SequencerContext) ComputeStepLocation(pos *RelativePosition) StepLocation {

	if pos == nil {
		return 0
	}

	abs := pos.Absolute()
	wkline := inst.EngineContext.WorkTileline
	rp2 := abs.Relative(wkline, nil)

	const min_sps = 0.001
	sps := inst.Info.SecondsPerStep
	if sps < min_sps {
		sps = min_sps
	}

	location := rp2.Pos / sps
	return StepLocation(location)
}

func (inst *SequencerContext) ComputeStepIndex(pos *RelativePosition) StepIndex {

	location := inst.ComputeStepLocation(pos)
	cnt_step := inst.Info.StepsPerCycle

	if cnt_step < 1 {
		return 0
	}

	idx := uint(location) % cnt_step
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
