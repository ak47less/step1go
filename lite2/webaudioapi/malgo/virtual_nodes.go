package malgo

import (
	"math"
	"sync"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

const (
	malgoDefaultBiquadFrequency         = 350.0
	malgoDefaultBiquadQ                 = 1.0
	malgoDefaultDelayTime               = 0.0
	malgoDefaultDelayCapacitySeconds    = 3.0
	malgoDefaultCompressorThreshold     = -24.0
	malgoDefaultCompressorKnee          = 30.0
	malgoDefaultCompressorRatio         = 12.0
	malgoDefaultCompressorAttack        = 0.003
	malgoDefaultCompressorRelease       = 0.25
	malgoDefaultAnalyserFFTSize         = 2048
	malgoDefaultAnalyserMinDecibels     = -100.0
	malgoDefaultAnalyserMaxDecibels     = -30.0
	malgoDefaultAnalyserSmoothing       = 0.8
	malgoDefaultChannelSplitterOutputs  = 6
	malgoDefaultChannelMergerInputs     = 6
	malgoDefaultConstantSourceOffset    = 1.0
	malgoDefaultScriptProcessorCapacity = 4096
)

var (
	_ webaudioapi.BiquadFilterNode       = new(malgoBiquadFilterNode)
	_ webaudioapi.IIRFilterNode          = new(malgoIIRFilterNode)
	_ webaudioapi.DelayNode              = new(malgoDelayNode)
	_ webaudioapi.DynamicsCompressorNode = new(malgoDynamicsCompressorNode)
	_ webaudioapi.AnalyserNode           = new(malgoAnalyserNode)
	_ webaudioapi.PannerNode             = new(malgoPannerNode)
	_ webaudioapi.ChannelSplitterNode    = new(malgoChannelSplitterNode)
	_ webaudioapi.ChannelMergerNode      = new(malgoChannelMergerNode)
	_ webaudioapi.ConstantSourceNode     = new(malgoConstantSourceNode)
	_ webaudioapi.WaveShaperNode         = new(malgoWaveShaperNode)
	_ webaudioapi.PeriodicWave           = new(malgoPeriodicWave)
	_ webaudioapi.ScriptProcessorNode    = new(malgoScriptProcessorNode)
)

type malgoBiquadState struct {
	x1 float64
	x2 float64
	y1 float64
	y2 float64
}

type malgoBiquadFilterNode struct {
	malgoAudioNode

	mu         sync.Mutex
	frequency  *malgoAudioParam
	detune     *malgoAudioParam
	q          *malgoAudioParam
	gain       *malgoAudioParam
	filterType webaudioapi.BiquadFilterType
	states     []malgoBiquadState
}

func (inst *malgoBiquadFilterNode) GetFrequency() webaudioapi.AudioParam {
	return inst.frequency
}

func (inst *malgoBiquadFilterNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

func (inst *malgoBiquadFilterNode) GetQ() webaudioapi.AudioParam {
	return inst.q
}

func (inst *malgoBiquadFilterNode) GetGain() webaudioapi.AudioParam {
	return inst.gain
}

func (inst *malgoBiquadFilterNode) GetType() webaudioapi.BiquadFilterType {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.filterType
}

func (inst *malgoBiquadFilterNode) SetType(filterType webaudioapi.BiquadFilterType) {
	inst.mu.Lock()
	inst.filterType = filterType
	inst.mu.Unlock()
}

func (inst *malgoBiquadFilterNode) GetFrequencyResponse(
	frequencyHz []float64,
	magnitude []float64,
	phase []float64,
) {
	inst.mu.Lock()
	filterType := inst.filterType
	inst.mu.Unlock()
	fillMalgoBiquadFrequencyResponse(
		filterType,
		inst.ctx.sampleRate,
		inst.frequency.GetValue(),
		inst.detune.GetValue(),
		inst.q.GetValue(),
		inst.gain.GetValue(),
		frequencyHz,
		magnitude,
		phase,
	)
}

func (inst *malgoBiquadFilterNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	output := mixSignals(inputs, frameCount, channels)

	inst.mu.Lock()
	defer inst.mu.Unlock()
	if len(inst.states) != channels {
		states := make([]malgoBiquadState, channels)
		copy(states, inst.states)
		inst.states = states
	}

	for frame := 0; frame < frameCount; frame++ {
		time := startTime + webaudioapi.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
		b0, b1, b2, a1, a2, ok := malgoBiquadCoefficients(
			inst.filterType,
			inst.ctx.sampleRate,
			inst.frequency.valueAt(time),
			inst.detune.valueAt(time),
			inst.q.valueAt(time),
			inst.gain.valueAt(time),
		)
		if !ok {
			continue
		}
		for channel := 0; channel < channels; channel++ {
			state := &inst.states[channel]
			x := float64(output[channel][frame])
			y := b0*x + b1*state.x1 + b2*state.x2 - a1*state.y1 - a2*state.y2
			state.x2, state.x1 = state.x1, x
			state.y2, state.y1 = state.y1, y
			output[channel][frame] = webaudioapi.SampleValue(y)
		}
	}
	return output
}

type malgoIIRFilterNode struct {
	malgoAudioNode
}

func (inst *malgoIIRFilterNode) GetFrequencyResponse(
	frequencyHz []float64,
	magnitude []float64,
	phase []float64,
) {
	for i := range frequencyHz {
		if i < len(magnitude) {
			magnitude[i] = 1
		}
		if i < len(phase) {
			phase[i] = 0
		}
	}
}

func (inst *malgoIIRFilterNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	return mixSignals(inputs, frameCount, channels)
}

type malgoDelayNode struct {
	malgoAudioNode

	mu        sync.Mutex
	delayTime *malgoAudioParam
	lines     [][]webaudioapi.SampleValue
	writePos  int
}

func (inst *malgoDelayNode) GetDelayTime() webaudioapi.AudioParam {
	return inst.delayTime
}

func (inst *malgoDelayNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	output := mixSignals(inputs, frameCount, channels)

	inst.mu.Lock()
	defer inst.mu.Unlock()
	lineLength := int(math.Ceil(float64(inst.ctx.sampleRate) * malgoDefaultDelayCapacitySeconds))
	if lineLength < 2 {
		lineLength = 2
	}
	if len(inst.lines) != channels {
		lines := make([][]webaudioapi.SampleValue, channels)
		copy(lines, inst.lines)
		inst.lines = lines
	}
	for channel := range inst.lines {
		if len(inst.lines[channel]) != lineLength {
			inst.lines[channel] = make([]webaudioapi.SampleValue, lineLength)
		}
	}
	if inst.writePos < 0 || inst.writePos >= lineLength {
		inst.writePos = 0
	}

	for frame := 0; frame < frameCount; frame++ {
		time := startTime + webaudioapi.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
		delay := inst.delayTime.valueAt(time)
		delayFrames := int(math.Round(delay * float64(inst.ctx.sampleRate)))
		if delayFrames < 0 {
			delayFrames = 0
		} else if delayFrames >= lineLength {
			delayFrames = lineLength - 1
		}
		for channel := range output {
			input := output[channel][frame]
			if delayFrames > 0 {
				readPos := (inst.writePos - delayFrames + lineLength) % lineLength
				output[channel][frame] = inst.lines[channel][readPos]
			}
			inst.lines[channel][inst.writePos] = input
		}
		inst.writePos = (inst.writePos + 1) % lineLength
	}
	return output
}

type malgoDynamicsCompressorNode struct {
	malgoAudioNode

	mu        sync.Mutex
	threshold *malgoAudioParam
	knee      *malgoAudioParam
	ratio     *malgoAudioParam
	attack    *malgoAudioParam
	release   *malgoAudioParam
	gain      float64
	reduction float64
}

func (inst *malgoDynamicsCompressorNode) GetThreshold() webaudioapi.AudioParam {
	return inst.threshold
}

func (inst *malgoDynamicsCompressorNode) GetKnee() webaudioapi.AudioParam {
	return inst.knee
}

func (inst *malgoDynamicsCompressorNode) GetRatio() webaudioapi.AudioParam {
	return inst.ratio
}

func (inst *malgoDynamicsCompressorNode) GetAttack() webaudioapi.AudioParam {
	return inst.attack
}

func (inst *malgoDynamicsCompressorNode) GetRelease() webaudioapi.AudioParam {
	return inst.release
}

func (inst *malgoDynamicsCompressorNode) GetReduction() float64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.reduction
}

func (inst *malgoDynamicsCompressorNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	output := mixSignals(inputs, frameCount, channels)

	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.gain <= 0 || inst.gain > 1 {
		inst.gain = 1
	}

	for frame := 0; frame < frameCount; frame++ {
		time := startTime + webaudioapi.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
		peak := 0.0
		for channel := range output {
			absolute := math.Abs(float64(output[channel][frame]))
			if absolute > peak {
				peak = absolute
			}
		}
		targetGain := 1.0
		if peak > 0 {
			levelDB := 20 * math.Log10(peak)
			threshold := inst.threshold.valueAt(time)
			ratio := math.Max(1, inst.ratio.valueAt(time))
			if levelDB > threshold {
				compressedDB := threshold + (levelDB-threshold)/ratio
				targetGain = math.Pow(10, (compressedDB-levelDB)/20)
			}
		}

		attack := math.Max(0, inst.attack.valueAt(time))
		release := math.Max(0, inst.release.valueAt(time))
		seconds := 0.0
		if targetGain < inst.gain {
			seconds = attack
		} else {
			seconds = release
		}
		if seconds <= 0 {
			inst.gain = targetGain
		} else {
			coefficient := math.Exp(-1 / (seconds * float64(inst.ctx.sampleRate)))
			inst.gain = targetGain + (inst.gain-targetGain)*coefficient
		}
		inst.reduction = 20 * math.Log10(inst.gain)
		for channel := range output {
			output[channel][frame] *= webaudioapi.SampleValue(inst.gain)
		}
	}
	return output
}

type malgoAnalyserNode struct {
	malgoAudioNode

	mu        sync.RWMutex
	fftSize   int
	minDB     float64
	maxDB     float64
	smoothing float64
}

func (inst *malgoAnalyserNode) GetFFTSize() int {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.fftSize
}

func (inst *malgoAnalyserNode) SetFFTSize(size int) {
	if !validMalgoAnalyserFFTSize(size) {
		return
	}
	inst.mu.Lock()
	inst.fftSize = size
	inst.mu.Unlock()
}

func (inst *malgoAnalyserNode) GetFrequencyBinCount() int {
	return inst.GetFFTSize() / 2
}

func (inst *malgoAnalyserNode) GetMinDecibels() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.minDB
}

func (inst *malgoAnalyserNode) SetMinDecibels(value float64) {
	inst.mu.Lock()
	if value < inst.maxDB && !math.IsNaN(value) && !math.IsInf(value, 0) {
		inst.minDB = value
	}
	inst.mu.Unlock()
}

func (inst *malgoAnalyserNode) GetMaxDecibels() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.maxDB
}

func (inst *malgoAnalyserNode) SetMaxDecibels(value float64) {
	inst.mu.Lock()
	if value > inst.minDB && !math.IsNaN(value) && !math.IsInf(value, 0) {
		inst.maxDB = value
	}
	inst.mu.Unlock()
}

func (inst *malgoAnalyserNode) GetSmoothingTimeConstant() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.smoothing
}

func (inst *malgoAnalyserNode) SetSmoothingTimeConstant(value float64) {
	if math.IsNaN(value) {
		return
	}
	inst.mu.Lock()
	inst.smoothing = math.Max(0, math.Min(1, value))
	inst.mu.Unlock()
}

func (inst *malgoAnalyserNode) GetFloatFrequencyData(destination []float32) {
	minimum := float32(inst.GetMinDecibels())
	for i := range destination {
		destination[i] = minimum
	}
}

func (inst *malgoAnalyserNode) GetByteFrequencyData(destination []uint8) {
	clear(destination)
}

func (inst *malgoAnalyserNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	return mixSignals(inputs, frameCount, channels)
}

type malgoPannerNode struct {
	malgoAudioNode

	mu            sync.RWMutex
	x             float64
	y             float64
	z             float64
	orientationX  float64
	orientationY  float64
	orientationZ  float64
	panningModel  webaudioapi.PanningModelType
	distanceModel webaudioapi.DistanceModelType
	refDistance   float64
	maxDistance   float64
	rolloffFactor float64
}

func (inst *malgoPannerNode) GetPositionX() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.x
}

func (inst *malgoPannerNode) GetPositionY() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.y
}

func (inst *malgoPannerNode) GetPositionZ() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.z
}

func (inst *malgoPannerNode) SetPosition(x float64, y float64, z float64) {
	inst.mu.Lock()
	inst.x, inst.y, inst.z = finiteMalgoOr(x, 0), finiteMalgoOr(y, 0), finiteMalgoOr(z, 0)
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetOrientationX() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationX
}

func (inst *malgoPannerNode) GetOrientationY() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationY
}

func (inst *malgoPannerNode) GetOrientationZ() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationZ
}

func (inst *malgoPannerNode) SetOrientation(x float64, y float64, z float64) {
	inst.mu.Lock()
	inst.orientationX, inst.orientationY, inst.orientationZ = finiteMalgoOr(x, 0), finiteMalgoOr(y, 0), finiteMalgoOr(z, 0)
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetPanningModel() webaudioapi.PanningModelType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.panningModel
}

func (inst *malgoPannerNode) SetPanningModel(model webaudioapi.PanningModelType) {
	inst.mu.Lock()
	inst.panningModel = model
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetDistanceModel() webaudioapi.DistanceModelType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.distanceModel
}

func (inst *malgoPannerNode) SetDistanceModel(model webaudioapi.DistanceModelType) {
	inst.mu.Lock()
	inst.distanceModel = model
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetRefDistance() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.refDistance
}

func (inst *malgoPannerNode) SetRefDistance(value float64) {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.refDistance = value
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetMaxDistance() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.maxDistance
}

func (inst *malgoPannerNode) SetMaxDistance(value float64) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.maxDistance = value
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) GetRolloffFactor() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.rolloffFactor
}

func (inst *malgoPannerNode) SetRolloffFactor(value float64) {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.rolloffFactor = value
	inst.mu.Unlock()
}

func (inst *malgoPannerNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	mono := mixSignals(inputs, frameCount, 1)
	output := [][]webaudioapi.SampleValue{
		make([]webaudioapi.SampleValue, frameCount),
		make([]webaudioapi.SampleValue, frameCount),
	}

	inst.mu.RLock()
	x, y, z := inst.x, inst.y, inst.z
	distanceModel := inst.distanceModel
	refDistance := inst.refDistance
	maxDistance := inst.maxDistance
	rolloff := inst.rolloffFactor
	inst.mu.RUnlock()

	distance := math.Sqrt(x*x + y*y + z*z)
	attenuation := pannerDistanceGain(distanceModel, distance, refDistance, maxDistance, rolloff)
	pan := x / math.Sqrt(x*x+1)
	angle := (pan + 1) * math.Pi / 4
	left := math.Cos(angle) * attenuation
	right := math.Sin(angle) * attenuation

	for frame := 0; frame < frameCount; frame++ {
		output[0][frame] = mono[0][frame] * webaudioapi.SampleValue(left)
		output[1][frame] = mono[0][frame] * webaudioapi.SampleValue(right)
	}
	return output
}

type malgoChannelSplitterNode struct {
	malgoAudioNode
	numberOfOutputs int
}

func (inst *malgoChannelSplitterNode) GetNumberOfOutputs() int {
	return inst.numberOfOutputs
}

func (inst *malgoChannelSplitterNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	output := make([][]webaudioapi.SampleValue, inst.numberOfOutputs)
	for channel := range output {
		output[channel] = make([]webaudioapi.SampleValue, frameCount)
	}
	for _, signal := range inputs {
		for channel := 0; channel < inst.numberOfOutputs && channel < len(signal); channel++ {
			for frame := 0; frame < frameCount && frame < len(signal[channel]); frame++ {
				output[channel][frame] += signal[channel][frame]
			}
		}
	}
	return output
}

type malgoChannelMergerNode struct {
	malgoAudioNode
	numberOfInputs int
}

func (inst *malgoChannelMergerNode) GetNumberOfInputs() int {
	return inst.numberOfInputs
}

func (inst *malgoChannelMergerNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	output := make([][]webaudioapi.SampleValue, inst.numberOfInputs)
	for channel := range output {
		output[channel] = make([]webaudioapi.SampleValue, frameCount)
	}

	inputIndex := 0
	for _, signal := range inputs {
		for channel := 0; channel < len(signal) && inputIndex < inst.numberOfInputs; channel++ {
			for frame := 0; frame < frameCount && frame < len(signal[channel]); frame++ {
				output[inputIndex][frame] += signal[channel][frame]
			}
			inputIndex++
		}
	}
	return output
}

type malgoConstantSourceNode struct {
	malgoAudioNode
	schedule malgoSourceSchedule
	offset   *malgoAudioParam
}

func (inst *malgoConstantSourceNode) GetOffset() webaudioapi.AudioParam {
	return inst.offset
}

func (inst *malgoConstantSourceNode) Start(when webaudioapi.Seconds) {
	inst.schedule.Start(when)
}

func (inst *malgoConstantSourceNode) Stop(when webaudioapi.Seconds) {
	inst.schedule.Stop(when)
}

func (inst *malgoConstantSourceNode) SetOnEnded(handler func()) {
	inst.schedule.SetOnEnded(handler)
}

func (inst *malgoConstantSourceNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	_ int,
	_ map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	output := [][]webaudioapi.SampleValue{make([]webaudioapi.SampleValue, frameCount)}
	schedule := &inst.schedule

	schedule.mu.Lock()
	defer schedule.mu.Unlock()
	if schedule.ended || !schedule.startScheduled {
		return output
	}

	ended := false
	handler := func() {}
	for frame := 0; frame < frameCount; frame++ {
		time := startTime + webaudioapi.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
		if schedule.started && schedule.stopScheduled && time >= schedule.stopTime {
			schedule.ended = true
			ended = true
			handler = schedule.onEnded
			break
		}
		if !schedule.started && time >= schedule.startTime {
			schedule.started = true
		}
		if schedule.started {
			output[0][frame] = webaudioapi.SampleValue(inst.offset.valueAt(time))
		}
	}
	if ended {
		defer schedule.notifyEnded(handler)
	}
	return output
}

type malgoWaveShaperNode struct {
	malgoAudioNode

	mu         sync.RWMutex
	curve      []webaudioapi.SampleValue
	oversample webaudioapi.OverSampleType
}

func (inst *malgoWaveShaperNode) GetCurve() []webaudioapi.SampleValue {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]webaudioapi.SampleValue(nil), inst.curve...)
}

func (inst *malgoWaveShaperNode) SetCurve(curve []webaudioapi.SampleValue) {
	inst.mu.Lock()
	inst.curve = append([]webaudioapi.SampleValue(nil), curve...)
	inst.mu.Unlock()
}

func (inst *malgoWaveShaperNode) GetOversample() webaudioapi.OverSampleType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.oversample
}

func (inst *malgoWaveShaperNode) SetOversample(oversample webaudioapi.OverSampleType) {
	inst.mu.Lock()
	inst.oversample = oversample
	inst.mu.Unlock()
}

func (inst *malgoWaveShaperNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	output := mixSignals(inputs, frameCount, channels)

	inst.mu.RLock()
	curve := append([]webaudioapi.SampleValue(nil), inst.curve...)
	inst.mu.RUnlock()
	if len(curve) == 0 {
		return output
	}
	for channel := range output {
		for frame := range output[channel] {
			x := math.Max(-1, math.Min(1, float64(output[channel][frame])))
			position := (x + 1) * 0.5 * float64(len(curve)-1)
			index := int(position)
			if index >= len(curve)-1 {
				output[channel][frame] = curve[len(curve)-1]
				continue
			}
			fraction := position - float64(index)
			value := float64(curve[index]) + (float64(curve[index+1])-float64(curve[index]))*fraction
			output[channel][frame] = webaudioapi.SampleValue(value)
		}
	}
	return output
}

type malgoPeriodicWave struct {
	mu   sync.RWMutex
	real []float64
	imag []float64
}

func (inst *malgoPeriodicWave) GetReal() []float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]float64(nil), inst.real...)
}

func (inst *malgoPeriodicWave) GetImag() []float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]float64(nil), inst.imag...)
}

type malgoScriptProcessorNode struct {
	malgoAudioNode
	bufferSize int
}

func (inst *malgoScriptProcessorNode) GetBufferSize() int {
	return inst.bufferSize
}

func (inst *malgoScriptProcessorNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {
	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	return mixSignals(inputs, frameCount, channels)
}

// CreateBiquadFilter implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateBiquadFilter() webaudioapi.BiquadFilterNode {
	now := inst.GetCurrentTime()
	node := &malgoBiquadFilterNode{
		frequency:  newMalgoAudioParam(malgoDefaultBiquadFrequency, now, inst.GetCurrentTime),
		detune:     newMalgoAudioParam(0, now, inst.GetCurrentTime),
		q:          newMalgoAudioParam(malgoDefaultBiquadQ, now, inst.GetCurrentTime),
		gain:       newMalgoAudioParam(0, now, inst.GetCurrentTime),
		filterType: webaudioapi.BiquadFilterTypeLowpass,
	}
	node.init(inst, node)
	return node
}

// CreateIIRFilter implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateIIRFilter() webaudioapi.IIRFilterNode {
	node := new(malgoIIRFilterNode)
	node.init(inst, node)
	return node
}

// CreateDelay implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateDelay() webaudioapi.DelayNode {
	now := inst.GetCurrentTime()
	node := &malgoDelayNode{
		delayTime: newMalgoAudioParam(malgoDefaultDelayTime, now, inst.GetCurrentTime),
	}
	node.init(inst, node)
	return node
}

// CreateDynamicsCompressor implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateDynamicsCompressor() webaudioapi.DynamicsCompressorNode {
	now := inst.GetCurrentTime()
	node := &malgoDynamicsCompressorNode{
		threshold: newMalgoAudioParam(malgoDefaultCompressorThreshold, now, inst.GetCurrentTime),
		knee:      newMalgoAudioParam(malgoDefaultCompressorKnee, now, inst.GetCurrentTime),
		ratio:     newMalgoAudioParam(malgoDefaultCompressorRatio, now, inst.GetCurrentTime),
		attack:    newMalgoAudioParam(malgoDefaultCompressorAttack, now, inst.GetCurrentTime),
		release:   newMalgoAudioParam(malgoDefaultCompressorRelease, now, inst.GetCurrentTime),
		gain:      1,
	}
	node.init(inst, node)
	return node
}

// CreateAnalyser implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateAnalyser() webaudioapi.AnalyserNode {
	node := &malgoAnalyserNode{
		fftSize:   malgoDefaultAnalyserFFTSize,
		minDB:     malgoDefaultAnalyserMinDecibels,
		maxDB:     malgoDefaultAnalyserMaxDecibels,
		smoothing: malgoDefaultAnalyserSmoothing,
	}
	node.init(inst, node)
	return node
}

// CreatePanner implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreatePanner() webaudioapi.PannerNode {
	node := &malgoPannerNode{
		orientationX:  1,
		panningModel:  webaudioapi.PanningModelTypeEqualPower,
		distanceModel: webaudioapi.DistanceModelTypeInverse,
		refDistance:   1,
		maxDistance:   10000,
		rolloffFactor: 1,
	}
	node.init(inst, node)
	return node
}

// CreateChannelSplitter implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateChannelSplitter() webaudioapi.ChannelSplitterNode {
	node := &malgoChannelSplitterNode{numberOfOutputs: malgoDefaultChannelSplitterOutputs}
	node.init(inst, node)
	return node
}

// CreateChannelMerger implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateChannelMerger() webaudioapi.ChannelMergerNode {
	node := &malgoChannelMergerNode{numberOfInputs: malgoDefaultChannelMergerInputs}
	node.init(inst, node)
	return node
}

// CreateConstantSource implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateConstantSource() webaudioapi.ConstantSourceNode {
	now := inst.GetCurrentTime()
	node := &malgoConstantSourceNode{
		offset: newMalgoAudioParam(malgoDefaultConstantSourceOffset, now, inst.GetCurrentTime),
	}
	node.init(inst, node)
	node.schedule.ctx = inst
	return node
}

// CreateWaveShaper implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateWaveShaper() webaudioapi.WaveShaperNode {
	node := &malgoWaveShaperNode{oversample: webaudioapi.OverSampleTypeNone}
	node.init(inst, node)
	return node
}

// CreatePeriodicWave implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreatePeriodicWave() webaudioapi.PeriodicWave {
	return &malgoPeriodicWave{
		real: []float64{0, 1},
		imag: []float64{0, 0},
	}
}

// CreateScriptProcessor implements [webaudioapi.BaseAudioContext].
func (inst *innerMalgoAudioContext) CreateScriptProcessor() webaudioapi.ScriptProcessorNode {
	node := &malgoScriptProcessorNode{bufferSize: malgoDefaultScriptProcessorCapacity}
	node.init(inst, node)
	return node
}

func validMalgoAnalyserFFTSize(size int) bool {
	if size < 32 || size > 32768 {
		return false
	}
	return size&(size-1) == 0
}

func finiteMalgoOr(value float64, fallback float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fallback
	}
	return value
}

func pannerDistanceGain(
	model webaudioapi.DistanceModelType,
	distance float64,
	refDistance float64,
	maxDistance float64,
	rolloff float64,
) float64 {
	if distance <= refDistance || refDistance <= 0 {
		return 1
	}
	distance = math.Min(distance, maxDistance)
	switch model {
	case webaudioapi.DistanceModelTypeLinear:
		return math.Max(0, 1-rolloff*(distance-refDistance)/(maxDistance-refDistance))
	case webaudioapi.DistanceModelTypeExponential:
		return math.Pow(distance/refDistance, -rolloff)
	case webaudioapi.DistanceModelTypeInverse:
		fallthrough
	default:
		return refDistance / (refDistance + rolloff*(distance-refDistance))
	}
}

func fillMalgoBiquadFrequencyResponse(
	filterType webaudioapi.BiquadFilterType,
	sampleRate webaudioapi.SampleRate,
	frequency float64,
	detune float64,
	q float64,
	gainDB float64,
	frequencyHz []float64,
	magnitude []float64,
	phase []float64,
) {
	b0, b1, b2, a1, a2, ok := malgoBiquadCoefficients(
		filterType,
		sampleRate,
		frequency,
		detune,
		q,
		gainDB,
	)
	for i := range frequencyHz {
		if !ok || frequencyHz[i] < 0 || sampleRate <= 0 {
			if i < len(magnitude) {
				magnitude[i] = 0
			}
			if i < len(phase) {
				phase[i] = 0
			}
			continue
		}
		omega := 2 * math.Pi * frequencyHz[i] / float64(sampleRate)
		cos1, sin1 := math.Cos(omega), math.Sin(omega)
		cos2, sin2 := math.Cos(2*omega), math.Sin(2*omega)
		numReal := b0 + b1*cos1 + b2*cos2
		numImag := -(b1*sin1 + b2*sin2)
		denReal := 1 + a1*cos1 + a2*cos2
		denImag := -(a1*sin1 + a2*sin2)
		denominator := denReal*denReal + denImag*denImag
		if denominator == 0 {
			if i < len(magnitude) {
				magnitude[i] = 0
			}
			if i < len(phase) {
				phase[i] = 0
			}
			continue
		}
		if i < len(magnitude) {
			magnitude[i] = math.Sqrt((numReal*numReal + numImag*numImag) / denominator)
		}
		if i < len(phase) {
			phase[i] = math.Atan2(numImag, numReal) - math.Atan2(denImag, denReal)
		}
	}
}

func malgoBiquadCoefficients(
	filterType webaudioapi.BiquadFilterType,
	sampleRate webaudioapi.SampleRate,
	frequency float64,
	detune float64,
	q float64,
	gainDB float64,
) (b0 float64, b1 float64, b2 float64, a1 float64, a2 float64, ok bool) {
	if sampleRate <= 0 {
		return 0, 0, 0, 0, 0, false
	}
	frequency *= math.Pow(2, detune/1200)
	frequency = math.Max(0, math.Min(frequency, float64(sampleRate)/2))
	w0 := 2 * math.Pi * frequency / float64(sampleRate)
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	q = math.Max(q, 1e-6)
	alpha := sinW0 / (2 * q)
	a := math.Pow(10, gainDB/40)

	switch filterType {
	case webaudioapi.BiquadFilterTypeHighpass:
		b0, b1, b2 = (1+cosW0)/2, -(1 + cosW0), (1+cosW0)/2
		a1, a2 = -2*cosW0, 1-alpha
	case webaudioapi.BiquadFilterTypeBandpass:
		b0, b1, b2 = alpha, 0, -alpha
		a1, a2 = -2*cosW0, 1-alpha
	case webaudioapi.BiquadFilterTypeLowshelf:
		shelf := 2 * math.Sqrt(a) * alpha
		b0 = a * ((a + 1) - (a-1)*cosW0 + shelf)
		b1 = 2 * a * ((a - 1) - (a+1)*cosW0)
		b2 = a * ((a + 1) - (a-1)*cosW0 - shelf)
		a1 = -2 * ((a - 1) + (a+1)*cosW0)
		a2 = (a + 1) + (a-1)*cosW0 - shelf
	case webaudioapi.BiquadFilterTypeHighshelf:
		shelf := 2 * math.Sqrt(a) * alpha
		b0 = a * ((a + 1) + (a-1)*cosW0 + shelf)
		b1 = -2 * a * ((a - 1) + (a+1)*cosW0)
		b2 = a * ((a + 1) + (a-1)*cosW0 - shelf)
		a1 = 2 * ((a - 1) - (a+1)*cosW0)
		a2 = (a + 1) - (a-1)*cosW0 - shelf
	case webaudioapi.BiquadFilterTypePeaking:
		b0, b1, b2 = 1+alpha*a, -2*cosW0, 1-alpha*a
		a1, a2 = -2*cosW0, 1-alpha/a
	case webaudioapi.BiquadFilterTypeNotch:
		b0, b1, b2 = 1, -2*cosW0, 1
		a1, a2 = -2*cosW0, 1-alpha
	case webaudioapi.BiquadFilterTypeAllpass:
		b0, b1, b2 = 1-alpha, -2*cosW0, 1+alpha
		a1, a2 = -2*cosW0, 1-alpha
	case webaudioapi.BiquadFilterTypeLowpass:
		fallthrough
	default:
		b0, b1, b2 = (1-cosW0)/2, 1-cosW0, (1-cosW0)/2
		a1, a2 = -2*cosW0, 1-alpha
	}

	a0 := 1 + alpha
	if filterType == webaudioapi.BiquadFilterTypeLowshelf ||
		filterType == webaudioapi.BiquadFilterTypeHighshelf {
		a0 = (a + 1) + (a-1)*cosW0 + 2*math.Sqrt(a)*alpha
		if filterType == webaudioapi.BiquadFilterTypeHighshelf {
			a0 = (a + 1) - (a-1)*cosW0 + 2*math.Sqrt(a)*alpha
		}
	} else if filterType == webaudioapi.BiquadFilterTypePeaking {
		a0 = 1 + alpha/a
	}
	if a0 == 0 {
		return 0, 0, 0, 0, 0, false
	}
	return b0 / a0, b1 / a0, b2 / a0, a1 / a0, a2 / a0, true
}
