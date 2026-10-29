package simplesynth

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/ak47less/step1go/lite"
)

const simpleSynthNoiseSeconds = 1

type SimpleSynthFactory struct {

	// SimpleSynthFactory  :  这个工厂提供一个简单的合成器实现, 能生成: 正弦波 ,三角波,锯齿波,方波, 噪声波, 这些简单的波形

}

var (
	_ lite.InstrumentFactory = new(SimpleSynthFactory)
	_ lite.Instrument        = new(SimpleSynth)
)

// ListInfo implements [lite.InstrumentFactory].
func (inst *SimpleSynthFactory) ListInfo() []*lite.InstrumentInfo {
	return []*lite.InstrumentInfo{
		{
			Name:   "default",
			Label:  "Simple Synth",
			Preset: "default",
		},
	}
}

// CreateInstrument implements [lite.InstrumentFactory].
func (inst *SimpleSynthFactory) CreateInstrument(
	ctx *lite.Context,
	info *lite.InstrumentInfo,
) (lite.Instrument, error) {
	if ctx == nil {
		return nil, fmt.Errorf("simplesynth: context is nil")
	}
	if ctx.AudioContext == nil {
		return nil, fmt.Errorf("simplesynth: audio context is nil")
	}
	destination := ctx.AudioContext.GetDestination()
	if destination == nil {
		return nil, fmt.Errorf("simplesynth: audio destination is nil")
	}

	instrumentInfo := lite.InstrumentInfo{
		Name:   "default",
		Label:  "Simple Synth",
		Preset: "default",
	}
	if info != nil {
		instrumentInfo = *info
	}

	return &SimpleSynth{
		ctx:         ctx,
		audio:       ctx.AudioContext,
		destination: destination,
		info:        instrumentInfo,
		waveform:    simpleSynthWaveform(instrumentInfo.Preset),
		voices:      make(map[lite.MidiNoteNumber]*simpleSynthVoice),
	}, nil
}

func (inst *SimpleSynthFactory) _impl() lite.InstrumentFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// SimpleSynth

type SimpleSynth struct {
	ctx         *lite.Context
	audio       lite.AudioContext
	destination lite.AudioDestinationNode
	info        lite.InstrumentInfo
	waveform    string

	dispatchMu sync.Mutex
	mu         sync.Mutex
	voices     map[lite.MidiNoteNumber]*simpleSynthVoice
}

type simpleSynthVoice struct {
	source lite.AudioScheduledSourceNode
	gain   lite.GainNode
}

// Dispatch implements [lite.Instrument].
func (inst *SimpleSynth) Dispatch(msg *lite.MidiMessage) error {
	if inst == nil || msg == nil {
		return nil
	}

	inst.dispatchMu.Lock()
	defer inst.dispatchMu.Unlock()

	if msg.NoteOff != nil {
		if err := inst.noteOff(msg.NoteOff); err != nil {
			return err
		}
	}
	if msg.NoteOn != nil {
		if msg.NoteOn.Velocity == 0 {
			return inst.noteOff(msg.NoteOn)
		}
		return inst.noteOn(msg.NoteOn)
	}
	return nil
}

// GetInfo implements [lite.Instrument].
func (inst *SimpleSynth) GetInfo() *lite.InstrumentInfo {
	if inst == nil {
		return nil
	}
	info := inst.info
	return &info
}

func (inst *SimpleSynth) noteOn(note *lite.MidiNote) error {
	if note == nil {
		return nil
	}
	if inst.audio == nil {
		return fmt.Errorf("simplesynth: audio context is nil")
	}
	if inst.destination == nil {
		return fmt.Errorf("simplesynth: audio destination is nil")
	}

	now := inst.audio.GetCurrentTime()

	inst.mu.Lock()
	previous := inst.voices[note.Key]
	delete(inst.voices, note.Key)
	inst.mu.Unlock()
	if previous != nil {
		previous.source.Stop(now)
		previous.source.Disconnect()
		previous.gain.Disconnect()
	}

	voice, err := inst.createVoice(note, now)
	if err != nil {
		return err
	}

	inst.mu.Lock()
	inst.voices[note.Key] = voice
	inst.mu.Unlock()

	when := inst.noteTime(note)
	voice.source.Start(when)
	return nil
}

func (inst *SimpleSynth) createVoice(note *lite.MidiNote, now lite.Seconds) (*simpleSynthVoice, error) {
	gain := inst.audio.CreateGain()
	if gain == nil {
		return nil, fmt.Errorf("simplesynth: CreateGain returned nil")
	}

	var source lite.AudioScheduledSourceNode
	if inst.waveform == "noise" {
		bufferSource := inst.audio.CreateBufferSource()
		if bufferSource == nil {
			return nil, fmt.Errorf("simplesynth: CreateBufferSource returned nil")
		}
		if err := inst.prepareNoise(bufferSource); err != nil {
			return nil, err
		}
		source = bufferSource
	} else {
		oscillator := inst.audio.CreateOscillator()
		if oscillator == nil {
			return nil, fmt.Errorf("simplesynth: CreateOscillator returned nil")
		}
		oscillator.SetType(simpleSynthOscillatorType(inst.waveform))

		var pitch lite.MidiKeyInfo
		note.Key.GetInfo(&pitch)
		oscillator.GetFrequency().SetValueAtTime(float64(pitch.Frequency), now)
		source = oscillator
	}

	level := float64(note.Velocity) / 256 // 127
	if level > 1 {
		level = 1
	}
	gain.GetGain().SetValueAtTime(level, now)

	source.Connect(gain)
	gain.Connect(inst.destination)

	voice := &simpleSynthVoice{
		source: source,
		gain:   gain,
	}
	key := note.Key
	source.SetOnEnded(func() {
		inst.voiceEnded(key, voice)
	})
	return voice, nil
}

func (inst *SimpleSynth) noteOff(note *lite.MidiNote) error {
	if note == nil {
		return nil
	}

	inst.mu.Lock()
	voice := inst.voices[note.Key]
	delete(inst.voices, note.Key)
	inst.mu.Unlock()
	if voice == nil {
		return nil
	}

	when := inst.noteTime(note)
	voice.source.Stop(when)
	if when <= inst.audio.GetCurrentTime() {
		voice.source.Disconnect()
		voice.gain.Disconnect()
	}
	return nil
}

func (inst *SimpleSynth) voiceEnded(key lite.MidiNoteNumber, voice *simpleSynthVoice) {
	inst.mu.Lock()
	if current := inst.voices[key]; current == voice {
		delete(inst.voices, key)
	}
	inst.mu.Unlock()

	voice.source.Disconnect()
	voice.gain.Disconnect()
}

func (inst *SimpleSynth) noteTime(note *lite.MidiNote) lite.Seconds {
	if note.At == nil {
		return inst.audio.GetCurrentTime()
	}
	position := &lite.RelativePosition{
		At:       note.At,
		Position: note.Position,
	}
	return lite.Seconds(position.Absolute())
}

func (inst *SimpleSynth) prepareNoise(source lite.AudioBufferSourceNode) error {
	sampleRate := inst.audio.GetSampleRate()
	if sampleRate == 0 {
		sampleRate = 44100
	}
	length := lite.SampleCount(sampleRate) * simpleSynthNoiseSeconds
	buffer := inst.audio.CreateBuffer(1, length, sampleRate)
	if buffer == nil {
		return fmt.Errorf("simplesynth: CreateBuffer returned nil")
	}

	state := uint32(0x9e3779b9)
	data := buffer.GetChannelData(0)
	for i := range data {
		state = state*1664525 + 1013904223
		data[i] = lite.SampleValue(float64(state)/float64(math.MaxUint32)*2 - 1)
	}
	source.SetBuffer(buffer)
	return nil
}

func simpleSynthWaveform(preset string) string {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "triangle":
		return "triangle"
	case "saw", "sawtooth":
		return "sawtooth"
	case "square":
		return "square"
	case "noise":
		return "noise"
	default:
		return "sine"
	}
}

func simpleSynthOscillatorType(waveform string) lite.OscillatorType {
	switch waveform {
	case "triangle":
		return lite.OscillatorTypeTriangle
	case "sawtooth":
		return lite.OscillatorTypeSawtooth
	case "square":
		return lite.OscillatorTypeSquare
	default:
		return lite.OscillatorTypeSine
	}
}
