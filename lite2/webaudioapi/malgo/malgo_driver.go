package malgo

import (
	"github.com/ak47less/step1go/lite2/webaudioapi"
	"github.com/gen2brain/malgo"
)

////////////////////////////////////////////////////////////////////////////////

type TheMalgoDriver struct {
}

// GetFactory implements [webaudioapi.Driver].
func (inst *TheMalgoDriver) GetFactory() webaudioapi.AudioContextFactory {
	return new(innerMalgoAudioContextFactory)
}

func (inst *TheMalgoDriver) _impl() webaudioapi.Driver {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerMalgoAudioContextFactory struct {
	config malgo.ContextConfig
}

// CreateAudioContext implements [webaudioapi.AudioContextFactory].
func (inst *innerMalgoAudioContextFactory) CreateAudioContext() (*webaudioapi.AudioContext, error) {
	panic("unimplemented")
}

func (inst *innerMalgoAudioContextFactory) _impl() webaudioapi.AudioContextFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
