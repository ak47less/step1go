package malgo

import (
	"testing"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

func TestMalgoAdditionalNodeCreators(t *testing.T) {
	ctx := newInnerMalgoAudioContext(nil)
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

func TestMalgoAdditionalNodeDefaults(t *testing.T) {
	ctx := newInnerMalgoAudioContext(nil)
	defer ctx.Close()

	biquad := ctx.CreateBiquadFilter()
	if got := biquad.GetType(); got != webaudioapi.BiquadFilterTypeLowpass {
		t.Fatalf("biquad type = %q, want %q", got, webaudioapi.BiquadFilterTypeLowpass)
	}
	if got := biquad.GetFrequency().GetValue(); got != malgoDefaultBiquadFrequency {
		t.Fatalf("biquad frequency = %f, want %f", got, malgoDefaultBiquadFrequency)
	}
	if got := ctx.CreateDelay().GetDelayTime().GetValue(); got != malgoDefaultDelayTime {
		t.Fatalf("delay time = %f, want %f", got, malgoDefaultDelayTime)
	}

	compressor := ctx.CreateDynamicsCompressor()
	if got := compressor.GetThreshold().GetValue(); got != malgoDefaultCompressorThreshold {
		t.Fatalf("compressor threshold = %f, want %f", got, malgoDefaultCompressorThreshold)
	}
	if got := compressor.GetRatio().GetValue(); got != malgoDefaultCompressorRatio {
		t.Fatalf("compressor ratio = %f, want %f", got, malgoDefaultCompressorRatio)
	}

	analyser := ctx.CreateAnalyser()
	if got := analyser.GetFFTSize(); got != malgoDefaultAnalyserFFTSize {
		t.Fatalf("analyser FFT size = %d, want %d", got, malgoDefaultAnalyserFFTSize)
	}
	if got := ctx.CreateChannelSplitter().GetNumberOfOutputs(); got != malgoDefaultChannelSplitterOutputs {
		t.Fatalf("channel splitter outputs = %d, want %d", got, malgoDefaultChannelSplitterOutputs)
	}
	if got := ctx.CreateChannelMerger().GetNumberOfInputs(); got != malgoDefaultChannelMergerInputs {
		t.Fatalf("channel merger inputs = %d, want %d", got, malgoDefaultChannelMergerInputs)
	}
	if got := ctx.CreateConstantSource().GetOffset().GetValue(); got != malgoDefaultConstantSourceOffset {
		t.Fatalf("constant source offset = %f, want %f", got, malgoDefaultConstantSourceOffset)
	}
	if got := ctx.CreateScriptProcessor().GetBufferSize(); got != malgoDefaultScriptProcessorCapacity {
		t.Fatalf("script processor buffer size = %d, want %d", got, malgoDefaultScriptProcessorCapacity)
	}
}

func TestMalgoConstantSourceWaveShaperGraph(t *testing.T) {
	ctx := newRunningTestContext(t)
	source := ctx.CreateConstantSource()
	shaper := ctx.CreateWaveShaper()
	analyser := ctx.CreateAnalyser()

	source.GetOffset().SetValue(0.25)
	shaper.SetCurve([]webaudioapi.SampleValue{-0.5, 0.5})
	source.Connect(shaper)
	shaper.Connect(analyser)
	analyser.Connect(ctx.GetDestination())
	source.Start(0)

	output := renderTestFrames(t, ctx, 8)
	for frame := 0; frame < 8; frame++ {
		requireFloat32Near(t, output[frame*2], 0.125)
		requireFloat32Near(t, output[frame*2+1], 0.125)
	}
}
