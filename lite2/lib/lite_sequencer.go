package lib

type SequencerContext struct {
	Sequencer Sequencer
}

type Sequencer interface {
	GetSequencerContext() *SequencerContext
}
