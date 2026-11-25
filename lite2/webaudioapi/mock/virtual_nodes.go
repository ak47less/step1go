package mock

import (
	"math"
	"sync"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

const (
	mockDefaultBiquadFrequency         = 350.0
	mockDefaultBiquadQ                 = 1.0
	mockDefaultDelayTime               = 0.0
	mockDefaultCompressorThreshold     = -24.0
	mockDefaultCompressorKnee          = 30.0
	mockDefaultCompressorRatio         = 12.0
	mockDefaultCompressorAttack        = 0.003
	mockDefaultCompressorRelease       = 0.25
	mockDefaultAnalyserFFTSize         = 2048
	mockDefaultAnalyserMinDecibels     = -100.0
	mockDefaultAnalyserMaxDecibels     = -30.0
	mockDefaultAnalyserSmoothing       = 0.8
	mockDefaultChannelSplitterOutputs  = 6
	mockDefaultChannelMergerInputs     = 6
	mockDefaultConstantSourceOffset    = 1.0
	mockDefaultScriptProcessorCapacity = 4096
)

var (
	_ webaudioapi.BiquadFilterNode       = new(innerMockBiquadFilterNode)
	_ webaudioapi.IIRFilterNode          = new(innerMockIIRFilterNode)
	_ webaudioapi.DelayNode              = new(innerMockDelayNode)
	_ webaudioapi.DynamicsCompressorNode = new(innerMockDynamicsCompressorNode)
	_ webaudioapi.AnalyserNode           = new(innerMockAnalyserNode)
	_ webaudioapi.PannerNode             = new(innerMockPannerNode)
	_ webaudioapi.ChannelSplitterNode    = new(innerMockChannelSplitterNode)
	_ webaudioapi.ChannelMergerNode      = new(innerMockChannelMergerNode)
	_ webaudioapi.ConstantSourceNode     = new(innerMockConstantSourceNode)
	_ webaudioapi.WaveShaperNode         = new(innerMockWaveShaperNode)
	_ webaudioapi.PeriodicWave           = new(innerMockPeriodicWave)
	_ webaudioapi.ScriptProcessorNode    = new(innerMockScriptProcessorNode)
	_ webaudioapi.AudioProcessingEvent   = new(innerMockAudioProcessingEvent)
)

type innerMockBiquadFilterNode struct {
	innerMockAudioNode

	mu         sync.RWMutex
	sampleRate webaudioapi.SampleRate
	frequency  *innerMockAudioParam
	detune     *innerMockAudioParam
	q          *innerMockAudioParam
	gain       *innerMockAudioParam
	filterType webaudioapi.BiquadFilterType
}

func (inst *innerMockBiquadFilterNode) GetFrequency() webaudioapi.AudioParam {
	return inst.frequency
}

func (inst *innerMockBiquadFilterNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

func (inst *innerMockBiquadFilterNode) GetQ() webaudioapi.AudioParam {
	return inst.q
}

func (inst *innerMockBiquadFilterNode) GetGain() webaudioapi.AudioParam {
	return inst.gain
}

func (inst *innerMockBiquadFilterNode) GetType() webaudioapi.BiquadFilterType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.filterType
}

func (inst *innerMockBiquadFilterNode) SetType(filterType webaudioapi.BiquadFilterType) {
	inst.mu.Lock()
	inst.filterType = filterType
	inst.mu.Unlock()
}

func (inst *innerMockBiquadFilterNode) GetFrequencyResponse(
	frequencyHz []float64,
	magnitude []float64,
	phase []float64,
) {
	inst.mu.RLock()
	filterType := inst.filterType
	sampleRate := inst.sampleRate
	inst.mu.RUnlock()
	fillMockBiquadFrequencyResponse(
		filterType,
		sampleRate,
		inst.frequency.GetValue(),
		inst.detune.GetValue(),
		inst.q.GetValue(),
		inst.gain.GetValue(),
		frequencyHz,
		magnitude,
		phase,
	)
}

type innerMockIIRFilterNode struct {
	innerMockAudioNode
}

func (inst *innerMockIIRFilterNode) GetFrequencyResponse(
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

type innerMockDelayNode struct {
	innerMockAudioNode
	delayTime *innerMockAudioParam
}

func (inst *innerMockDelayNode) GetDelayTime() webaudioapi.AudioParam {
	return inst.delayTime
}

type innerMockDynamicsCompressorNode struct {
	innerMockAudioNode

	mu        sync.RWMutex
	threshold *innerMockAudioParam
	knee      *innerMockAudioParam
	ratio     *innerMockAudioParam
	attack    *innerMockAudioParam
	release   *innerMockAudioParam
	reduction float64
}

func (inst *innerMockDynamicsCompressorNode) GetThreshold() webaudioapi.AudioParam {
	return inst.threshold
}

func (inst *innerMockDynamicsCompressorNode) GetKnee() webaudioapi.AudioParam {
	return inst.knee
}

func (inst *innerMockDynamicsCompressorNode) GetRatio() webaudioapi.AudioParam {
	return inst.ratio
}

func (inst *innerMockDynamicsCompressorNode) GetAttack() webaudioapi.AudioParam {
	return inst.attack
}

func (inst *innerMockDynamicsCompressorNode) GetRelease() webaudioapi.AudioParam {
	return inst.release
}

func (inst *innerMockDynamicsCompressorNode) GetReduction() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.reduction
}

type innerMockAnalyserNode struct {
	innerMockAudioNode

	mu        sync.RWMutex
	fftSize   int
	minDB     float64
	maxDB     float64
	smoothing float64
}

func (inst *innerMockAnalyserNode) GetFFTSize() int {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.fftSize
}

func (inst *innerMockAnalyserNode) SetFFTSize(size int) {
	if !validAnalyserFFTSize(size) {
		return
	}
	inst.mu.Lock()
	inst.fftSize = size
	inst.mu.Unlock()
}

func (inst *innerMockAnalyserNode) GetFrequencyBinCount() int {
	return inst.GetFFTSize() / 2
}

func (inst *innerMockAnalyserNode) GetMinDecibels() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.minDB
}

func (inst *innerMockAnalyserNode) SetMinDecibels(value float64) {
	inst.mu.Lock()
	if value < inst.maxDB && !math.IsNaN(value) && !math.IsInf(value, 0) {
		inst.minDB = value
	}
	inst.mu.Unlock()
}

func (inst *innerMockAnalyserNode) GetMaxDecibels() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.maxDB
}

func (inst *innerMockAnalyserNode) SetMaxDecibels(value float64) {
	inst.mu.Lock()
	if value > inst.minDB && !math.IsNaN(value) && !math.IsInf(value, 0) {
		inst.maxDB = value
	}
	inst.mu.Unlock()
}

func (inst *innerMockAnalyserNode) GetSmoothingTimeConstant() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.smoothing
}

func (inst *innerMockAnalyserNode) SetSmoothingTimeConstant(value float64) {
	if math.IsNaN(value) {
		return
	}
	inst.mu.Lock()
	inst.smoothing = math.Max(0, math.Min(1, value))
	inst.mu.Unlock()
}

func (inst *innerMockAnalyserNode) GetFloatFrequencyData(destination []float32) {
	minDB := float32(inst.GetMinDecibels())
	for i := range destination {
		destination[i] = minDB
	}
}

func (inst *innerMockAnalyserNode) GetByteFrequencyData(destination []uint8) {
	clear(destination)
}

type innerMockPannerNode struct {
	innerMockAudioNode

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

func (inst *innerMockPannerNode) GetPositionX() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.x
}

func (inst *innerMockPannerNode) GetPositionY() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.y
}

func (inst *innerMockPannerNode) GetPositionZ() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.z
}

func (inst *innerMockPannerNode) SetPosition(x float64, y float64, z float64) {
	inst.mu.Lock()
	inst.x, inst.y, inst.z = finiteOr(x, 0), finiteOr(y, 0), finiteOr(z, 0)
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetOrientationX() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationX
}

func (inst *innerMockPannerNode) GetOrientationY() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationY
}

func (inst *innerMockPannerNode) GetOrientationZ() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.orientationZ
}

func (inst *innerMockPannerNode) SetOrientation(x float64, y float64, z float64) {
	inst.mu.Lock()
	inst.orientationX, inst.orientationY, inst.orientationZ = finiteOr(x, 0), finiteOr(y, 0), finiteOr(z, 0)
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetPanningModel() webaudioapi.PanningModelType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.panningModel
}

func (inst *innerMockPannerNode) SetPanningModel(model webaudioapi.PanningModelType) {
	inst.mu.Lock()
	inst.panningModel = model
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetDistanceModel() webaudioapi.DistanceModelType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.distanceModel
}

func (inst *innerMockPannerNode) SetDistanceModel(model webaudioapi.DistanceModelType) {
	inst.mu.Lock()
	inst.distanceModel = model
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetRefDistance() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.refDistance
}

func (inst *innerMockPannerNode) SetRefDistance(value float64) {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.refDistance = value
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetMaxDistance() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.maxDistance
}

func (inst *innerMockPannerNode) SetMaxDistance(value float64) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.maxDistance = value
	inst.mu.Unlock()
}

func (inst *innerMockPannerNode) GetRolloffFactor() float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.rolloffFactor
}

func (inst *innerMockPannerNode) SetRolloffFactor(value float64) {
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	inst.mu.Lock()
	inst.rolloffFactor = value
	inst.mu.Unlock()
}

type innerMockChannelSplitterNode struct {
	innerMockAudioNode
	numberOfOutputs int
}

func (inst *innerMockChannelSplitterNode) GetNumberOfOutputs() int {
	return inst.numberOfOutputs
}

type innerMockChannelMergerNode struct {
	innerMockAudioNode
	numberOfInputs int
}

func (inst *innerMockChannelMergerNode) GetNumberOfInputs() int {
	return inst.numberOfInputs
}

type innerMockConstantSourceNode struct {
	*innerMockAudioScheduledSourceNode
	offset *innerMockAudioParam
}

func (inst *innerMockConstantSourceNode) GetOffset() webaudioapi.AudioParam {
	return inst.offset
}

type innerMockWaveShaperNode struct {
	innerMockAudioNode

	mu         sync.RWMutex
	curve      []webaudioapi.SampleValue
	oversample webaudioapi.OverSampleType
}

func (inst *innerMockWaveShaperNode) GetCurve() []webaudioapi.SampleValue {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]webaudioapi.SampleValue(nil), inst.curve...)
}

func (inst *innerMockWaveShaperNode) SetCurve(curve []webaudioapi.SampleValue) {
	inst.mu.Lock()
	inst.curve = append([]webaudioapi.SampleValue(nil), curve...)
	inst.mu.Unlock()
}

func (inst *innerMockWaveShaperNode) GetOversample() webaudioapi.OverSampleType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.oversample
}

func (inst *innerMockWaveShaperNode) SetOversample(oversample webaudioapi.OverSampleType) {
	inst.mu.Lock()
	inst.oversample = oversample
	inst.mu.Unlock()
}

type innerMockPeriodicWave struct {
	mu   sync.RWMutex
	real []float64
	imag []float64
}

func (inst *innerMockPeriodicWave) GetReal() []float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]float64(nil), inst.real...)
}

func (inst *innerMockPeriodicWave) GetImag() []float64 {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]float64(nil), inst.imag...)
}

type innerMockScriptProcessorNode struct {
	innerMockAudioNode

	mu             sync.RWMutex
	bufferSize     int
	onAudioProcess webaudioapi.AudioProcessHandler
}

func (inst *innerMockScriptProcessorNode) GetBufferSize() int {
	return inst.bufferSize
}

func (inst *innerMockScriptProcessorNode) GetOnAudioProcess() webaudioapi.AudioProcessHandler {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.onAudioProcess
}

func (inst *innerMockScriptProcessorNode) SetOnAudioProcess(handler webaudioapi.AudioProcessHandler) {
	inst.mu.Lock()
	inst.onAudioProcess = handler
	inst.mu.Unlock()
}

type innerMockAudioProcessingEvent struct {
	playbackTime webaudioapi.Seconds
	inputBuffer  webaudioapi.AudioBuffer
	outputBuffer webaudioapi.AudioBuffer
}

func (inst *innerMockAudioProcessingEvent) GetPlaybackTime() webaudioapi.Seconds {
	return inst.playbackTime
}

func (inst *innerMockAudioProcessingEvent) GetInputBuffer() webaudioapi.AudioBuffer {
	return inst.inputBuffer
}

func (inst *innerMockAudioProcessingEvent) GetOutputBuffer() webaudioapi.AudioBuffer {
	return inst.outputBuffer
}

// CreateBiquadFilter implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateBiquadFilter() webaudioapi.BiquadFilterNode {
	now := inst.currentTime()
	return &innerMockBiquadFilterNode{
		sampleRate: inst.sampleRate,
		frequency:  newInnerMockAudioParam(mockDefaultBiquadFrequency, now, inst.GetCurrentTime),
		detune:     newInnerMockAudioParam(0, now, inst.GetCurrentTime),
		q:          newInnerMockAudioParam(mockDefaultBiquadQ, now, inst.GetCurrentTime),
		gain:       newInnerMockAudioParam(0, now, inst.GetCurrentTime),
		filterType: webaudioapi.BiquadFilterTypeLowpass,
	}
}

// CreateIIRFilter implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateIIRFilter() webaudioapi.IIRFilterNode {
	return new(innerMockIIRFilterNode)
}

// CreateDelay implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateDelay() webaudioapi.DelayNode {
	now := inst.currentTime()
	return &innerMockDelayNode{
		delayTime: newInnerMockAudioParam(mockDefaultDelayTime, now, inst.GetCurrentTime),
	}
}

// CreateDynamicsCompressor implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateDynamicsCompressor() webaudioapi.DynamicsCompressorNode {
	now := inst.currentTime()
	return &innerMockDynamicsCompressorNode{
		threshold: newInnerMockAudioParam(mockDefaultCompressorThreshold, now, inst.GetCurrentTime),
		knee:      newInnerMockAudioParam(mockDefaultCompressorKnee, now, inst.GetCurrentTime),
		ratio:     newInnerMockAudioParam(mockDefaultCompressorRatio, now, inst.GetCurrentTime),
		attack:    newInnerMockAudioParam(mockDefaultCompressorAttack, now, inst.GetCurrentTime),
		release:   newInnerMockAudioParam(mockDefaultCompressorRelease, now, inst.GetCurrentTime),
	}
}

// CreateAnalyser implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateAnalyser() webaudioapi.AnalyserNode {
	return &innerMockAnalyserNode{
		fftSize:   mockDefaultAnalyserFFTSize,
		minDB:     mockDefaultAnalyserMinDecibels,
		maxDB:     mockDefaultAnalyserMaxDecibels,
		smoothing: mockDefaultAnalyserSmoothing,
	}
}

// CreatePanner implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreatePanner() webaudioapi.PannerNode {
	return &innerMockPannerNode{
		orientationX:  1,
		panningModel:  webaudioapi.PanningModelTypeEqualPower,
		distanceModel: webaudioapi.DistanceModelTypeInverse,
		refDistance:   1,
		maxDistance:   10000,
		rolloffFactor: 1,
	}
}

// CreateChannelSplitter implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateChannelSplitter() webaudioapi.ChannelSplitterNode {
	return &innerMockChannelSplitterNode{numberOfOutputs: mockDefaultChannelSplitterOutputs}
}

// CreateChannelMerger implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateChannelMerger() webaudioapi.ChannelMergerNode {
	return &innerMockChannelMergerNode{numberOfInputs: mockDefaultChannelMergerInputs}
}

// CreateConstantSource implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateConstantSource() webaudioapi.ConstantSourceNode {
	now := inst.currentTime()
	return &innerMockConstantSourceNode{
		innerMockAudioScheduledSourceNode: newInnerMockAudioScheduledSourceNode(inst),
		offset:                            newInnerMockAudioParam(mockDefaultConstantSourceOffset, now, inst.GetCurrentTime),
	}
}

// CreateWaveShaper implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateWaveShaper() webaudioapi.WaveShaperNode {
	return &innerMockWaveShaperNode{oversample: webaudioapi.OverSampleTypeNone}
}

// CreatePeriodicWave implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreatePeriodicWave() webaudioapi.PeriodicWave {
	return &innerMockPeriodicWave{
		real: []float64{0, 1},
		imag: []float64{0, 0},
	}
}

// CreateScriptProcessor implements [webaudioapi.BaseAudioContext].
func (inst *innerMockAudioContext) CreateScriptProcessor() webaudioapi.ScriptProcessorNode {
	return &innerMockScriptProcessorNode{bufferSize: mockDefaultScriptProcessorCapacity}
}

func validAnalyserFFTSize(size int) bool {
	if size < 32 || size > 32768 {
		return false
	}
	return size&(size-1) == 0
}

func finiteOr(value float64, fallback float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fallback
	}
	return value
}

func fillMockBiquadFrequencyResponse(
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
	b0, b1, b2, a1, a2, ok := mockBiquadCoefficients(
		filterType,
		sampleRate,
		frequency,
		detune,
		q,
		gainDB,
	)
	for i := range frequencyHz {
		if !ok || frequencyHz[i] < 0 || float64(sampleRate) <= 0 {
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

func mockBiquadCoefficients(
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
