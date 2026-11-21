package lib

type SequencerContext struct {
	AC *ApplicationContext

	Sequencer Sequencer

	Cube Cube
}

type Sequencer interface {
	GetSequencerContext() *SequencerContext
}

////////////////////////////////////////////////////////////////////////////////

type DefaultSequencer struct {
	context *SequencerContext
}

// GetSequencerContext implements [Sequencer].
func (inst *DefaultSequencer) GetSequencerContext() *SequencerContext {
	panic("unimplemented")
}

// GetLifeCycle implements [Life].
func (inst *DefaultSequencer) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit: inst.onInit,
	}
	return lc
}

func (inst *DefaultSequencer) onInit(ac *ApplicationContext) error {

	sc := ac.SequencerContext

	sc.Sequencer = inst
	ac.Sequencer = inst
	inst.context = sc

	return nil
}

func (inst *DefaultSequencer) innerInitCube(ac *ApplicationContext) error {

	cfg2 := new(CubeConfig)
	cube := &inst.context.Cube
	pt1 := ac.Properties
	gett := pt1.Getter()

	const (
		k_banks_len  = "cube.banks.length"
		k_steps_len  = "cube.steps.length"
		k_tracks_len = "cube.tracks.length"
		k_banks_cap  = "cube.banks.capacity"
		k_steps_cap  = "cube.steps.capacity"
		k_tracks_cap = "cube.tracks.capacity"
	)

	banks_capacity := gett.GetUint(k_banks_cap)
	steps_capacity := gett.GetUint(k_steps_cap)
	tracks_capacity := gett.GetUint(k_tracks_cap)
	banks_len := gett.GetUint(k_banks_len)
	steps_len := gett.GetUint(k_steps_len)
	tracks_len := gett.GetUint(k_tracks_len)

	cfg2.BankListCapacity = banks_capacity
	cfg2.StepListCapacity = steps_capacity
	cfg2.TrackListCapacity = tracks_capacity
	cfg2.BankListLength = banks_len
	cfg2.StepListLength = steps_len
	cfg2.TrackListLength = tracks_len

	err := gett.Error()
	if err != nil {
		return err
	}

	return cube.Init(cfg2)
}

func (inst *DefaultSequencer) _impl() (Sequencer, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
