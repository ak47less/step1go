package lib

type SequencerContext struct {
	AC *ApplicationContext

	Sequencer Sequencer
}

type Sequencer interface {
	GetSequencerContext() *SequencerContext
}
