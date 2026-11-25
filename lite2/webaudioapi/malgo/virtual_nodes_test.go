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

func TestMalgoScriptProcessorAudioProcess(t *testing.T) {
	ctx := newRunningTestContext(t)
	source := ctx.CreateConstantSource()
	processor := ctx.CreateScriptProcessor()

	source.GetOffset().SetValue(0.25)
	source.Connect(processor)
	processor.Connect(ctx.GetDestination())

	callbackCount := 0
	processor.SetOnAudioProcess(func(event webaudioapi.AudioProcessingEvent) {
		callbackCount++
		input := event.GetInputBuffer()
		output := event.GetOutputBuffer()
		if input == nil || output == nil {
			t.Fatal("audioprocess event buffers should not be nil")
		}
		if input.GetLength() != webaudioapi.SampleCount(processor.GetBufferSize()) {
			t.Fatalf("input length = %d, want %d", input.GetLength(), processor.GetBufferSize())
		}
		if output.GetLength() != webaudioapi.SampleCount(processor.GetBufferSize()) {
			t.Fatalf("output length = %d, want %d", output.GetLength(), processor.GetBufferSize())
		}
		if event.GetPlaybackTime() <= 0 {
			t.Fatalf("playback time = %f, want a positive value", event.GetPlaybackTime())
		}
		for channel := 0; channel < input.GetNumberOfChannels(); channel++ {
			for frame, value := range input.GetChannelData(channel) {
				requireFloat32Near(t, float32(value), 0.25)
				output.GetChannelData(channel)[frame] = 0.5
			}
		}
	})
	if processor.GetOnAudioProcess() == nil {
		t.Fatal("SetOnAudioProcess did not store the handler")
	}

	source.Start(0)
	bufferSize := processor.GetBufferSize()
	output := renderTestFrames(t, ctx, bufferSize*2)
	if callbackCount != 2 {
		t.Fatalf("audioprocess callback count = %d, want 2", callbackCount)
	}
	for frame := 0; frame < bufferSize; frame++ {
		requireFloat32Near(t, output[frame*2], 0)
		requireFloat32Near(t, output[frame*2+1], 0)
	}
	for frame := bufferSize; frame < bufferSize*2; frame++ {
		requireFloat32Near(t, output[frame*2], 0.5)
		requireFloat32Near(t, output[frame*2+1], 0.5)
	}

	processor.SetOnAudioProcess(nil)
	if processor.GetOnAudioProcess() != nil {
		t.Fatal("SetOnAudioProcess(nil) did not clear the handler")
	}
}
