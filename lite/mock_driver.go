package lite

import (
	"math"
	"sync"
	"time"
)

const (
	mockDefaultSampleRate   SampleRate = 44100
	mockDefaultChannelCount            = 2
	mockDefaultFrequency               = 440.0
	mockDefaultGain                    = 1.0
	mockDefaultPlaybackRate            = 1.0
)

var (
	_ Driver = new(innerMockDriver)

	_ AudioContextFactory      = new(innerMockAudioContextFactory)
	_ AudioContext             = new(innerMockAudioContext)
	_ OscillatorNode           = new(innerMockOscillatorNode)
	_ GainNode                 = new(innerMockGainNode)
	_ AudioDestinationNode     = new(innerMockAudioDestinationNode)
	_ AudioBuffer              = new(innerMockAudioBuffer)
	_ AudioBufferSourceNode    = new(innerMockAudioBufferSourceNode)
	_ AudioParam               = new(innerMockAudioParam)
	_ AudioScheduledSourceNode = new(innerMockAudioScheduledSourceNode)
)

type innerMockDriver struct {
}

// NewMockDriver 创建一个不依赖真实音频后端的驱动。
func NewMockDriver() Driver {
	return new(innerMockDriver)
}

// GetAudioContextFactory implements [Driver].
func (inst *innerMockDriver) GetAudioContextFactory() AudioContextFactory {
	return new(innerMockAudioContextFactory)
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioContextFactory struct {
}

// NewAudioContext implements [AudioContextFactory].
func (inst *innerMockAudioContextFactory) NewAudioContext() AudioContext {
	return newInnerMockAudioContext()
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioNode struct {
	mu          sync.RWMutex
	connections []AudioNode
}

// Connect implements [AudioNode].
func (inst *innerMockAudioNode) Connect(destination AudioNode) {
	if destination == nil {
		return
	}
	inst.mu.Lock()
	inst.connections = append(inst.connections, destination)
	inst.mu.Unlock()
}

// Disconnect implements [AudioNode].
func (inst *innerMockAudioNode) Disconnect() {
	inst.mu.Lock()
	inst.connections = nil
	inst.mu.Unlock()
}

func (inst *innerMockAudioNode) getConnections() []AudioNode {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return append([]AudioNode(nil), inst.connections...)
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
	time         Seconds
	timeConstant Seconds
}

type innerMockAudioParam struct {
	mu sync.Mutex

	baseValue float64
	baseTime  Seconds
	events    []mockAudioParamEvent
	now       func() Seconds
}

func newInnerMockAudioParam(value float64, baseTime Seconds, now func() Seconds) *innerMockAudioParam {
	if now == nil {
		now = func() Seconds { return 0 }
	}
	return &innerMockAudioParam{
		baseValue: value,
		baseTime:  baseTime,
		now:       now,
	}
}

// GetValue implements [AudioParam].
func (inst *innerMockAudioParam) GetValue() float64 {
	t := inst.now()
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.valueAt(t)
}

// SetValue implements [AudioParam].
func (inst *innerMockAudioParam) SetValue(value float64) {
	t := inst.now()
	inst.mu.Lock()
	inst.baseValue = value
	inst.baseTime = t
	inst.events = nil
	inst.mu.Unlock()
}

// SetValueAtTime implements [AudioParam].
func (inst *innerMockAudioParam) SetValueAtTime(value float64, startTime Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventSet,
		value: value,
		time:  normalizeMockTime(startTime),
	})
}

// LinearRampToValueAtTime implements [AudioParam].
func (inst *innerMockAudioParam) LinearRampToValueAtTime(value float64, endTime Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventLinearRamp,
		value: value,
		time:  normalizeMockTime(endTime),
	})
}

// ExponentialRampToValueAtTime implements [AudioParam].
func (inst *innerMockAudioParam) ExponentialRampToValueAtTime(value float64, endTime Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:  mockAudioParamEventExponentialRamp,
		value: value,
		time:  normalizeMockTime(endTime),
	})
}

// SetTargetAtTime implements [AudioParam].
func (inst *innerMockAudioParam) SetTargetAtTime(target float64, startTime Seconds, timeConstant Seconds) {
	inst.addEvent(mockAudioParamEvent{
		kind:         mockAudioParamEventTarget,
		value:        target,
		time:         normalizeMockTime(startTime),
		timeConstant: normalizeMockTime(timeConstant),
	})
}

// CancelScheduledValues implements [AudioParam].
func (inst *innerMockAudioParam) CancelScheduledValues(startTime Seconds) {
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

func (inst *innerMockAudioParam) valueAt(t Seconds) float64 {
	value := inst.baseValue
	cursor := inst.baseTime

	mode := mockAudioParamEventSet
	target := value
	timeConstant := Seconds(0)

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
	startTime Seconds,
	startValue float64,
	t Seconds,
	mode mockAudioParamEventKind,
	target float64,
	timeConstant Seconds,
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

func interpolateMockValue(startTime Seconds, startValue float64, endTime Seconds, endValue float64, t Seconds) float64 {
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
	startTime Seconds,
	startValue float64,
	endTime Seconds,
	endValue float64,
	t Seconds,
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
	startTime      Seconds
	started        bool

	stopScheduled bool
	stopTime      Seconds
	ended         bool

	onEnded func()
}

func newInnerMockAudioScheduledSourceNode(ctx *innerMockAudioContext) *innerMockAudioScheduledSourceNode {
	return &innerMockAudioScheduledSourceNode{ctx: ctx}
}

// Start implements [AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) Start(when Seconds) {
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

// Stop implements [AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) Stop(when Seconds) {
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

// SetOnEnded implements [AudioScheduledSourceNode].
func (inst *innerMockAudioScheduledSourceNode) SetOnEnded(handler func()) {
	inst.mu.Lock()
	inst.onEnded = handler
	ended := inst.ended
	inst.mu.Unlock()

	if ended && handler != nil {
		handler()
	}
}

func (inst *innerMockAudioScheduledSourceNode) tick(now Seconds) {
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
	nodeType  OscillatorType
}

// GetFrequency implements [OscillatorNode].
func (inst *innerMockOscillatorNode) GetFrequency() AudioParam {
	return inst.frequency
}

// GetDetune implements [OscillatorNode].
func (inst *innerMockOscillatorNode) GetDetune() AudioParam {
	return inst.detune
}

// GetType implements [OscillatorNode].
func (inst *innerMockOscillatorNode) GetType() OscillatorType {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.nodeType
}

// SetType implements [OscillatorNode].
func (inst *innerMockOscillatorNode) SetType(nodeType OscillatorType) {
	inst.mu.Lock()
	inst.nodeType = nodeType
	inst.mu.Unlock()
}

////////////////////////////////////////////////////////////////////////////////

type innerMockGainNode struct {
	innerMockAudioNode
	gain *innerMockAudioParam
}

// GetGain implements [GainNode].
func (inst *innerMockGainNode) GetGain() AudioParam {
	return inst.gain
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioDestinationNode struct {
	innerMockAudioNode
	maxChannelCount int
}

// GetMaxChannelCount implements [AudioDestinationNode].
func (inst *innerMockAudioDestinationNode) GetMaxChannelCount() int {
	return inst.maxChannelCount
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioBuffer struct {
	sampleRate SampleRate
	channels   [][]SampleValue
}

// GetSampleRate implements [AudioBuffer].
func (inst *innerMockAudioBuffer) GetSampleRate() SampleRate {
	return inst.sampleRate
}

// GetLength implements [AudioBuffer].
func (inst *innerMockAudioBuffer) GetLength() SampleCount {
	if len(inst.channels) == 0 {
		return 0
	}
	return SampleCount(len(inst.channels[0]))
}

// GetNumberOfChannels implements [AudioBuffer].
func (inst *innerMockAudioBuffer) GetNumberOfChannels() int {
	return len(inst.channels)
}

// GetChannelData implements [AudioBuffer].
func (inst *innerMockAudioBuffer) GetChannelData(channel int) []SampleValue {
	if channel < 0 || channel >= len(inst.channels) {
		return nil
	}
	return inst.channels[channel]
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioBufferSourceNode struct {
	*innerMockAudioScheduledSourceNode

	mu sync.RWMutex

	buffer       AudioBuffer
	playbackRate *innerMockAudioParam
	detune       *innerMockAudioParam
}

// GetBuffer implements [AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetBuffer() AudioBuffer {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.buffer
}

// SetBuffer implements [AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) SetBuffer(buffer AudioBuffer) {
	inst.mu.Lock()
	inst.buffer = buffer
	inst.mu.Unlock()
}

// GetPlaybackRate implements [AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetPlaybackRate() AudioParam {
	return inst.playbackRate
}

// GetDetune implements [AudioBufferSourceNode].
func (inst *innerMockAudioBufferSourceNode) GetDetune() AudioParam {
	return inst.detune
}

////////////////////////////////////////////////////////////////////////////////

type innerMockAudioContext struct {
	mu sync.Mutex

	state      AudioContextState
	sampleRate SampleRate

	elapsed      Seconds
	runningSince time.Time

	destination *innerMockAudioDestinationNode
	sources     map[*innerMockAudioScheduledSourceNode]struct{}
}

func newInnerMockAudioContext() *innerMockAudioContext {
	ctx := &innerMockAudioContext{
		state:      AudioContextStateSuspended,
		sampleRate: mockDefaultSampleRate,
		sources:    make(map[*innerMockAudioScheduledSourceNode]struct{}),
	}
	ctx.destination = &innerMockAudioDestinationNode{
		maxChannelCount: mockDefaultChannelCount,
	}
	return ctx
}

// Close implements [AudioContext].
func (inst *innerMockAudioContext) Close() {
	inst.mu.Lock()
	if inst.state == AudioContextStateClosed {
		inst.mu.Unlock()
		return
	}
	inst.freezeClock()
	inst.state = AudioContextStateClosed

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

// Resume implements [AudioContext].
func (inst *innerMockAudioContext) Resume() {
	inst.mu.Lock()
	if inst.state == AudioContextStateClosed || inst.state == AudioContextStateRunning {
		inst.mu.Unlock()
		return
	}
	inst.state = AudioContextStateRunning
	inst.runningSince = time.Now()
	inst.mu.Unlock()

	inst.runSources()
}

// Suspend implements [AudioContext].
func (inst *innerMockAudioContext) Suspend() {
	inst.mu.Lock()
	if inst.state == AudioContextStateRunning {
		inst.freezeClock()
	}
	if inst.state != AudioContextStateClosed {
		inst.state = AudioContextStateSuspended
	}
	inst.mu.Unlock()
}

// CreateOscillator implements [AudioContext].
func (inst *innerMockAudioContext) CreateOscillator() OscillatorNode {
	source := newInnerMockAudioScheduledSourceNode(inst)
	now := inst.currentTime()
	return &innerMockOscillatorNode{
		innerMockAudioScheduledSourceNode: source,
		frequency:                         newInnerMockAudioParam(mockDefaultFrequency, now, inst.GetCurrentTime),
		detune:                            newInnerMockAudioParam(0, now, inst.GetCurrentTime),
		nodeType:                          OscillatorTypeSine,
	}
}

// CreateGain implements [AudioContext].
func (inst *innerMockAudioContext) CreateGain() GainNode {
	now := inst.currentTime()
	return &innerMockGainNode{
		gain: newInnerMockAudioParam(mockDefaultGain, now, inst.GetCurrentTime),
	}
}

// CreateBuffer implements [AudioContext].
func (inst *innerMockAudioContext) CreateBuffer(
	numberOfChannels int,
	length SampleCount,
	sampleRate SampleRate,
) AudioBuffer {

	if numberOfChannels < 0 {
		numberOfChannels = 0
	}
	if sampleRate == 0 {
		sampleRate = inst.sampleRate
	}

	maxInt := SampleCount(^uint(0) >> 1)
	if length > maxInt {
		length = 0
	}

	channels := make([][]SampleValue, numberOfChannels)
	for i := range channels {
		channels[i] = make([]SampleValue, int(length))
	}
	return &innerMockAudioBuffer{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// CreateBufferSource implements [AudioContext].
func (inst *innerMockAudioContext) CreateBufferSource() AudioBufferSourceNode {
	source := newInnerMockAudioScheduledSourceNode(inst)
	now := inst.currentTime()
	return &innerMockAudioBufferSourceNode{
		innerMockAudioScheduledSourceNode: source,
		playbackRate:                      newInnerMockAudioParam(mockDefaultPlaybackRate, now, inst.GetCurrentTime),
		detune:                            newInnerMockAudioParam(0, now, inst.GetCurrentTime),
	}
}

// GetCurrentTime implements [AudioContext].
func (inst *innerMockAudioContext) GetCurrentTime() Seconds {
	inst.runSources()
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.currentTimeLocked()
}

// GetSampleRate implements [AudioContext].
func (inst *innerMockAudioContext) GetSampleRate() SampleRate {
	return inst.sampleRate
}

// GetDestination implements [AudioContext].
func (inst *innerMockAudioContext) GetDestination() AudioDestinationNode {
	return inst.destination
}

// GetState implements [AudioContext].
func (inst *innerMockAudioContext) GetState() AudioContextState {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.state
}

func (inst *innerMockAudioContext) currentTime() Seconds {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.currentTimeLocked()
}

func (inst *innerMockAudioContext) currentTimeLocked() Seconds {
	t := inst.elapsed
	if inst.state == AudioContextStateRunning {
		t += Seconds(time.Since(inst.runningSince).Seconds())
	}
	return t
}

func (inst *innerMockAudioContext) freezeClock() {
	if inst.state != AudioContextStateRunning {
		return
	}
	inst.elapsed += Seconds(time.Since(inst.runningSince).Seconds())
	inst.runningSince = time.Time{}
}

func (inst *innerMockAudioContext) registerSource(source *innerMockAudioScheduledSourceNode) {
	inst.mu.Lock()
	if inst.state != AudioContextStateClosed {
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
	if inst.state != AudioContextStateRunning {
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

func normalizeMockTime(value Seconds) Seconds {
	if value < 0 || math.IsNaN(float64(value)) {
		return 0
	}
	return value
}

////////////////////////////////////////////////////////////////////////////////
// EOF
