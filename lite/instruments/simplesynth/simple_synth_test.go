package simplesynth

import (
	"testing"

	"github.com/ak47less/step1go/lite"
)

type recordingAudioContext struct {
	now         lite.Seconds
	oscillators []*recordingOscillator
	destination *recordingDestination
}

func newRecordingAudioContext() *recordingAudioContext {
	ctx := &recordingAudioContext{}
	ctx.destination = new(recordingDestination)
	return ctx
}

func (inst *recordingAudioContext) Close() {}

func (inst *recordingAudioContext) Resume() {}

func (inst *recordingAudioContext) Suspend() {}

func (inst *recordingAudioContext) CreateOscillator() lite.OscillatorNode {
	source := &recordingOscillator{
		frequency: new(recordingAudioParam),
		detune:    new(recordingAudioParam),
	}
	inst.oscillators = append(inst.oscillators, source)
	return source
}

func (inst *recordingAudioContext) CreateGain() lite.GainNode {
	return &recordingGain{gain: new(recordingAudioParam)}
}

func (inst *recordingAudioContext) CreateBuffer(
	int,
	lite.SampleCount,
	lite.SampleRate,
) lite.AudioBuffer {
	return nil
}

func (inst *recordingAudioContext) CreateBufferSource() lite.AudioBufferSourceNode {
	return nil
}

func (inst *recordingAudioContext) GetCurrentTime() lite.Seconds {
	return inst.now
}

func (inst *recordingAudioContext) GetSampleRate() lite.SampleRate {
	return 44100
}

func (inst *recordingAudioContext) GetDestination() lite.AudioDestinationNode {
	return inst.destination
}

func (inst *recordingAudioContext) GetState() lite.AudioContextState {
	return lite.AudioContextStateSuspended
}

type recordingAudioParam struct{}

func (inst *recordingAudioParam) GetValue() float64 {
	return 0
}

func (inst *recordingAudioParam) SetValue(float64) {}

func (inst *recordingAudioParam) SetValueAtTime(float64, lite.Seconds) {}

func (inst *recordingAudioParam) LinearRampToValueAtTime(float64, lite.Seconds) {}

func (inst *recordingAudioParam) ExponentialRampToValueAtTime(float64, lite.Seconds) {}

func (inst *recordingAudioParam) SetTargetAtTime(float64, lite.Seconds, lite.Seconds) {}

func (inst *recordingAudioParam) CancelScheduledValues(lite.Seconds) {}

type recordingGain struct {
	gain *recordingAudioParam
}

func (inst *recordingGain) Connect(lite.AudioNode) {}

func (inst *recordingGain) Disconnect() {}

func (inst *recordingGain) GetGain() lite.AudioParam {
	return inst.gain
}

type recordingDestination struct{}

func (inst *recordingDestination) Connect(lite.AudioNode) {}

func (inst *recordingDestination) Disconnect() {}

func (inst *recordingDestination) GetMaxChannelCount() int {
	return 2
}

type recordingOscillator struct {
	frequency *recordingAudioParam
	detune    *recordingAudioParam
	kind      lite.OscillatorType
	starts    []lite.Seconds
	stops     []lite.Seconds
	onEnded   func()
}

func (inst *recordingOscillator) Connect(lite.AudioNode) {}

func (inst *recordingOscillator) Disconnect() {}

func (inst *recordingOscillator) Start(when lite.Seconds) {
	inst.starts = append(inst.starts, when)
}

func (inst *recordingOscillator) Stop(when lite.Seconds) {
	inst.stops = append(inst.stops, when)
}

func (inst *recordingOscillator) SetOnEnded(handler func()) {
	inst.onEnded = handler
}

func (inst *recordingOscillator) GetFrequency() lite.AudioParam {
	return inst.frequency
}

func (inst *recordingOscillator) GetDetune() lite.AudioParam {
	return inst.detune
}

func (inst *recordingOscillator) GetType() lite.OscillatorType {
	return inst.kind
}

func (inst *recordingOscillator) SetType(kind lite.OscillatorType) {
	inst.kind = kind
}

func TestDispatchSchedulesNoteOffForScheduledNoteOn(t *testing.T) {
	ctx := newRecordingAudioContext()
	synth := &SimpleSynth{
		audio:       ctx,
		destination: ctx.destination,
		waveform:    "sine",
		voices:      make(map[lite.MidiNoteNumber]*simpleSynthVoice),
	}
	on := &lite.MidiNote{
		Position: 1,
		Key:      60,
		Velocity: 100,
	}
	off := &lite.MidiNote{
		Position: 1.5,
		Key:      on.Key,
	}

	err := synth.Dispatch(&lite.MidiMessage{
		NoteOn:  on,
		NoteOff: off,
	})
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if len(ctx.oscillators) != 1 {
		t.Fatalf("created %d oscillators, want 1", len(ctx.oscillators))
	}

	source := ctx.oscillators[0]
	if len(source.starts) != 1 || source.starts[0] != lite.Seconds(on.Position) {
		t.Fatalf("Start calls = %v, want [%v]", source.starts, on.Position)
	}
	if len(source.stops) != 1 || source.stops[0] != lite.Seconds(off.Position) {
		t.Fatalf("Stop calls = %v, want [%v]", source.stops, off.Position)
	}
}
