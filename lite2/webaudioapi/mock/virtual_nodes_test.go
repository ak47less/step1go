package mock

import (
	"testing"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

func TestMockAdditionalNodeCreators(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	nodes := []webaudioapi.AudioNode{
		ctx.CreateBiquadFilter(),
		ctx.CreateIIRFilter(),
		ctx.CreateDelay(),
		ctx.CreateDynamicsCompressor(),
		ctx.CreateAnalyser(),
		ctx.CreatePanner(),
		ctx.CreateChannelSplitter(),
		ctx.CreateChannelMerger(),
		ctx.CreateConstantSource(),
		ctx.CreateWaveShaper(),
		ctx.CreateScriptProcessor(),
	}
	for i, node := range nodes {
		if node == nil {
			t.Fatalf("node creator %d returned nil", i)
		}
		node.Connect(ctx.GetDestination())
	}
	if ctx.CreatePeriodicWave() == nil {
		t.Fatal("CreatePeriodicWave returned nil")
	}
}

func TestMockAdditionalNodeDefaults(t *testing.T) {
	ctx := newInnerMockAudioContext()
	defer ctx.Close()

	biquad := ctx.CreateBiquadFilter()
	if got := biquad.GetType(); got != webaudioapi.BiquadFilterTypeLowpass {
		t.Fatalf("biquad type = %q, want %q", got, webaudioapi.BiquadFilterTypeLowpass)
	}
	requireMockFloatNear(t, biquad.GetFrequency().GetValue(), mockDefaultBiquadFrequency)
	requireMockFloatNear(t, biquad.GetQ().GetValue(), mockDefaultBiquadQ)

	requireMockFloatNear(t, ctx.CreateDelay().GetDelayTime().GetValue(), mockDefaultDelayTime)

	compressor := ctx.CreateDynamicsCompressor()
	requireMockFloatNear(t, compressor.GetThreshold().GetValue(), mockDefaultCompressorThreshold)
	requireMockFloatNear(t, compressor.GetKnee().GetValue(), mockDefaultCompressorKnee)
	requireMockFloatNear(t, compressor.GetRatio().GetValue(), mockDefaultCompressorRatio)
	requireMockFloatNear(t, compressor.GetAttack().GetValue(), mockDefaultCompressorAttack)
	requireMockFloatNear(t, compressor.GetRelease().GetValue(), mockDefaultCompressorRelease)

	analyser := ctx.CreateAnalyser()
	if got := analyser.GetFFTSize(); got != mockDefaultAnalyserFFTSize {
		t.Fatalf("analyser FFT size = %d, want %d", got, mockDefaultAnalyserFFTSize)
	}
	if got := analyser.GetFrequencyBinCount(); got != mockDefaultAnalyserFFTSize/2 {
		t.Fatalf("analyser frequency bin count = %d, want %d", got, mockDefaultAnalyserFFTSize/2)
	}
	requireMockFloatNear(t, analyser.GetMinDecibels(), mockDefaultAnalyserMinDecibels)
	requireMockFloatNear(t, analyser.GetMaxDecibels(), mockDefaultAnalyserMaxDecibels)
	requireMockFloatNear(t, analyser.GetSmoothingTimeConstant(), mockDefaultAnalyserSmoothing)

	panner := ctx.CreatePanner()
	if got := panner.GetPanningModel(); got != webaudioapi.PanningModelTypeEqualPower {
		t.Fatalf("panning model = %q, want %q", got, webaudioapi.PanningModelTypeEqualPower)
	}
	if got := panner.GetDistanceModel(); got != webaudioapi.DistanceModelTypeInverse {
		t.Fatalf("distance model = %q, want %q", got, webaudioapi.DistanceModelTypeInverse)
	}

	if got := ctx.CreateChannelSplitter().GetNumberOfOutputs(); got != mockDefaultChannelSplitterOutputs {
		t.Fatalf("channel splitter outputs = %d, want %d", got, mockDefaultChannelSplitterOutputs)
	}
	if got := ctx.CreateChannelMerger().GetNumberOfInputs(); got != mockDefaultChannelMergerInputs {
		t.Fatalf("channel merger inputs = %d, want %d", got, mockDefaultChannelMergerInputs)
	}
	if got := ctx.CreateConstantSource().GetOffset().GetValue(); got != mockDefaultConstantSourceOffset {
		t.Fatalf("constant source offset = %f, want %f", got, mockDefaultConstantSourceOffset)
	}
	if got := ctx.CreateWaveShaper().GetOversample(); got != webaudioapi.OverSampleTypeNone {
		t.Fatalf("wave shaper oversample = %q, want %q", got, webaudioapi.OverSampleTypeNone)
	}
	if got := ctx.CreateScriptProcessor().GetBufferSize(); got != mockDefaultScriptProcessorCapacity {
		t.Fatalf("script processor buffer size = %d, want %d", got, mockDefaultScriptProcessorCapacity)
	}
}
