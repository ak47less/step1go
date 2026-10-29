package malgo4lite

import (
	"github.com/ak47less/step1go/lite"
	"github.com/gen2brain/malgo"
)

////////////////////////////////////////////////////////////////////////////////
// class : MalgoDriver

type MalgoDriver struct {
}

// GetAudioContextFactory implements [lite.Driver].
func (inst *MalgoDriver) GetAudioContextFactory() lite.AudioContextFactory {
	return new(MalgoAudioContextFactory)
}

func (inst *MalgoDriver) _impl() lite.Driver {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class : MalgoAudioContextFactory

type MalgoAudioContextFactory struct {
}

// NewAudioContext implements [lite.AudioContextFactory].
func (inst *MalgoAudioContextFactory) NewAudioContext() lite.AudioContext {

	mc, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		panic(err)
	}

	ac := new(MalgoAudioContext)
	ac.mc = mc

	return ac
}

func (inst *MalgoAudioContextFactory) _impl() lite.AudioContextFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class: MalgoAudioContext

type MalgoAudioContext struct {
	mc *malgo.AllocatedContext
}

// Close implements [lite.AudioContext].
func (inst *MalgoAudioContext) Close() {
	panic("unimplemented")
}

// CreateBuffer implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateBuffer(numberOfChannels int, length lite.SampleCount, sampleRate lite.SampleRate) lite.AudioBuffer {
	panic("unimplemented")
}

// CreateBufferSource implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateBufferSource() lite.AudioBufferSourceNode {
	panic("unimplemented")
}

// CreateGain implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateGain() lite.GainNode {
	panic("unimplemented")
}

// CreateOscillator implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateOscillator() lite.OscillatorNode {
	panic("unimplemented")
}

// GetCurrentTime implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetCurrentTime() lite.Seconds {
	panic("unimplemented")
}

// GetDestination implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetDestination() lite.AudioDestinationNode {
	panic("unimplemented")
}

// GetSampleRate implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetSampleRate() lite.SampleRate {
	panic("unimplemented")
}

// GetState implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetState() lite.AudioContextState {
	panic("unimplemented")
}

// Resume implements [lite.AudioContext].
func (inst *MalgoAudioContext) Resume() {
	panic("unimplemented")
}

// Suspend implements [lite.AudioContext].
func (inst *MalgoAudioContext) Suspend() {
	panic("unimplemented")
}

func (inst *MalgoAudioContext) _impl() lite.AudioContext {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
