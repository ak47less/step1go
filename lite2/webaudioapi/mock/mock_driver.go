package mock

import (
	"math"
	"sync"
	"time"

	"github.com/ak47less/step1go/lite2/webaudioapi"
)

const (
	mockDefaultSampleRate   webaudioapi.SampleRate = 44100
	mockDefaultChannelCount                        = 2
	mockDefaultFrequency                           = 440.0
	mockDefaultGain                                = 1.0
	mockDefaultPlaybackRate                        = 1.0
)

var (
	_ webaudioapi.Driver = new(TheMockDriver)

	_ webaudioapi.AudioContextFactory      = new(innerMockAudioContextFactory)
	_ webaudioapi.AudioContext             = new(innerMockAudioContext)
	_ webaudioapi.OscillatorNode           = new(innerMockOscillatorNode)
	_ webaudioapi.GainNode                 = new(innerMockGainNode)
	_ webaudioapi.AudioDestinationNode     = new(innerMockAudioDestinationNode)
	_ webaudioapi.AudioBuffer              = new(innerMockAudioBuffer)
	_ webaudioapi.AudioBufferSourceNode    = new(innerMockAudioBufferSourceNode)
	_ webaudioapi.AudioParam               = new(innerMockAudioParam)
	_ webaudioapi.AudioScheduledSourceNode = new(innerMockAudioScheduledSourceNode)
)

type TheMockDriver struct {
}

// NewMockDriver 创建一个不依赖真实音频后端的驱动。
func NewMockDriver() webaudioapi.Driver {
	return new(TheMockDriver)
}

// GetFactory implements [webaudioapi.Driver].
func (inst *TheMockDriver) GetFactory() webaudioapi.AudioContextFactory {
	return new(innerMockAudioContextFactory)
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioContextFactory struct {
}

// CreateAudioContext implements [webaudioapi.AudioContextFactory].
func (inst *innerMockAudioContextFactory) CreateAudioContext() (webaudioapi.AudioContext, error) {
	return newInnerMockAudioContext(), nil
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioNode struct {
	mu          sync.RWMutex
	connections []webaudioapi.AudioNode
}

// Connect implements [webaudioapi.AudioNode].
func (inst *innerMockAudioNode) Connect(destination webaudioapi.AudioNode) {
	if destination == nil {
		return
	}
	inst.mu.Lock()
	inst.connections = append(inst.connections, destination)
	inst.mu.Unlock()
}

// Disconnect implements [webaudioapi.AudioNode].
func (inst *innerMockAudioNode) Disconnect() {
	inst.mu.Lock()
	inst.connections = nil
	inst.mu.Unlock()
}

func (inst *innerMockAudioNode) getConnections() []webaudioapi.AudioNode {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]webaudioapi.AudioNode(nil), inst.connections...)
}

////////////////////////////////////////////////////////////////////////////////

type mockAudioParamEventKind uint8

const (
	mockAudioParamEventSet mockAudioParamEventKind = iota
	mockAudioParamEventLinearRamp
	mockAudioParamEventExponentialRamp
	mockAudioParamEventTarget
)

type mockAudioParamEvent struct {
	kind         mockAudioParamEventKind
	value        float64
	time         webaudioapi.Seconds
	timeConstant webaudioapi.Seconds
}

type innerMockAudioParam struct {
	mu sync.Mutex

	baseValue float64
	baseTime  webaudioapi.Seconds
	events    []mockAudioParamEvent
	now       func() webaudioapi.Seconds
}

func newInnerMockAudioParam(value float64, baseTime webaudioapi.Seconds, now func() webaudioapi.Seconds) *innerMockAudioParam {
	if now == nil {
		now = func() webaudioapi.Seconds { return 0 }
	}
	return &innerMockAudioParam{
		baseValue: value,
		baseTime:  baseTime,
		now:       now,
	}
}

// GetValue implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) GetValue() float64 {
	t := inst.now()
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.valueAt(t)
}

// SetValue implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) SetValue(value float64) {
	t := inst.now()
	inst.mu.Lock()
	inst.baseValue = value
	inst.baseTime = t
	inst.events = nil
	inst.mu.Unlock()
}

// SetValueAtTime implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) SetValueAtTime(value float64, startTime webaudioapi.Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventSet,
		value: value,
		time:  normalizeMockTime(startTime),
	})
}

// LinearRampToValueAtTime implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) LinearRampToValueAtTime(value float64, endTime webaudioapi.Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventLinearRamp,
		value: value,
		time:  normalizeMockTime(endTime),
	})
}

// ExponentialRampToValueAtTime implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) ExponentialRampToValueAtTime(value float64, endTime webaudioapi.Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventExponentialRamp,
		value: value,
		time:  normalizeMockTime(endTime),
	})
}

// SetTargetAtTime implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) SetTargetAtTime(target float64, startTime webaudioapi.Seconds, timeConstant webaudioapi.Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:         mockAudioParamEventTarget,
		value:        target,
		time:         normalizeMockTime(startTime),
		timeConstant: normalizeMockTime(timeConstant),
	})
}

// CancelScheduledValues implements [webaudioapi.AudioParam].
func (inst *innerMockAudioParam) CancelScheduledValues(startTime webaudioapi.Seconds) {
	startTime = normalizeMockTime(startTime)
	now := inst.now()

	inst.mu.Lock()
	defer inst.mu.Unlock()

	inst.baseValue = inst.valueAt(now)
	inst.baseTime = now

	events := inst.events[:0]
	for _, event := range inst.events {
		if event.time > now && event.time < startTime {
			events = append(events, event)
		}
	}
	inst.events = events
}

func (inst *innerMockAudioParam) addEvent(event mockAudioParamEvent) {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	index := len(inst.events)
	for i, item := range inst.events {
		if item.time > event.time {
			index = i
			break
		}
	}

	inst.events = append(inst.events, mockAudioParamEvent{})
	copy(inst.events[index+1:], inst.events[index:])
	inst.events[index] = event
}

func (inst *innerMockAudioParam) valueAt(t webaudioapi.Seconds) float64 {
	value := inst.baseValue
	cursor := inst.baseTime

	mode := mockAudioParamEventSet
	target := value
	timeConstant := webaudioapi.Seconds(0)

	for _, event := range inst.events {
		if t < event.time {
			switch event.kind {
			case mockAudioParamEventLinearRamp:
				return interpolateMockValue(cursor, value, event.time, event.value, t)
			case mockAudioParamEventExponentialRamp:
				return interpolateMockExponentialValue(cursor, value, event.time, event.value, t)
			default:
				return evaluateMockAudioParam(cursor, value, t, mode, target, timeConstant)
			}
		}

		value = evaluateMockAudioParam(cursor, value, event.time, mode, target, timeConstant)
		cursor = event.time

		switch event.kind {
		case mockAudioParamEventTarget:
			mode = mockAudioParamEventTarget
			target = event.value
			timeConstant = event.timeConstant
		default:
			mode = mockAudioParamEventSet
			value = event.value
		}
	}

	return evaluateMockAudioParam(cursor, value, t, mode, target, timeConstant)
}

func evaluateMockAudioParam(
	startTime webaudioapi.Seconds,
	startValue float64,
	t webaudioapi.Seconds,
	mode mockAudioParamEventKind,
	target float64,
	timeConstant webaudioapi.Seconds,
) float64 {

	if t <= startTime {
		return startValue
	}

	elapsed := t - startTime
	if mode != mockAudioParamEventTarget {
		return startValue
	}
	if timeConstant <= 0 {
		return target
	}

	factor := 1 - math.Exp(-float64(elapsed/timeConstant))
	return startValue + (target-startValue)*factor
}

func interpolateMockValue(startTime webaudioapi.Seconds, startValue float64, endTime webaudioapi.Seconds, endValue float64, t webaudioapi.Seconds) float64 {
	if endTime <= startTime {
		return endValue
	}
	ratio := float64((t - startTime) / (endTime - startTime))
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	return startValue + (endValue-startValue)*ratio
}

func interpolateMockExponentialValue(
	startTime webaudioapi.Seconds,
	startValue float64,
	endTime webaudioapi.Seconds,
	endValue float64,
	t webaudioapi.Seconds,
) float64 {

	if startValue <= 0 || endValue <= 0 || endTime <= startTime {
		return interpolateMockValue(startTime, startValue, endTime, endValue, t)
	}
	ratio := float64((t - startTime) / (endTime - startTime))
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	return startValue * math.Pow(endValue/startValue, ratio)
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioScheduledSourceNode struct {
	innerMockAudioNode

	mu sync.Mutex

	ctx *innerMockAudioContext

	startScheduled bool
	startTime      webaudioapi.Seconds
	started        bool

	stopScheduled bool
	stopTime      webaudioapi.Seconds
	ended         bool

	onEnded func()
}

func newInnerMockAudioScheduledSourceNode(ctx *innerMockAudioContext) *innerMockAudioScheduledSourceNode {
	return &innerMockAudioScheduledSourceNode{ctx: ctx}
}

// Start implements [webaudioapi.AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) Start(when webaudioapi.Seconds) {
	when = normalizeMockTime(when)

	inst.mu.Lock()
	if inst.startScheduled || inst.ended {
		inst.mu.Unlock()
		return
	}
	inst.startScheduled = true
	inst.startTime = when
	inst.mu.Unlock()

	inst.ctx.registerSource(inst)
	inst.ctx.runSources()
}

// Stop implements [webaudioapi.AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) Stop(when webaudioapi.Seconds) {
	when = normalizeMockTime(when)

	inst.mu.Lock()
	if inst.ended {
		inst.mu.Unlock()
		return
	}
	inst.stopScheduled = true
	inst.stopTime = when
	inst.mu.Unlock()

	inst.ctx.registerSource(inst)
	inst.ctx.runSources()
}

// SetOnEnded implements [webaudioapi.AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) SetOnEnded(handler func()) {
	inst.mu.Lock()
	inst.onEnded = handler
	ended := inst.ended
	inst.mu.Unlock()

	if ended && handler != nil {
		handler()
	}
}

func (inst *innerMockAudioScheduledSourceNode) tick(now webaudioapi.Seconds) {
	inst.mu.Lock()
	if inst.ended || !inst.startScheduled {
		inst.mu.Unlock()
		return
	}
	if !inst.started && now >= inst.startTime {
		inst.started = true
	}

	handler := func() {}
	finished := inst.started && inst.stopScheduled && now >= inst.stopTime
	if finished {
		inst.ended = true
		if inst.onEnded != nil {
			handler = inst.onEnded
		}
	}
	inst.mu.Unlock()

	if finished {
		handler()
		inst.ctx.unregisterSource(inst)
	}
}

func (inst *innerMockAudioScheduledSourceNode) finish() {
	inst.mu.Lock()
	if inst.ended || !inst.started {
		inst.mu.Unlock()
		return
	}
	inst.ended = true
	handler := inst.onEnded
	inst.mu.Unlock()

	if handler != nil {
		handler()
	}
}

////////////////////////////////////////////////////////////////////////////////

type innerMockOscillatorNode struct {
	*innerMockAudioScheduledSourceNode

	mu sync.RWMutex

	frequency *innerMockAudioParam
	detune    *innerMockAudioParam
	nodeType  webaudioapi.OscillatorType
}

// GetFrequency implements [webaudioapi.OscillatorNode].
func (inst *innerMockOscillatorNode) GetFrequency() webaudioapi.AudioParam {
	return inst.frequency
}

// GetDetune implements [webaudioapi.OscillatorNode].
func (inst *innerMockOscillatorNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

// GetType implements [webaudioapi.OscillatorNode].
func (inst *innerMockOscillatorNode) GetType() webaudioapi.OscillatorType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.nodeType
}

// SetType implements [webaudioapi.OscillatorNode].
func (inst *innerMockOscillatorNode) SetType(nodeType webaudioapi.OscillatorType) {
	inst.mu.Lock()
	inst.nodeType = nodeType
	inst.mu.Unlock()
}

////////////////////////////////////////////////////////////////////////////////

type innerMockGainNode struct {
	innerMockAudioNode
	gain *innerMockAudioParam
}

// GetGain implements [webaudioapi.GainNode].
func (inst *innerMockGainNode) GetGain() webaudioapi.AudioParam {
	return inst.gain
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioDestinationNode struct {
	innerMockAudioNode
	maxChannelCount int
}

// GetMaxChannelCount implements [webaudioapi.AudioDestinationNode].
func (inst *innerMockAudioDestinationNode) GetMaxChannelCount() int {
	return inst.maxChannelCount
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioBuffer struct {
	sampleRate webaudioapi.SampleRate
	channels   [][]webaudioapi.SampleValue
}

// GetSampleRate implements [webaudioapi.AudioBuffer].
func (inst *innerMockAudioBuffer) GetSampleRate() webaudioapi.SampleRate {
	return inst.sampleRate
}

// GetLength implements [webaudioapi.AudioBuffer].
func (inst *innerMockAudioBuffer) GetLength() webaudioapi.SampleCount {
	if len(inst.channels) == 0 {
		return 0
	}
	return webaudioapi.SampleCount(len(inst.channels[0]))
}

// GetNumberOfChannels implements [webaudioapi.AudioBuffer].
func (inst *innerMockAudioBuffer) GetNumberOfChannels() int {
	return len(inst.channels)
}

// GetChannelData implements [webaudioapi.AudioBuffer].
func (inst *innerMockAudioBuffer) GetChannelData(channel int) []webaudioapi.SampleValue {
	if channel < 0 || channel >= len(inst.channels) {
		return nil
	}
	return inst.channels[channel]
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioBufferSourceNode struct {
	*innerMockAudioScheduledSourceNode

	mu sync.RWMutex

	buffer       webaudioapi.AudioBuffer
	playbackRate *innerMockAudioParam
	detune       *innerMockAudioParam
}

// GetBuffer implements [webaudioapi.AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetBuffer() webaudioapi.AudioBuffer {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.buffer
}

// SetBuffer implements [webaudioapi.AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) SetBuffer(buffer webaudioapi.AudioBuffer) {
	inst.mu.Lock()
	inst.buffer = buffer
	inst.mu.Unlock()
}

// GetPlaybackRate implements [webaudioapi.AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetPlaybackRate() webaudioapi.AudioParam {
	return inst.playbackRate
}

// GetDetune implements [webaudioapi.AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioContext struct {
	mu sync.Mutex

	state      webaudioapi.AudioContextState
	sampleRate webaudioapi.SampleRate

	elapsed      webaudioapi.Seconds
	runningSince time.Time

	destination *innerMockAudioDestinationNode
	sources     map[*innerMockAudioScheduledSourceNode]struct{}
}

func newInnerMockAudioContext() *innerMockAudioContext {
	ctx := &innerMockAudioContext{
		state:      webaudioapi.AudioContextStateSuspended,
		sampleRate: mockDefaultSampleRate,
		sources:    make(map[*innerMockAudioScheduledSourceNode]struct{}),
	}
	ctx.destination = &innerMockAudioDestinationNode{
		maxChannelCount: mockDefaultChannelCount,
	}
	return ctx
}

// Close implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) Close() {
	inst.mu.Lock()
	if inst.state == webaudioapi.AudioContextStateClosed {
		inst.mu.Unlock()
		return
	}
	inst.freezeClock()
	inst.state = webaudioapi.AudioContextStateClosed

	sources := make([]*innerMockAudioScheduledSourceNode, 0, len(inst.sources))
	for source := range inst.sources {
		sources = append(sources, source)
	}
	inst.sources = make(map[*innerMockAudioScheduledSourceNode]struct{})
	inst.mu.Unlock()

	for _, source := range sources {
		source.finish()
	}
}

// Resume implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) Resume() {
	inst.mu.Lock()
	if inst.state == webaudioapi.AudioContextStateClosed || inst.state == webaudioapi.AudioContextStateRunning {
		inst.mu.Unlock()
		return
	}
	inst.state = webaudioapi.AudioContextStateRunning
	inst.runningSince = time.Now()
	inst.mu.Unlock()

	inst.runSources()
}

// Suspend implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) Suspend() {
	inst.mu.Lock()
	if inst.state == webaudioapi.AudioContextStateRunning {
		inst.freezeClock()
	}
	if inst.state != webaudioapi.AudioContextStateClosed {
		inst.state = webaudioapi.AudioContextStateSuspended
	}
	inst.mu.Unlock()
}

// CreateOscillator implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) CreateOscillator() webaudioapi.OscillatorNode {
	source := newInnerMockAudioScheduledSourceNode(inst)
	now := inst.currentTime()
	return &innerMockOscillatorNode{
		innerMockAudioScheduledSourceNode: source,
		frequency:                         newInnerMockAudioParam(mockDefaultFrequency, now, inst.GetCurrentTime),
		detune:                            newInnerMockAudioParam(0, now, inst.GetCurrentTime),
		nodeType:                          webaudioapi.OscillatorTypeSine,
	}
}

// CreateGain implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) CreateGain() webaudioapi.GainNode {
	now := inst.currentTime()
	return &innerMockGainNode{
		gain: newInnerMockAudioParam(mockDefaultGain, now, inst.GetCurrentTime),
	}
}

// CreateBuffer implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) CreateBuffer(
	numberOfChannels int,
	length webaudioapi.SampleCount,
	sampleRate webaudioapi.SampleRate,
) webaudioapi.AudioBuffer {

	if numberOfChannels < 0 {
		numberOfChannels = 0
	}
	if sampleRate == 0 {
		sampleRate = inst.sampleRate
	}

	maxInt := webaudioapi.SampleCount(^uint(0) >> 1)
	if length > maxInt {
		length = 0
	}

	channels := make([][]webaudioapi.SampleValue, numberOfChannels)
	for i := range channels {
		channels[i] = make([]webaudioapi.SampleValue, int(length))
	}
	return &innerMockAudioBuffer{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// CreateBufferSource implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) CreateBufferSource() webaudioapi.AudioBufferSourceNode {
	source := newInnerMockAudioScheduledSourceNode(inst)
	now := inst.currentTime()
	return &innerMockAudioBufferSourceNode{
		innerMockAudioScheduledSourceNode: source,
		playbackRate:                      newInnerMockAudioParam(mockDefaultPlaybackRate, now, inst.GetCurrentTime),
		detune:                            newInnerMockAudioParam(0, now, inst.GetCurrentTime),
	}
}

// GetCurrentTime implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) GetCurrentTime() webaudioapi.Seconds {
	inst.runSources()
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.currentTimeLocked()
}

// GetSampleRate implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) GetSampleRate() webaudioapi.SampleRate {
	return inst.sampleRate
}

// GetDestination implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) GetDestination() webaudioapi.AudioDestinationNode {
	return inst.destination
}

// GetState implements [webaudioapi.AudioContext].
func (inst *innerMockAudioContext) GetState() webaudioapi.AudioContextState {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.state
}

func (inst *innerMockAudioContext) currentTime() webaudioapi.Seconds {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.currentTimeLocked()
}

func (inst *innerMockAudioContext) currentTimeLocked() webaudioapi.Seconds {
	t := inst.elapsed
	if inst.state == webaudioapi.AudioContextStateRunning {
		t += webaudioapi.Seconds(time.Since(inst.runningSince).Seconds())
	}
	return t
}

func (inst *innerMockAudioContext) freezeClock() {
	if inst.state != webaudioapi.AudioContextStateRunning {
		return
	}
	inst.elapsed += webaudioapi.Seconds(time.Since(inst.runningSince).Seconds())
	inst.runningSince = time.Time{}
}

func (inst *innerMockAudioContext) registerSource(source *innerMockAudioScheduledSourceNode) {
	inst.mu.Lock()
	if inst.state != webaudioapi.AudioContextStateClosed {
		inst.sources[source] = struct{}{}
	}
	inst.mu.Unlock()
}

func (inst *innerMockAudioContext) unregisterSource(source *innerMockAudioScheduledSourceNode) {
	inst.mu.Lock()
	delete(inst.sources, source)
	inst.mu.Unlock()
}

func (inst *innerMockAudioContext) runSources() {
	inst.mu.Lock()
	if inst.state != webaudioapi.AudioContextStateRunning {
		inst.mu.Unlock()
		return
	}
	now := inst.currentTimeLocked()
	sources := make([]*innerMockAudioScheduledSourceNode, 0, len(inst.sources))
	for source := range inst.sources {
		sources = append(sources, source)
	}
	inst.mu.Unlock()

	for _, source := range sources {
		source.tick(now)
	}
}

func normalizeMockTime(value webaudioapi.Seconds) webaudioapi.Seconds {
	if value < 0 || math.IsNaN(float64(value)) {
		return 0
	}
	return value
}

////////////////////////////////////////////////////////////////////////////////
// EOF
