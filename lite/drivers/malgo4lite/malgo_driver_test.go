package malgo4lite

import (
	"math"
	"testing"
	"time"
	"unsafe"

	"github.com/ak47less/step1go/lite"
)

func newRunningTestContext(t *testing.T) *MalgoAudioContext {
	t.Helper()

	ctx := newMalgoAudioContext(nil)
	ctx.state = lite.AudioContextStateRunning
	ctx.runningSince = time.Now()
	t.Cleanup(ctx.Close)
	return ctx
}

func renderTestFrames(t *testing.T, ctx *MalgoAudioContext, frameCount int) []float32 {
	t.Helper()

	output := make([]byte, frameCount*malgoDefaultChannelCount*4)
	ctx.onData(output, nil, uint32(frameCount))
	return unsafe.Slice(
		(*float32)(unsafe.Pointer(unsafe.SliceData(output))),
		len(output)/4,
	)
}

func requireFloat32Near(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 1e-6 {
		t.Fatalf("got %f, want %f", got, want)
	}
}

func TestMalgoDriverContextLifecycle(t *testing.T) {
	var driver lite.Driver = new(MalgoDriver)
	factory := driver.GetAudioContextFactory()
	if factory == nil {
		t.Fatal("GetAudioContextFactory returned nil")
	}

	ctx := factory.NewAudioContext()
	if ctx == nil {
		t.Fatal("NewAudioContext returned nil")
	}
	defer ctx.Close()

	if state := ctx.GetState(); state != lite.AudioContextStateSuspended {
		t.Fatalf("initial state = %q, want %q", state, lite.AudioContextStateSuspended)
	}
	if sampleRate := ctx.GetSampleRate(); sampleRate != malgoDefaultSampleRate {
		t.Fatalf("sample rate = %d, want %d", sampleRate, malgoDefaultSampleRate)
	}
	if destination := ctx.GetDestination(); destination == nil {
		t.Fatal("GetDestination returned nil")
	} else if count := destination.GetMaxChannelCount(); count != malgoDefaultChannelCount {
		t.Fatalf("max channel count = %d, want %d", count, malgoDefaultChannelCount)
	}

	if oscillator := ctx.CreateOscillator(); oscillator == nil {
		t.Fatal("CreateOscillator returned nil")
	}
	if gain := ctx.CreateGain(); gain == nil {
		t.Fatal("CreateGain returned nil")
	}
	if buffer := ctx.CreateBuffer(2, 8, 0); buffer == nil {
		t.Fatal("CreateBuffer returned nil")
	}
	if source := ctx.CreateBufferSource(); source == nil {
		t.Fatal("CreateBufferSource returned nil")
	}

	ctx.Close()
	if state := ctx.GetState(); state != lite.AudioContextStateClosed {
		t.Fatalf("state after Close = %q, want %q", state, lite.AudioContextStateClosed)
	}

	// Close must be idempotent.
	ctx.Close()
	ctx.Suspend()
	if state := ctx.GetState(); state != lite.AudioContextStateClosed {
		t.Fatalf("state after repeated Close/Suspend = %q, want %q", state, lite.AudioContextStateClosed)
	}
}

func TestMalgoAudioDefaults(t *testing.T) {
	ctx := newMalgoAudioContext(nil)
	defer ctx.Close()

	oscillator := ctx.CreateOscillator()
	if oscillator.GetType() != lite.OscillatorTypeSine {
		t.Fatalf("oscillator type = %q, want %q", oscillator.GetType(), lite.OscillatorTypeSine)
	}
	if frequency := oscillator.GetFrequency().GetValue(); frequency != malgoDefaultFrequency {
		t.Fatalf("frequency = %f, want %f", frequency, malgoDefaultFrequency)
	}
	if detune := oscillator.GetDetune().GetValue(); detune != 0 {
		t.Fatalf("detune = %f, want 0", detune)
	}

	oscillator.SetType(lite.OscillatorTypeTriangle)
	if oscillator.GetType() != lite.OscillatorTypeTriangle {
		t.Fatalf("oscillator type = %q, want %q", oscillator.GetType(), lite.OscillatorTypeTriangle)
	}

	gain := ctx.CreateGain()
	if value := gain.GetGain().GetValue(); value != malgoDefaultGain {
		t.Fatalf("gain = %f, want %f", value, malgoDefaultGain)
	}

	source := ctx.CreateBufferSource()
	if source.GetBuffer() != nil {
		t.Fatal("new buffer source should not have a buffer")
	}
	if rate := source.GetPlaybackRate().GetValue(); rate != malgoDefaultPlaybackRate {
		t.Fatalf("playback rate = %f, want %f", rate, malgoDefaultPlaybackRate)
	}
	if detune := source.GetDetune().GetValue(); detune != 0 {
		t.Fatalf("buffer source detune = %f, want 0", detune)
	}
}

func TestMalgoCreateBuffer(t *testing.T) {
	ctx := newMalgoAudioContext(nil)
	defer ctx.Close()

	buffer := ctx.CreateBuffer(2, 4, 0)
	if got := buffer.GetNumberOfChannels(); got != 2 {
		t.Fatalf("channel count = %d, want 2", got)
	}
	if got := buffer.GetLength(); got != 4 {
		t.Fatalf("length = %d, want 4", got)
	}
	if got := buffer.GetSampleRate(); got != malgoDefaultSampleRate {
		t.Fatalf("sample rate = %d, want %d", got, malgoDefaultSampleRate)
	}
	if buffer.GetChannelData(-1) != nil || buffer.GetChannelData(2) != nil {
		t.Fatal("out-of-range channel data should be nil")
	}

	left := buffer.GetChannelData(0)
	left[1] = 0.5
	requireFloat32Near(t, float32(buffer.GetChannelData(0)[1]), 0.5)

	empty := ctx.CreateBuffer(-1, 4, 48000)
	if got := empty.GetNumberOfChannels(); got != 0 {
		t.Fatalf("negative channel count produced %d channels", got)
	}
	if got := empty.GetLength(); got != 0 {
		t.Fatalf("zero-channel buffer length = %d, want 0", got)
	}
	if got := empty.GetSampleRate(); got != 48000 {
		t.Fatalf("empty buffer sample rate = %d, want 48000", got)
	}
}

func TestMalgoOscillatorGainGraph(t *testing.T) {
	ctx := newRunningTestContext(t)
	oscillator := ctx.CreateOscillator()
	gain := ctx.CreateGain()

	oscillator.SetType(lite.OscillatorTypeSquare)
	oscillator.GetFrequency().SetValue(10)
	gain.GetGain().SetValue(0.25)

	// Duplicate connections should not double the signal.
	oscillator.Connect(gain)
	oscillator.Connect(gain)
	gain.Connect(ctx.GetDestination())
	oscillator.Start(0)

	output := renderTestFrames(t, ctx, 256)
	for frame := 0; frame < 256; frame++ {
		left := output[frame*malgoDefaultChannelCount]
		right := output[frame*malgoDefaultChannelCount+1]
		requireFloat32Near(t, left, 0.25)
		requireFloat32Near(t, right, left)
	}

	oscillator.Disconnect()
	output = renderTestFrames(t, ctx, 256)
	for _, sample := range output {
		requireFloat32Near(t, sample, 0)
	}
}

func TestMalgoBufferSourcePlayback(t *testing.T) {
	ctx := newRunningTestContext(t)
	buffer := ctx.CreateBuffer(1, 4, ctx.GetSampleRate())
	copy(buffer.GetChannelData(0), []lite.SampleValue{0.25, 0.5, -0.25, -0.5})

	source := ctx.CreateBufferSource()
	source.SetBuffer(buffer)
	if source.GetBuffer() != buffer {
		t.Fatal("GetBuffer did not return the buffer passed to SetBuffer")
	}
	source.Connect(ctx.GetDestination())

	ended := make(chan struct{}, 1)
	source.SetOnEnded(func() {
		ended <- struct{}{}
	})
	source.Start(0)

	output := renderTestFrames(t, ctx, 512)
	want := []float32{0.25, 0.25, 0.5, 0.5, -0.25, -0.25, -0.5, -0.5}
	for i, value := range want {
		requireFloat32Near(t, output[i], value)
	}
	for _, sample := range output[len(want):] {
		requireFloat32Near(t, sample, 0)
	}

	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("onended callback was not called")
	}
}

func TestMalgoDestinationPreventsPolyphonicClipping(t *testing.T) {
	ctx := newRunningTestContext(t)
	frameCount := 256
	pattern := []lite.SampleValue{0.75, 0.375, 0, -0.375, -0.75}
	buffer := ctx.CreateBuffer(1, lite.SampleCount(frameCount), ctx.GetSampleRate())
	for i := range buffer.GetChannelData(0) {
		buffer.GetChannelData(0)[i] = pattern[i%len(pattern)]
	}

	sources := make([]lite.AudioBufferSourceNode, malgoDefaultChannelCount)
	for i := range sources {
		sources[i] = ctx.CreateBufferSource()
		sources[i].SetBuffer(buffer)
		sources[i].Connect(ctx.GetDestination())
		sources[i].Start(0)
	}

	output := renderTestFrames(t, ctx, frameCount)
	busGain := float64(malgoBusCeiling) / 1.5
	for frame := 0; frame < frameCount; frame++ {
		want := float32(float64(pattern[frame%len(pattern)]) * 2 * busGain)
		for channel := 0; channel < malgoDefaultChannelCount; channel++ {
			sample := output[frame*malgoDefaultChannelCount+channel]
			requireFloat32Near(t, sample, want)
			if math.Abs(float64(sample)) > float64(malgoBusCeiling) {
				t.Fatalf("frame %d channel %d = %f exceeds bus ceiling %f", frame, channel, sample, malgoBusCeiling)
			}
		}
	}
}

func TestMalgoDestinationLeavesNormalSignalUnchanged(t *testing.T) {
	ctx := newRunningTestContext(t)
	buffer := ctx.CreateBuffer(1, 8, ctx.GetSampleRate())
	copy(buffer.GetChannelData(0), []lite.SampleValue{0.1, -0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8})

	source := ctx.CreateBufferSource()
	source.SetBuffer(buffer)
	source.Connect(ctx.GetDestination())
	source.Start(0)

	output := renderTestFrames(t, ctx, 8)
	for frame, value := range buffer.GetChannelData(0) {
		requireFloat32Near(t, output[frame*malgoDefaultChannelCount], float32(value))
		requireFloat32Near(t, output[frame*malgoDefaultChannelCount+1], float32(value))
	}
}

func TestMalgoScheduledSourceStop(t *testing.T) {
	ctx := newRunningTestContext(t)
	source := ctx.CreateBufferSource()
	buffer := ctx.CreateBuffer(1, 4, ctx.GetSampleRate())
	copy(buffer.GetChannelData(0), []lite.SampleValue{0.5, 0.5, 0.5, 0.5})
	source.SetBuffer(buffer)
	source.Connect(ctx.GetDestination())

	ended := make(chan struct{}, 1)
	source.SetOnEnded(func() {
		ended <- struct{}{}
	})
	source.Start(0)
	source.Stop(0)

	output := renderTestFrames(t, ctx, 64)
	requireFloat32Near(t, output[0], 0.5)
	for _, sample := range output[malgoDefaultChannelCount:] {
		requireFloat32Near(t, sample, 0)
	}

	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("onended callback was not called after Stop")
	}
}

func TestMalgoAudioParamAutomation(t *testing.T) {
	now := lite.Seconds(0)
	param := newMalgoAudioParam(1, 0, func() lite.Seconds {
		return now
	})

	param.SetValueAtTime(2, 1)
	param.LinearRampToValueAtTime(4, 3)
	now = 2
	if value := param.GetValue(); math.Abs(value-3) > 1e-6 {
		t.Fatalf("linear ramp value = %f, want 3", value)
	}

	param.SetValue(5)
	if value := param.GetValue(); value != 5 {
		t.Fatalf("SetValue = %f, want 5", value)
	}

	now = 0
	exponential := newMalgoAudioParam(1, 0, func() lite.Seconds {
		return now
	})
	exponential.ExponentialRampToValueAtTime(4, 2)
	now = 1
	if value := exponential.GetValue(); math.Abs(value-2) > 1e-6 {
		t.Fatalf("exponential ramp value = %f, want 2", value)
	}
}

func TestOscillatorWaveforms(t *testing.T) {
	tests := []struct {
		name  string
		kind  lite.OscillatorType
		phase float64
		want  float32
	}{
		{name: "sine", kind: lite.OscillatorTypeSine, phase: 0.25, want: 1},
		{name: "square", kind: lite.OscillatorTypeSquare, phase: 0.25, want: 1},
		{name: "sawtooth", kind: lite.OscillatorTypeSawtooth, phase: 0.25, want: -0.5},
		{name: "triangle", kind: lite.OscillatorTypeTriangle, phase: 0.25, want: 0},
		{name: "custom", kind: lite.OscillatorTypeCustom, phase: 0.25, want: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireFloat32Near(t, float32(oscillatorSample(test.kind, test.phase)), test.want)
		})
	}
}
