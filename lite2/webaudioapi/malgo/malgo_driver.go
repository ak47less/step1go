package malgo

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ak47less/step1go/lite2/webaudioapi"
	"github.com/gen2brain/malgo"
)

const (
	malgoDefaultSampleRate   webaudioapi.SampleRate = 44100
	malgoDefaultChannelCount                        = 2
	malgoDefaultFrequency                           = 440.0
	malgoDefaultGain                                = 1.0
	malgoDefaultPlaybackRate                        = 1.0
	malgoMaxRenderDepth                             = 128

	// Leave a small amount of headroom at the destination so that rounding and
	// inter-sample peaks cannot reach full scale.
	malgoBusCeiling        webaudioapi.SampleValue = 0.98
	malgoBusReleaseSeconds                         = 0.25
)

var (
	_ webaudioapi.Driver                   = new(TheMalgoDriver)
	_ webaudioapi.AudioContextFactory      = new(innerMalgoAudioContextFactory)
	_ webaudioapi.AudioContext             = new(innerMalgoAudioContext)
	_ webaudioapi.AudioNode                = new(malgoAudioNode)
	_ webaudioapi.AudioParam               = new(malgoAudioParam)
	_ webaudioapi.OscillatorNode           = new(malgoOscillatorNode)
	_ webaudioapi.GainNode                 = new(malgoGainNode)
	_ webaudioapi.AudioDestinationNode     = new(malgoAudioDestinationNode)
	_ webaudioapi.AudioBuffer              = new(malgoAudioBuffer)
	_ webaudioapi.AudioBufferSourceNode    = new(malgoAudioBufferSourceNode)
	_ webaudioapi.AudioScheduledSourceNode = new(malgoOscillatorNode)

	errMalgoContextClosed = errors.New("malgo: audio context is closed")
)

////////////////////////////////////////////////////////////////////////////////
// class : TheMalgoDriver

// TheMalgoDriver 提供基于 miniaudio/malgo 的音频上下文。
type TheMalgoDriver struct {
}

// GetFactory implements [webaudioapi.Driver].
func (inst *TheMalgoDriver) GetFactory() webaudioapi.AudioContextFactory {
	return new(innerMalgoAudioContextFactory)
}

func (inst *TheMalgoDriver) _impl() webaudioapi.Driver {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class : innerMalgoAudioContextFactory

// innerMalgoAudioContextFactory 创建共享 malgo 后端的音频上下文。
type innerMalgoAudioContextFactory struct {
	config malgo.ContextConfig
}

// CreateAudioContext implements [webaudioapi.AudioContextFactory].
func (inst *innerMalgoAudioContextFactory) CreateAudioContext() (webaudioapi.AudioContext, error) {
	mc, err := malgo.InitContext(nil, inst.config, nil)
	if err != nil {
		return nil, err
	}
	return newInnerMalgoAudioContext(mc), nil
}

func (inst *innerMalgoAudioContextFactory) _impl() webaudioapi.AudioContextFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class: innerMalgoAudioContext

// innerMalgoAudioContext 实现 [webaudioapi.AudioContext]。
// 播放设备在第一次 Resume 时创建，因而可以离屏创建音频图和缓冲区。
type innerMalgoAudioContext struct {
	mc *malgo.AllocatedContext

	// lifecycleMu 串行化设备启动、停止和释放操作。
	lifecycleMu sync.Mutex
	// mu 同时保护上下文状态和音频图。
	mu sync.RWMutex
	// sourcesMu 单独保护源集合，使音频回调结束时不会与渲染锁互相等待。
	sourcesMu sync.Mutex

	state        webaudioapi.AudioContextState
	sampleRate   webaudioapi.SampleRate
	elapsed      webaudioapi.Seconds
	runningSince time.Time
	closed       atomic.Bool

	device      *malgo.Device
	destination *malgoAudioDestinationNode
	inputs      map[webaudioapi.AudioNode][]webaudioapi.AudioNode
	sources     map[*malgoSourceSchedule]struct{}
	busLimiter  malgoBusLimiter
}

func newInnerMalgoAudioContext(mc *malgo.AllocatedContext) *innerMalgoAudioContext {
	inst := &innerMalgoAudioContext{
		mc:         mc,
		state:      webaudioapi.AudioContextStateSuspended,
		sampleRate: malgoDefaultSampleRate,
		inputs:     make(map[webaudioapi.AudioNode][]webaudioapi.AudioNode),
		sources:    make(map[*malgoSourceSchedule]struct{}),
	}
	inst.destination = newMalgoAudioDestinationNode(inst)
	return inst
}

// Close implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) Close() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.RLock()
	closed := inst.state == webaudioapi.AudioContextStateClosed
	inst.mu.RUnlock()
	if closed {
		return
	}

	// Stop waits for the active data callback, so do not hold mu here.
	if inst.device != nil {
		_ = inst.device.Stop()
		inst.device.Uninit()
		inst.device = nil
	}

	inst.mu.Lock()
	inst.freezeClockLocked()
	inst.state = webaudioapi.AudioContextStateClosed
	inst.mu.Unlock()

	inst.closed.Store(true)
	inst.sourcesMu.Lock()
	sources := make([]*malgoSourceSchedule, 0, len(inst.sources))
	for source := range inst.sources {
		sources = append(sources, source)
	}
	inst.sources = make(map[*malgoSourceSchedule]struct{})
	inst.sourcesMu.Unlock()

	for _, source := range sources {
		source.finish()
	}

	if inst.mc != nil {
		_ = inst.mc.Uninit()
		inst.mc.Free()
		inst.mc = nil
	}
}

// Resume implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) Resume() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.RLock()
	state := inst.state
	inst.mu.RUnlock()
	if state == webaudioapi.AudioContextStateClosed || state == webaudioapi.AudioContextStateRunning {
		return
	}

	device, err := inst.ensureDevice()
	if err != nil {
		panic(err)
	}

	inst.mu.Lock()
	inst.state = webaudioapi.AudioContextStateRunning
	inst.runningSince = time.Now()
	inst.mu.Unlock()

	if err = device.Start(); err != nil {
		inst.mu.Lock()
		inst.freezeClockLocked()
		inst.state = webaudioapi.AudioContextStateSuspended
		inst.mu.Unlock()
		panic(err)
	}
}

// Suspend implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) Suspend() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.Lock()
	if inst.state == webaudioapi.AudioContextStateClosed {
		inst.mu.Unlock()
		return
	}
	if inst.state == webaudioapi.AudioContextStateRunning {
		inst.freezeClockLocked()
		inst.state = webaudioapi.AudioContextStateSuspended
	}
	device := inst.device
	inst.mu.Unlock()

	if device != nil && device.IsStarted() {
		_ = device.Stop()
	}
}

// CreateOscillator implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) CreateOscillator() webaudioapi.OscillatorNode {
	now := inst.GetCurrentTime()
	node := &malgoOscillatorNode{
		frequency: newMalgoAudioParam(malgoDefaultFrequency, now, inst.GetCurrentTime),
		detune:    newMalgoAudioParam(0, now, inst.GetCurrentTime),
		nodeType:  webaudioapi.OscillatorTypeSine,
	}
	node.init(inst, node)
	node.schedule.ctx = inst
	return node
}

// CreateGain implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) CreateGain() webaudioapi.GainNode {
	node := &malgoGainNode{
		gain: newMalgoAudioParam(malgoDefaultGain, inst.GetCurrentTime(), inst.GetCurrentTime),
	}
	node.init(inst, node)
	return node
}

// CreateBuffer implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) CreateBuffer(
	numberOfChannels int,
	length webaudioapi.SampleCount,
	sampleRate webaudioapi.SampleRate,
) webaudioapi.AudioBuffer {

	if numberOfChannels < 0 {
		numberOfChannels = 0
	}
	if sampleRate == 0 {
		sampleRate = inst.GetSampleRate()
	}

	maxInt := int(^uint(0) >> 1)
	if length > webaudioapi.SampleCount(maxInt) {
		length = 0
	}

	channels := make([][]webaudioapi.SampleValue, numberOfChannels)
	for i := range channels {
		channels[i] = make([]webaudioapi.SampleValue, int(length))
	}
	return &malgoAudioBuffer{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// CreateBufferSource implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) CreateBufferSource() webaudioapi.AudioBufferSourceNode {
	now := inst.GetCurrentTime()
	node := &malgoAudioBufferSourceNode{
		playbackRate: newMalgoAudioParam(malgoDefaultPlaybackRate, now, inst.GetCurrentTime),
		detune:       newMalgoAudioParam(0, now, inst.GetCurrentTime),
	}
	node.init(inst, node)
	node.schedule.ctx = inst
	return node
}

// GetCurrentTime implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) GetCurrentTime() webaudioapi.Seconds {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.currentTimeLocked()
}

// GetSampleRate implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) GetSampleRate() webaudioapi.SampleRate {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.sampleRate
}

// GetDestination implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) GetDestination() webaudioapi.AudioDestinationNode {
	return inst.destination
}

// GetState implements [webaudioapi.AudioContext].
func (inst *innerMalgoAudioContext) GetState() webaudioapi.AudioContextState {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.state
}

func (inst *innerMalgoAudioContext) _impl() webaudioapi.AudioContext {
	return inst
}

func (inst *innerMalgoAudioContext) ensureDevice() (*malgo.Device, error) {
	if inst.device != nil {
		return inst.device, nil
	}
	if inst.mc == nil {
		return nil, errMalgoContextClosed
	}

	config := malgo.DefaultDeviceConfig(malgo.Playback)
	config.SampleRate = uint32(inst.sampleRate)
	config.PeriodSizeInMilliseconds = 10
	config.Playback.Format = malgo.FormatF32
	config.Playback.Channels = malgoDefaultChannelCount
	config.NoClip = 1

	device, err := malgo.InitDevice(
		inst.mc.Context,
		config,
		malgo.DeviceCallbacks{Data: inst.onData},
	)
	if err != nil {
		return nil, err
	}
	inst.device = device
	return device, nil
}

func (inst *innerMalgoAudioContext) onData(output []byte, _ []byte, frameCount uint32) {
	if len(output) < 4 {
		return
	}

	samples := unsafe.Slice(
		(*float32)(unsafe.Pointer(unsafe.SliceData(output))),
		len(output)/4,
	)

	inst.mu.RLock()
	if inst.state != webaudioapi.AudioContextStateRunning {
		inst.mu.RUnlock()
		clear(samples)
		return
	}

	startTime := inst.currentTimeLocked()
	signal := inst.destination.renderSignal(
		startTime,
		int(frameCount),
		0,
		make(map[webaudioapi.AudioNode]bool),
	)
	inst.mu.RUnlock()

	for i := range samples {
		frame := i / malgoDefaultChannelCount
		channel := i % malgoDefaultChannelCount
		var value webaudioapi.SampleValue
		if channel < len(signal) && frame < len(signal[channel]) {
			value = signal[channel][frame]
		}
		if value > 1 {
			value = 1
		} else if value < -1 {
			value = -1
		}
		samples[i] = float32(value)
	}
}

func (inst *innerMalgoAudioContext) currentTimeLocked() webaudioapi.Seconds {
	now := inst.elapsed
	if inst.state == webaudioapi.AudioContextStateRunning {
		now += webaudioapi.Seconds(time.Since(inst.runningSince).Seconds())
	}
	return now
}

func (inst *innerMalgoAudioContext) freezeClockLocked() {
	if inst.state != webaudioapi.AudioContextStateRunning {
		return
	}
	inst.elapsed += webaudioapi.Seconds(time.Since(inst.runningSince).Seconds())
	inst.runningSince = time.Time{}
}

func (inst *innerMalgoAudioContext) connect(source webaudioapi.AudioNode, destination webaudioapi.AudioNode) {
	sourceProvider, ok := source.(malgoAudioNodeProvider)
	if !ok || sourceProvider.malgoAudioNodeBase().ctx != inst {
		return
	}
	destinationProvider, ok := destination.(malgoAudioNodeProvider)
	if !ok || destinationProvider.malgoAudioNodeBase().ctx != inst {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.state == webaudioapi.AudioContextStateClosed {
		return
	}
	for _, item := range inst.inputs[destination] {
		if item == source {
			return
		}
	}
	inst.inputs[destination] = append(inst.inputs[destination], source)
}

func (inst *innerMalgoAudioContext) disconnect(source webaudioapi.AudioNode) {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	for destination, inputs := range inst.inputs {
		filtered := inputs[:0]
		for _, item := range inputs {
			if item != source {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) == 0 {
			delete(inst.inputs, destination)
		} else {
			inst.inputs[destination] = filtered
		}
	}
}

func (inst *innerMalgoAudioContext) registerSource(source *malgoSourceSchedule) {
	inst.sourcesMu.Lock()
	if !inst.closed.Load() {
		inst.sources[source] = struct{}{}
	}
	inst.sourcesMu.Unlock()
}

func (inst *innerMalgoAudioContext) unregisterSource(source *malgoSourceSchedule) {
	inst.sourcesMu.Lock()
	delete(inst.sources, source)
	inst.sourcesMu.Unlock()
}

func (inst *innerMalgoAudioContext) renderNode(
	node webaudioapi.AudioNode,
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {

	if node == nil || frameCount <= 0 || depth > malgoMaxRenderDepth || visiting[node] {
		return nil
	}
	renderer, ok := node.(interface {
		renderSignal(
			webaudioapi.Seconds,
			int,
			int,
			map[webaudioapi.AudioNode]bool,
		) [][]webaudioapi.SampleValue
	})
	if !ok {
		return nil
	}

	visiting[node] = true
	defer delete(visiting, node)
	return renderer.renderSignal(startTime, frameCount, depth, visiting)
}

func (inst *innerMalgoAudioContext) renderInputs(
	node webaudioapi.AudioNode,
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][][]webaudioapi.SampleValue {
	inputs := inst.inputs[node]
	signals := make([][][]webaudioapi.SampleValue, 0, len(inputs))
	for _, input := range inputs {
		signal := inst.renderNode(input, startTime, frameCount, depth, visiting)
		if signal != nil {
			signals = append(signals, signal)
		}
	}
	return signals
}

////////////////////////////////////////////////////////////////////////////////
// audio graph nodes

type malgoAudioNode struct {
	ctx   *innerMalgoAudioContext
	owner webaudioapi.AudioNode
}

type malgoAudioNodeProvider interface {
	malgoAudioNodeBase() *malgoAudioNode
}

func (inst *malgoAudioNode) init(ctx *innerMalgoAudioContext, owner webaudioapi.AudioNode) {
	inst.ctx = ctx
	inst.owner = owner
}

func (inst *malgoAudioNode) malgoAudioNodeBase() *malgoAudioNode {
	return inst
}

// Connect implements [webaudioapi.AudioNode].
func (inst *malgoAudioNode) Connect(destination webaudioapi.AudioNode) {
	if destination != nil {
		inst.ctx.connect(inst.owner, destination)
	}
}

// Disconnect implements [webaudioapi.AudioNode].
func (inst *malgoAudioNode) Disconnect() {
	inst.ctx.disconnect(inst.owner)
}

type malgoGainNode struct {
	malgoAudioNode
	gain *malgoAudioParam
}

// GetGain implements [webaudioapi.GainNode].
func (inst *malgoGainNode) GetGain() webaudioapi.AudioParam {
	return inst.gain
}

func (inst *malgoGainNode) renderSignal(
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

	for channel := range output {
		for frame := 0; frame < frameCount; frame++ {
			time := startTime + webaudioapi.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
			output[channel][frame] *= webaudioapi.SampleValue(inst.gain.valueAt(time))
		}
	}
	return output
}

type malgoAudioDestinationNode struct {
	malgoAudioNode
}

func newMalgoAudioDestinationNode(ctx *innerMalgoAudioContext) *malgoAudioDestinationNode {
	inst := new(malgoAudioDestinationNode)
	inst.init(ctx, inst)
	return inst
}

// GetMaxChannelCount implements [webaudioapi.AudioDestinationNode].
func (inst *malgoAudioDestinationNode) GetMaxChannelCount() int {
	return malgoDefaultChannelCount
}

func (inst *malgoAudioDestinationNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	depth int,
	visiting map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {

	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	output := mixSignals(inputs, frameCount, malgoDefaultChannelCount)
	return inst.ctx.busLimiter.process(output, inst.ctx.sampleRate)
}

// malgoBusLimiter prevents summed signals from clipping at the destination.
// It applies one linked gain to all channels for a render block. Peaks are
// attacked immediately; after they fall, gain is restored linearly.
type malgoBusLimiter struct {
	mu   sync.Mutex
	gain webaudioapi.SampleValue
}

func (inst *malgoBusLimiter) process(
	signal [][]webaudioapi.SampleValue,
	sampleRate webaudioapi.SampleRate,
) [][]webaudioapi.SampleValue {
	if len(signal) == 0 {
		return signal
	}

	peak := webaudioapi.SampleValue(0)
	for _, channel := range signal {
		for _, sample := range channel {
			absolute := webaudioapi.SampleValue(math.Abs(float64(sample)))
			if absolute > peak {
				peak = absolute
			}
		}
	}

	target := webaudioapi.SampleValue(1)
	if peak > malgoBusCeiling {
		target = malgoBusCeiling / peak
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.gain <= 0 || inst.gain > 1 {
		inst.gain = 1
	}
	if target < inst.gain {
		// Immediate attack is required here because this callback renders the
		// entire block before it is written to the device.
		inst.gain = target
	}

	releaseSamples := int(float64(sampleRate) * malgoBusReleaseSeconds)
	if releaseSamples < 1 {
		releaseSamples = 1
	}
	releaseStep := webaudioapi.SampleValue(0)
	if inst.gain < target {
		releaseStep = (target - inst.gain) / webaudioapi.SampleValue(releaseSamples)
	}

	for frame := 0; frame < len(signal[0]); frame++ {
		if releaseStep > 0 {
			inst.gain += releaseStep
			if inst.gain > target {
				inst.gain = target
			}
		}
		for channel := range signal {
			if frame < len(signal[channel]) {
				signal[channel][frame] *= inst.gain
			}
		}
	}
	return signal
}

func signalChannelCount(signals [][][]webaudioapi.SampleValue) int {
	result := 0
	for _, signal := range signals {
		if len(signal) > result {
			result = len(signal)
		}
	}
	return result
}

func mixSignals(signals [][][]webaudioapi.SampleValue, frameCount int, channels int) [][]webaudioapi.SampleValue {
	output := make([][]webaudioapi.SampleValue, channels)
	for channel := range output {
		output[channel] = make([]webaudioapi.SampleValue, frameCount)
	}

	for _, signal := range signals {
		if len(signal) == 1 {
			for channel := 0; channel < channels; channel++ {
				for frame := 0; frame < frameCount && frame < len(signal[0]); frame++ {
					output[channel][frame] += signal[0][frame]
				}
			}
			continue
		}
		for channel := 0; channel < channels && channel < len(signal); channel++ {
			for frame := 0; frame < frameCount && frame < len(signal[channel]); frame++ {
				output[channel][frame] += signal[channel][frame]
			}
		}
	}
	return output
}

////////////////////////////////////////////////////////////////////////////////
// scheduled source nodes

type malgoSourceSchedule struct {
	ctx *innerMalgoAudioContext

	mu sync.Mutex

	startScheduled bool
	startTime      webaudioapi.Seconds
	started        bool

	stopScheduled bool
	stopTime      webaudioapi.Seconds
	ended         bool

	onEnded func()
}

// Start implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoSourceSchedule) Start(when webaudioapi.Seconds) {
	when = normalizeMalgoTime(when)

	inst.mu.Lock()
	if inst.startScheduled || inst.ended {
		inst.mu.Unlock()
		return
	}
	inst.startScheduled = true
	inst.startTime = when
	inst.mu.Unlock()

	inst.ctx.registerSource(inst)
}

// Stop implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoSourceSchedule) Stop(when webaudioapi.Seconds) {
	when = normalizeMalgoTime(when)

	inst.mu.Lock()
	if inst.ended {
		inst.mu.Unlock()
		return
	}
	inst.stopScheduled = true
	inst.stopTime = when
	inst.mu.Unlock()

	inst.ctx.registerSource(inst)
}

// SetOnEnded implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoSourceSchedule) SetOnEnded(handler func()) {
	inst.mu.Lock()
	inst.onEnded = handler
	ended := inst.ended
	inst.mu.Unlock()

	if ended && handler != nil {
		go handler()
	}
}

func (inst *malgoSourceSchedule) finish() {
	inst.mu.Lock()
	if inst.ended || !inst.started {
		inst.mu.Unlock()
		return
	}
	inst.ended = true
	handler := inst.onEnded
	inst.mu.Unlock()

	inst.ctx.unregisterSource(inst)
	if handler != nil {
		go handler()
	}
}

func (inst *malgoSourceSchedule) notifyEnded(handler func()) {
	inst.ctx.unregisterSource(inst)
	if handler != nil {
		go handler()
	}
}

type malgoOscillatorNode struct {
	malgoAudioNode
	schedule malgoSourceSchedule

	mu        sync.Mutex
	frequency *malgoAudioParam
	detune    *malgoAudioParam
	nodeType  webaudioapi.OscillatorType
	phase     float64
}

// GetFrequency implements [webaudioapi.OscillatorNode].
func (inst *malgoOscillatorNode) GetFrequency() webaudioapi.AudioParam {
	return inst.frequency
}

// GetDetune implements [webaudioapi.OscillatorNode].
func (inst *malgoOscillatorNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

// GetType implements [webaudioapi.OscillatorNode].
func (inst *malgoOscillatorNode) GetType() webaudioapi.OscillatorType {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.nodeType
}

// SetType implements [webaudioapi.OscillatorNode].
func (inst *malgoOscillatorNode) SetType(nodeType webaudioapi.OscillatorType) {
	inst.mu.Lock()
	inst.nodeType = nodeType
	inst.mu.Unlock()
}

// Start implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) Start(when webaudioapi.Seconds) {
	inst.schedule.Start(when)
}

// Stop implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) Stop(when webaudioapi.Seconds) {
	inst.schedule.Stop(when)
}

// SetOnEnded implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) SetOnEnded(handler func()) {
	inst.schedule.SetOnEnded(handler)
}

func (inst *malgoOscillatorNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	_ int,
	_ map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {

	output := [][]webaudioapi.SampleValue{make([]webaudioapi.SampleValue, frameCount)}
	schedule := &inst.schedule

	schedule.mu.Lock()
	if schedule.ended || !schedule.startScheduled {
		schedule.mu.Unlock()
		return output
	}

	inst.mu.Lock()
	phase := inst.phase
	nodeType := inst.nodeType
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
		if !schedule.started {
			continue
		}

		output[0][frame] = oscillatorSample(nodeType, phase)
		frequency := inst.frequency.valueAt(time)
		detune := inst.detune.valueAt(time)
		effectiveFrequency := frequency * math.Pow(2, detune/1200)
		phase += effectiveFrequency / float64(inst.ctx.sampleRate)
		phase -= math.Floor(phase)
	}
	inst.phase = phase
	inst.mu.Unlock()
	schedule.mu.Unlock()

	if ended {
		schedule.notifyEnded(handler)
	}
	return output
}

func oscillatorSample(nodeType webaudioapi.OscillatorType, phase float64) webaudioapi.SampleValue {
	switch nodeType {
	case webaudioapi.OscillatorTypeSquare:
		if phase < 0.5 {
			return 1
		}
		return -1
	case webaudioapi.OscillatorTypeSawtooth:
		return webaudioapi.SampleValue(2*phase - 1)
	case webaudioapi.OscillatorTypeTriangle:
		return webaudioapi.SampleValue(1 - 4*math.Abs(phase-0.5))
	case webaudioapi.OscillatorTypeSine, webaudioapi.OscillatorTypeCustom:
		fallthrough
	default:
		return webaudioapi.SampleValue(math.Sin(2 * math.Pi * phase))
	}
}

type malgoAudioBufferSourceNode struct {
	malgoAudioNode
	schedule malgoSourceSchedule

	mu           sync.Mutex
	buffer       webaudioapi.AudioBuffer
	playbackRate *malgoAudioParam
	detune       *malgoAudioParam
	position     float64
}

// GetBuffer implements [webaudioapi.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetBuffer() webaudioapi.AudioBuffer {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.buffer
}

// SetBuffer implements [webaudioapi.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) SetBuffer(buffer webaudioapi.AudioBuffer) {
	inst.mu.Lock()
	inst.buffer = buffer
	inst.mu.Unlock()
}

// GetPlaybackRate implements [webaudioapi.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetPlaybackRate() webaudioapi.AudioParam {
	return inst.playbackRate
}

// GetDetune implements [webaudioapi.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetDetune() webaudioapi.AudioParam {
	return inst.detune
}

// Start implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) Start(when webaudioapi.Seconds) {
	inst.schedule.Start(when)
}

// Stop implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) Stop(when webaudioapi.Seconds) {
	inst.schedule.Stop(when)
}

// SetOnEnded implements [webaudioapi.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) SetOnEnded(handler func()) {
	inst.schedule.SetOnEnded(handler)
}

func (inst *malgoAudioBufferSourceNode) renderSignal(
	startTime webaudioapi.Seconds,
	frameCount int,
	_ int,
	_ map[webaudioapi.AudioNode]bool,
) [][]webaudioapi.SampleValue {

	inst.mu.Lock()
	buffer := inst.buffer
	position := inst.position
	inst.mu.Unlock()

	channels := 1
	var length webaudioapi.SampleCount
	var bufferSampleRate webaudioapi.SampleRate
	var channelData [][]webaudioapi.SampleValue
	if buffer != nil {
		channels = buffer.GetNumberOfChannels()
		if channels <= 0 {
			channels = 0
		}
		length = buffer.GetLength()
		bufferSampleRate = buffer.GetSampleRate()
		channelData = make([][]webaudioapi.SampleValue, channels)
		for channel := 0; channel < channels; channel++ {
			channelData[channel] = buffer.GetChannelData(channel)
		}
	}
	if channels == 0 {
		channels = 1
	}
	if bufferSampleRate == 0 {
		bufferSampleRate = inst.ctx.sampleRate
	}

	output := make([][]webaudioapi.SampleValue, channels)
	for channel := range output {
		output[channel] = make([]webaudioapi.SampleValue, frameCount)
	}

	schedule := &inst.schedule
	schedule.mu.Lock()
	if schedule.ended || !schedule.startScheduled {
		schedule.mu.Unlock()
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
		if !schedule.started {
			continue
		}
		if buffer == nil {
			continue
		}
		if length == 0 || position >= float64(length) {
			schedule.ended = true
			ended = true
			handler = schedule.onEnded
			break
		}

		index := int(position)
		fraction := position - float64(index)
		for channel := 0; channel < len(channelData); channel++ {
			data := channelData[channel]
			if index >= len(data) {
				continue
			}
			value := float64(data[index])
			if index+1 < len(data) {
				value += (float64(data[index+1]) - value) * fraction
			}
			output[channel][frame] = webaudioapi.SampleValue(value)
		}

		playbackRate := inst.playbackRate.valueAt(time)
		detune := inst.detune.valueAt(time)
		step := playbackRate * math.Pow(2, detune/1200)
		step *= float64(bufferSampleRate) / float64(inst.ctx.sampleRate)
		position += step
	}

	inst.mu.Lock()
	inst.position = position
	inst.mu.Unlock()
	schedule.mu.Unlock()

	if ended {
		schedule.notifyEnded(handler)
	}
	return output
}

////////////////////////////////////////////////////////////////////////////////
// AudioBuffer

type malgoAudioBuffer struct {
	sampleRate webaudioapi.SampleRate
	channels   [][]webaudioapi.SampleValue
}

// GetSampleRate implements [webaudioapi.AudioBuffer].
func (inst *malgoAudioBuffer) GetSampleRate() webaudioapi.SampleRate {
	return inst.sampleRate
}

// GetLength implements [webaudioapi.AudioBuffer].
func (inst *malgoAudioBuffer) GetLength() webaudioapi.SampleCount {
	if len(inst.channels) == 0 {
		return 0
	}
	return webaudioapi.SampleCount(len(inst.channels[0]))
}

// GetNumberOfChannels implements [webaudioapi.AudioBuffer].
func (inst *malgoAudioBuffer) GetNumberOfChannels() int {
	return len(inst.channels)
}

// GetChannelData implements [webaudioapi.AudioBuffer].
func (inst *malgoAudioBuffer) GetChannelData(channel int) []webaudioapi.SampleValue {
	if channel < 0 || channel >= len(inst.channels) {
		return nil
	}
	return inst.channels[channel]
}

////////////////////////////////////////////////////////////////////////////////
// AudioParam

type malgoAudioParamEventKind uint8

const (
	malgoAudioParamEventSet malgoAudioParamEventKind = iota
	malgoAudioParamEventLinearRamp
	malgoAudioParamEventExponentialRamp
	malgoAudioParamEventTarget
)

type malgoAudioParamEvent struct {
	kind         malgoAudioParamEventKind
	value        float64
	time         webaudioapi.Seconds
	timeConstant webaudioapi.Seconds
}

type malgoAudioParam struct {
	mu sync.Mutex

	baseValue float64
	baseTime  webaudioapi.Seconds
	events    []malgoAudioParamEvent
	now       func() webaudioapi.Seconds
}

func newMalgoAudioParam(
	value float64,
	baseTime webaudioapi.Seconds,
	now func() webaudioapi.Seconds,
) *malgoAudioParam {
	if now == nil {
		now = func() webaudioapi.Seconds { return 0 }
	}
	return &malgoAudioParam{
		baseValue: value,
		baseTime:  baseTime,
		now:       now,
	}
}

// GetValue implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) GetValue() float64 {
	return inst.valueAt(inst.now())
}

// SetValue implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) SetValue(value float64) {
	now := inst.now()
	inst.mu.Lock()
	inst.baseValue = value
	inst.baseTime = now
	inst.events = nil
	inst.mu.Unlock()
}

// SetValueAtTime implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) SetValueAtTime(value float64, startTime webaudioapi.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventSet,
		value: value,
		time:  normalizeMalgoTime(startTime),
	})
}

// LinearRampToValueAtTime implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) LinearRampToValueAtTime(value float64, endTime webaudioapi.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventLinearRamp,
		value: value,
		time:  normalizeMalgoTime(endTime),
	})
}

// ExponentialRampToValueAtTime implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) ExponentialRampToValueAtTime(value float64, endTime webaudioapi.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventExponentialRamp,
		value: value,
		time:  normalizeMalgoTime(endTime),
	})
}

// SetTargetAtTime implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) SetTargetAtTime(
	target float64,
	startTime webaudioapi.Seconds,
	timeConstant webaudioapi.Seconds,
) {
	inst.addEvent(malgoAudioParamEvent{
		kind:         malgoAudioParamEventTarget,
		value:        target,
		time:         normalizeMalgoTime(startTime),
		timeConstant: normalizeMalgoTime(timeConstant),
	})
}

// CancelScheduledValues implements [webaudioapi.AudioParam].
func (inst *malgoAudioParam) CancelScheduledValues(startTime webaudioapi.Seconds) {
	startTime = normalizeMalgoTime(startTime)
	now := inst.now()

	inst.mu.Lock()
	defer inst.mu.Unlock()

	inst.baseValue = inst.valueAtLocked(now)
	inst.baseTime = now

	events := inst.events[:0]
	for _, event := range inst.events {
		if event.time > now && event.time < startTime {
			events = append(events, event)
		}
	}
	inst.events = events
}

func (inst *malgoAudioParam) addEvent(event malgoAudioParamEvent) {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	index := len(inst.events)
	for i, item := range inst.events {
		if item.time > event.time {
			index = i
			break
		}
	}

	inst.events = append(inst.events, malgoAudioParamEvent{})
	copy(inst.events[index+1:], inst.events[index:])
	inst.events[index] = event
}

func (inst *malgoAudioParam) valueAt(time webaudioapi.Seconds) float64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.valueAtLocked(time)
}

func (inst *malgoAudioParam) valueAtLocked(time webaudioapi.Seconds) float64 {
	value := inst.baseValue
	cursor := inst.baseTime

	mode := malgoAudioParamEventSet
	target := value
	timeConstant := webaudioapi.Seconds(0)

	for _, event := range inst.events {
		if time < event.time {
			switch event.kind {
			case malgoAudioParamEventLinearRamp:
				return interpolateMalgoValue(cursor, value, event.time, event.value, time)
			case malgoAudioParamEventExponentialRamp:
				return interpolateMalgoExponentialValue(
					cursor,
					value,
					event.time,
					event.value,
					time,
				)
			default:
				return evaluateMalgoAudioParam(cursor, value, time, mode, target, timeConstant)
			}
		}

		value = evaluateMalgoAudioParam(cursor, value, event.time, mode, target, timeConstant)
		cursor = event.time

		switch event.kind {
		case malgoAudioParamEventTarget:
			mode = malgoAudioParamEventTarget
			target = event.value
			timeConstant = event.timeConstant
		default:
			mode = malgoAudioParamEventSet
			value = event.value
		}
	}

	return evaluateMalgoAudioParam(cursor, value, time, mode, target, timeConstant)
}

func evaluateMalgoAudioParam(
	startTime webaudioapi.Seconds,
	startValue float64,
	time webaudioapi.Seconds,
	mode malgoAudioParamEventKind,
	target float64,
	timeConstant webaudioapi.Seconds,
) float64 {

	if time <= startTime {
		return startValue
	}
	if mode != malgoAudioParamEventTarget {
		return startValue
	}
	if timeConstant <= 0 {
		return target
	}

	elapsed := time - startTime
	factor := 1 - math.Exp(-float64(elapsed/timeConstant))
	return startValue + (target-startValue)*factor
}

func interpolateMalgoValue(
	startTime webaudioapi.Seconds,
	startValue float64,
	endTime webaudioapi.Seconds,
	endValue float64,
	time webaudioapi.Seconds,
) float64 {

	if endTime <= startTime {
		return endValue
	}
	ratio := float64((time - startTime) / (endTime - startTime))
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	return startValue + (endValue-startValue)*ratio
}

func interpolateMalgoExponentialValue(
	startTime webaudioapi.Seconds,
	startValue float64,
	endTime webaudioapi.Seconds,
	endValue float64,
	time webaudioapi.Seconds,
) float64 {

	if startValue <= 0 || endValue <= 0 || endTime <= startTime {
		return interpolateMalgoValue(startTime, startValue, endTime, endValue, time)
	}
	ratio := float64((time - startTime) / (endTime - startTime))
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	return startValue * math.Pow(endValue/startValue, ratio)
}

func normalizeMalgoTime(value webaudioapi.Seconds) webaudioapi.Seconds {
	if value < 0 || math.IsNaN(float64(value)) {
		return 0
	}
	return value
}

////////////////////////////////////////////////////////////////////////////////
// EOF
