package mock

import (
	"math"
	"testing"
	"time"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

func newRunningMockContext(t *testing.T) *innerMockAudioContext {
	t.Helper()

	ctx := newInnerMockAudioContext()
	ctx.Resume()
	t.Cleanup(ctx.Close)
	return ctx
}

func requireMockFloatNear(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("got %f, want %f", got, want)
	}
}

func TestMockDriverLifecycle(t *testing.T) {
	var driver webaudioapi.Driver = NewMockDriver()
	factory := driver.GetFactory()
	if factory == nil {
		t.Fatal("GetFactory returned nil")
	}

	ctx, err := factory.CreateAudioContext()
	if err != nil {
		t.Fatal(err)
	}
	if ctx == nil {
		t.Fatal("CreateAudioContext returned nil")
	}
	defer ctx.Close()

	if state := ctx.GetState(); state != webaudioapi.AudioContextStateSuspended {
		t.Fatalf("initial state = %q, want %q", state, webaudioapi.AudioContextStateSuspended)
	}
	if sampleRate := ctx.GetSampleRate(); sampleRate != mockDefaultSampleRate {
		t.Fatalf("sample rate = %v, want %v", sampleRate, mockDefaultSampleRate)
	}
	destination := ctx.GetDestination()
	if destination == nil {
		t.Fatal("GetDestination returned nil")
	}
	if channels := destination.GetMaxChannelCount(); channels != mockDefaultChannelCount {
		t.Fatalf("max channel count = %d, want %d", channels, mockDefaultChannelCount)
	}

	ctx.Resume()
	if state := ctx.GetState(); state != webaudioapi.AudioContextStateRunning {
		t.Fatalf("state after Resume = %q, want %q", state, webaudioapi.AudioContextStateRunning)
	}

	before := ctx.GetCurrentTime()
	time.Sleep(5 * time.Millisecond)
	if after := ctx.GetCurrentTime(); after <= before {
		t.Fatalf("running clock did not advance: before=%f after=%f", before, after)
	}

	ctx.Suspend()
	if state := ctx.GetState(); state != webaudioapi.AudioContextStateSuspended {
		t.Fatalf("state after Suspend = %q, want %q", state, webaudioapi.AudioContextStateSuspended)
	}
	frozen := ctx.GetCurrentTime()
	time.Sleep(5 * time.Millisecond)
	if got := ctx.GetCurrentTime(); got != frozen {
		t.Fatalf("suspended clock changed from %f to %f", frozen, got)
	}

	ctx.Close()
	if state := ctx.GetState(); state != webaudioapi.AudioContextStateClosed {
		t.Fatalf("state after Close = %q, want %q", state, webaudioapi.AudioContextStateClosed)
	}

	// Closed is terminal and Close is idempotent.
	ctx.Close()
	ctx.Resume()
	ctx.Suspend()
	if state := ctx.GetState(); state != webaudioapi.AudioContextStateClosed {
		t.Fatalf("closed context changed state to %q", state)
	}
}

func TestMockAudioDefaults(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	oscillator := ctx.CreateOscillator()
	if got := oscillator.GetType(); got != webaudioapi.OscillatorTypeSine {
		t.Fatalf("oscillator type = %q, want %q", got, webaudioapi.OscillatorTypeSine)
	}
	if got := oscillator.GetFrequency().GetValue(); got != mockDefaultFrequency {
		t.Fatalf("frequency = %f, want %f", got, mockDefaultFrequency)
	}
	if got := oscillator.GetDetune().GetValue(); got != 0 {
		t.Fatalf("oscillator detune = %f, want 0", got)
	}

	oscillator.SetType(webaudioapi.OscillatorTypeSquare)
	if got := oscillator.GetType(); got != webaudioapi.OscillatorTypeSquare {
		t.Fatalf("oscillator type = %q, want %q", got, webaudioapi.OscillatorTypeSquare)
	}

	gain := ctx.CreateGain()
	if got := gain.GetGain().GetValue(); got != mockDefaultGain {
		t.Fatalf("gain = %f, want %f", got, mockDefaultGain)
	}

	source := ctx.CreateBufferSource()
	if source.GetBuffer() != nil {
		t.Fatal("new buffer source should not have a buffer")
	}
	if got := source.GetPlaybackRate().GetValue(); got != mockDefaultPlaybackRate {
		t.Fatalf("playback rate = %f, want %f", got, mockDefaultPlaybackRate)
	}
	if got := source.GetDetune().GetValue(); got != 0 {
		t.Fatalf("buffer source detune = %f, want 0", got)
	}
}

func TestMockCreateBuffer(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	buffer := ctx.CreateBuffer(2, 8, 0)
	if got := buffer.GetNumberOfChannels(); got != 2 {
		t.Fatalf("channel count = %d, want 2", got)
	}
	if got := buffer.GetLength(); got != 8 {
		t.Fatalf("length = %d, want 8", got)
	}
	if got := buffer.GetSampleRate(); got != mockDefaultSampleRate {
		t.Fatalf("sample rate = %v, want %v", got, mockDefaultSampleRate)
	}
	if buffer.GetChannelData(-1) != nil || buffer.GetChannelData(2) != nil {
		t.Fatal("out-of-range channel data should be nil")
	}

	buffer.GetChannelData(0)[3] = 0.75
	if got := buffer.GetChannelData(0)[3]; got != 0.75 {
		t.Fatalf("channel sample = %f, want 0.75", got)
	}

	empty := ctx.CreateBuffer(-1, 4, 48000)
	if got := empty.GetNumberOfChannels(); got != 0 {
		t.Fatalf("negative channel count produced %d channels", got)
	}
	if got := empty.GetLength(); got != 0 {
		t.Fatalf("zero-channel buffer length = %d, want 0", got)
	}
	if got := empty.GetSampleRate(); got != 48000 {
		t.Fatalf("empty buffer sample rate = %v, want 48000", got)
	}
}

func TestMockAudioNodeConnections(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	oscillator := ctx.CreateOscillator().(*innerMockOscillatorNode)
	gain := ctx.CreateGain().(*innerMockGainNode)
	destination := ctx.GetDestination()

	oscillator.Connect(gain)
	oscillator.Connect(gain)
	oscillator.Connect(nil)
	gain.Connect(destination)

	if got := len(oscillator.getConnections()); got != 2 {
		t.Fatalf("oscillator connections = %d, want 2", got)
	}
	if got := oscillator.getConnections()[0]; got != gain {
		t.Fatalf("oscillator connection = %T, want %T", got, gain)
	}
	if got := len(gain.getConnections()); got != 1 {
		t.Fatalf("gain connections = %d, want 1", got)
	}
	if got := gain.getConnections()[0]; got != destination {
		t.Fatalf("gain connection = %T, want %T", got, destination)
	}

	oscillator.Disconnect()
	if got := len(oscillator.getConnections()); got != 0 {
		t.Fatalf("connections after Disconnect = %d, want 0", got)
	}
	if got := len(gain.getConnections()); got != 1 {
		t.Fatalf("Disconnect changed unrelated node connections: %d", got)
	}
}

func TestMockScheduledSource(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	source := ctx.CreateBufferSource()
	ended := make(chan struct{}, 1)
	source.SetOnEnded(func() {
		ended <- struct{}{}
	})

	// Scheduling while suspended must not complete until the context runs.
	source.Start(0)
	source.Stop(0)
	select {
	case <-ended:
		t.Fatal("source ended while context was suspended")
	case <-time.After(5 * time.Millisecond):
	}

	ctx.Resume()
	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("onended callback was not called after Resume")
	}

	// Registering a callback after completion invokes it immediately.
	called := false
	source.SetOnEnded(func() {
		called = true
	})
	if !called {
		t.Fatal("late onended callback was not called")
	}
}

func TestMockCloseFinishesRunningSource(t *testing.T) {
	ctx := newRunningMockContext(t)
	source := ctx.CreateOscillator()

	ended := make(chan struct{}, 1)
	source.SetOnEnded(func() {
		ended <- struct{}{}
	})
	source.Start(0)
	ctx.Close()

	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("running source did not finish when context closed")
	}
}

func TestMockAudioParamAutomation(t *testing.T) {
	now := webaudioapi.Seconds(0)
	param := newInnerMockAudioParam(1, 0, func() webaudioapi.Seconds {
		return now
	})

	param.SetValueAtTime(2, 1)
	param.LinearRampToValueAtTime(4, 3)
	now = 2
	requireMockFloatNear(t, param.GetValue(), 3)

	param.SetValue(5)
	requireMockFloatNear(t, param.GetValue(), 5)

	now = 0
	exponential := newInnerMockAudioParam(1, 0, func() webaudioapi.Seconds {
		return now
	})
	exponential.ExponentialRampToValueAtTime(4, 2)
	now = 1
	requireMockFloatNear(t, exponential.GetValue(), 2)

	target := newInnerMockAudioParam(1, 0, func() webaudioapi.Seconds {
		return now
	})
	target.SetTargetAtTime(0, 0, 0.5)
	now = 0.5
	requireMockFloatNear(t, target.GetValue(), math.Exp(-1))
}

func TestMockAudioParamCancelScheduledValues(t *testing.T) {
	now := webaudioapi.Seconds(1)
	param := newInnerMockAudioParam(1, 0, func() webaudioapi.Seconds {
		return now
	})

	param.SetValueAtTime(2, 2)
	param.SetValueAtTime(3, 3)
	param.SetValueAtTime(4, 4)
	param.CancelScheduledValues(3)

	now = 2
	requireMockFloatNear(t, param.GetValue(), 2)

	now = 3
	requireMockFloatNear(t, param.GetValue(), 2)

	now = 4
	requireMockFloatNear(t, param.GetValue(), 2)
}

func TestNormalizeMockTime(t *testing.T) {
	tests := []struct {
		name string
		in   webaudioapi.Seconds
		want webaudioapi.Seconds
	}{
		{name: "positive", in: 1.5, want: 1.5},
		{name: "zero", in: 0, want: 0},
		{name: "negative", in: -1, want: 0},
		{name: "nan", in: webaudioapi.Seconds(math.NaN()), want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeMockTime(test.in); got != test.want {
				t.Fatalf("normalizeMockTime(%v) = %v, want %v", test.in, got, test.want)
			}
		})
	}
}
