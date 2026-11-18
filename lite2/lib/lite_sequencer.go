package lib

type SequencerContext struct {
	AC *ApplicationContext

	Sequencer Sequencer
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

func (inst *DefaultSequencer) _impl() (Sequencer, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
