package malgo4lite

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ak47less/step1go/lite"
	"github.com/gen2brain/malgo"
)

const (
	malgoDefaultSampleRate   lite.SampleRate = 44100
	malgoDefaultChannelCount                 = 2
	malgoDefaultFrequency                    = 440.0
	malgoDefaultGain                         = 1.0
	malgoDefaultPlaybackRate                 = 1.0
	malgoMaxRenderDepth                      = 128
)

var (
	_ lite.Driver                   = new(MalgoDriver)
	_ lite.AudioContextFactory      = new(MalgoAudioContextFactory)
	_ lite.AudioContext             = new(MalgoAudioContext)
	_ lite.AudioNode                = new(malgoAudioNode)
	_ lite.AudioParam               = new(malgoAudioParam)
	_ lite.OscillatorNode           = new(malgoOscillatorNode)
	_ lite.GainNode                 = new(malgoGainNode)
	_ lite.AudioDestinationNode     = new(malgoAudioDestinationNode)
	_ lite.AudioBuffer              = new(malgoAudioBuffer)
	_ lite.AudioBufferSourceNode    = new(malgoAudioBufferSourceNode)
	_ lite.AudioScheduledSourceNode = new(malgoOscillatorNode)

	errMalgoContextClosed = errors.New("malgo: audio context is closed")
)

////////////////////////////////////////////////////////////////////////////////
// class : MalgoDriver

// MalgoDriver 提供基于 miniaudio/malgo 的音频上下文。
type MalgoDriver struct {
}

// GetAudioContextFactory implements [lite.Driver].
func (inst *MalgoDriver) GetAudioContextFactory() lite.AudioContextFactory {
	return new(MalgoAudioContextFactory)
}

func (inst *MalgoDriver) _impl() lite.Driver {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class : MalgoAudioContextFactory

// MalgoAudioContextFactory 创建共享 malgo 后端的音频上下文。
type MalgoAudioContextFactory struct {
}

// NewAudioContext implements [lite.AudioContextFactory].
func (inst *MalgoAudioContextFactory) NewAudioContext() lite.AudioContext {
	mc, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		panic(err)
	}
	return newMalgoAudioContext(mc)
}

func (inst *MalgoAudioContextFactory) _impl() lite.AudioContextFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// class: MalgoAudioContext

// MalgoAudioContext 实现 [lite.AudioContext]. 播放设备在第一次 Resume 时创建，
// 因而可以离屏创建音频图和缓冲区。
type MalgoAudioContext struct {
	mc *malgo.AllocatedContext

	// lifecycleMu 串行化设备启动、停止和释放操作。
	lifecycleMu sync.Mutex
	// mu 同时保护上下文状态和音频图。
	mu sync.RWMutex
	// sourcesMu 单独保护源集合，使音频回调结束时不会与渲染锁互相等待。
	sourcesMu sync.Mutex

	state        lite.AudioContextState
	sampleRate   lite.SampleRate
	elapsed      lite.Seconds
	runningSince time.Time
	closed       atomic.Bool

	device      *malgo.Device
	destination *malgoAudioDestinationNode
	inputs      map[lite.AudioNode][]lite.AudioNode
	sources     map[*malgoSourceSchedule]struct{}
}

func newMalgoAudioContext(mc *malgo.AllocatedContext) *MalgoAudioContext {
	inst := &MalgoAudioContext{
		mc:         mc,
		state:      lite.AudioContextStateSuspended,
		sampleRate: malgoDefaultSampleRate,
		inputs:     make(map[lite.AudioNode][]lite.AudioNode),
		sources:    make(map[*malgoSourceSchedule]struct{}),
	}
	inst.destination = newMalgoAudioDestinationNode(inst)
	return inst
}

// Close implements [lite.AudioContext].
func (inst *MalgoAudioContext) Close() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.RLock()
	closed := inst.state == lite.AudioContextStateClosed
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
	inst.state = lite.AudioContextStateClosed
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

// Resume implements [lite.AudioContext].
func (inst *MalgoAudioContext) Resume() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.RLock()
	state := inst.state
	inst.mu.RUnlock()
	if state == lite.AudioContextStateClosed || state == lite.AudioContextStateRunning {
		return
	}

	device, err := inst.ensureDevice()
	if err != nil {
		panic(err)
	}

	inst.mu.Lock()
	inst.state = lite.AudioContextStateRunning
	inst.runningSince = time.Now()
	inst.mu.Unlock()

	if err = device.Start(); err != nil {
		inst.mu.Lock()
		inst.freezeClockLocked()
		inst.state = lite.AudioContextStateSuspended
		inst.mu.Unlock()
		panic(err)
	}
}

// Suspend implements [lite.AudioContext].
func (inst *MalgoAudioContext) Suspend() {
	inst.lifecycleMu.Lock()
	defer inst.lifecycleMu.Unlock()

	inst.mu.Lock()
	if inst.state == lite.AudioContextStateClosed {
		inst.mu.Unlock()
		return
	}
	if inst.state == lite.AudioContextStateRunning {
		inst.freezeClockLocked()
		inst.state = lite.AudioContextStateSuspended
	}
	device := inst.device
	inst.mu.Unlock()

	if device != nil && device.IsStarted() {
		_ = device.Stop()
	}
}

// CreateOscillator implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateOscillator() lite.OscillatorNode {
	now := inst.GetCurrentTime()
	node := &malgoOscillatorNode{
		frequency: newMalgoAudioParam(malgoDefaultFrequency, now, inst.GetCurrentTime),
		detune:    newMalgoAudioParam(0, now, inst.GetCurrentTime),
		nodeType:  lite.OscillatorTypeSine,
	}
	node.init(inst, node)
	node.schedule.ctx = inst
	return node
}

// CreateGain implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateGain() lite.GainNode {
	node := &malgoGainNode{
		gain: newMalgoAudioParam(malgoDefaultGain, inst.GetCurrentTime(), inst.GetCurrentTime),
	}
	node.init(inst, node)
	return node
}

// CreateBuffer implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateBuffer(
	numberOfChannels int,
	length lite.SampleCount,
	sampleRate lite.SampleRate,
) lite.AudioBuffer {

	if numberOfChannels < 0 {
		numberOfChannels = 0
	}
	if sampleRate == 0 {
		sampleRate = inst.GetSampleRate()
	}

	maxInt := int(^uint(0) >> 1)
	if length > lite.SampleCount(maxInt) {
		length = 0
	}

	channels := make([][]lite.SampleValue, numberOfChannels)
	for i := range channels {
		channels[i] = make([]lite.SampleValue, int(length))
	}
	return &malgoAudioBuffer{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// CreateBufferSource implements [lite.AudioContext].
func (inst *MalgoAudioContext) CreateBufferSource() lite.AudioBufferSourceNode {
	now := inst.GetCurrentTime()
	node := &malgoAudioBufferSourceNode{
		playbackRate: newMalgoAudioParam(malgoDefaultPlaybackRate, now, inst.GetCurrentTime),
		detune:       newMalgoAudioParam(0, now, inst.GetCurrentTime),
	}
	node.init(inst, node)
	node.schedule.ctx = inst
	return node
}

// GetCurrentTime implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetCurrentTime() lite.Seconds {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.currentTimeLocked()
}

// GetSampleRate implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetSampleRate() lite.SampleRate {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.sampleRate
}

// GetDestination implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetDestination() lite.AudioDestinationNode {
	return inst.destination
}

// GetState implements [lite.AudioContext].
func (inst *MalgoAudioContext) GetState() lite.AudioContextState {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.state
}

func (inst *MalgoAudioContext) _impl() lite.AudioContext {
	return inst
}

func (inst *MalgoAudioContext) ensureDevice() (*malgo.Device, error) {
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

func (inst *MalgoAudioContext) onData(output []byte, _ []byte, frameCount uint32) {
	if len(output) < 4 {
		return
	}

	samples := unsafe.Slice(
		(*float32)(unsafe.Pointer(unsafe.SliceData(output))),
		len(output)/4,
	)

	inst.mu.RLock()
	if inst.state != lite.AudioContextStateRunning {
		inst.mu.RUnlock()
		clear(samples)
		return
	}

	startTime := inst.currentTimeLocked()
	signal := inst.destination.renderSignal(
		startTime,
		int(frameCount),
		0,
		make(map[lite.AudioNode]bool),
	)
	inst.mu.RUnlock()

	for i := range samples {
		frame := i / malgoDefaultChannelCount
		channel := i % malgoDefaultChannelCount
		var value lite.SampleValue
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

func (inst *MalgoAudioContext) currentTimeLocked() lite.Seconds {
	now := inst.elapsed
	if inst.state == lite.AudioContextStateRunning {
		now += lite.Seconds(time.Since(inst.runningSince).Seconds())
	}
	return now
}

func (inst *MalgoAudioContext) freezeClockLocked() {
	if inst.state != lite.AudioContextStateRunning {
		return
	}
	inst.elapsed += lite.Seconds(time.Since(inst.runningSince).Seconds())
	inst.runningSince = time.Time{}
}

func (inst *MalgoAudioContext) connect(source lite.AudioNode, destination lite.AudioNode) {
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
	if inst.state == lite.AudioContextStateClosed {
		return
	}
	for _, item := range inst.inputs[destination] {
		if item == source {
			return
		}
	}
	inst.inputs[destination] = append(inst.inputs[destination], source)
}

func (inst *MalgoAudioContext) disconnect(source lite.AudioNode) {
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

func (inst *MalgoAudioContext) registerSource(source *malgoSourceSchedule) {
	inst.sourcesMu.Lock()
	if !inst.closed.Load() {
		inst.sources[source] = struct{}{}
	}
	inst.sourcesMu.Unlock()
}

func (inst *MalgoAudioContext) unregisterSource(source *malgoSourceSchedule) {
	inst.sourcesMu.Lock()
	delete(inst.sources, source)
	inst.sourcesMu.Unlock()
}

func (inst *MalgoAudioContext) renderNode(
	node lite.AudioNode,
	startTime lite.Seconds,
	frameCount int,
	depth int,
	visiting map[lite.AudioNode]bool,
) [][]lite.SampleValue {

	if node == nil || frameCount <= 0 || depth > malgoMaxRenderDepth || visiting[node] {
		return nil
	}
	renderer, ok := node.(interface {
		renderSignal(
			lite.Seconds,
			int,
			int,
			map[lite.AudioNode]bool,
		) [][]lite.SampleValue
	})
	if !ok {
		return nil
	}

	visiting[node] = true
	defer delete(visiting, node)
	return renderer.renderSignal(startTime, frameCount, depth, visiting)
}

func (inst *MalgoAudioContext) renderInputs(
	node lite.AudioNode,
	startTime lite.Seconds,
	frameCount int,
	depth int,
	visiting map[lite.AudioNode]bool,
) [][][]lite.SampleValue {
	inputs := inst.inputs[node]
	signals := make([][][]lite.SampleValue, 0, len(inputs))
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
	ctx   *MalgoAudioContext
	owner lite.AudioNode
}

type malgoAudioNodeProvider interface {
	malgoAudioNodeBase() *malgoAudioNode
}

func (inst *malgoAudioNode) init(ctx *MalgoAudioContext, owner lite.AudioNode) {
	inst.ctx = ctx
	inst.owner = owner
}

func (inst *malgoAudioNode) malgoAudioNodeBase() *malgoAudioNode {
	return inst
}

// Connect implements [lite.AudioNode].
func (inst *malgoAudioNode) Connect(destination lite.AudioNode) {
	if destination != nil {
		inst.ctx.connect(inst.owner, destination)
	}
}

// Disconnect implements [lite.AudioNode].
func (inst *malgoAudioNode) Disconnect() {
	inst.ctx.disconnect(inst.owner)
}

type malgoGainNode struct {
	malgoAudioNode
	gain *malgoAudioParam
}

// GetGain implements [lite.GainNode].
func (inst *malgoGainNode) GetGain() lite.AudioParam {
	return inst.gain
}

func (inst *malgoGainNode) renderSignal(
	startTime lite.Seconds,
	frameCount int,
	depth int,
	visiting map[lite.AudioNode]bool,
) [][]lite.SampleValue {

	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	channels := signalChannelCount(inputs)
	if channels == 0 {
		channels = 1
	}
	output := mixSignals(inputs, frameCount, channels)

	for channel := range output {
		for frame := 0; frame < frameCount; frame++ {
			time := startTime + lite.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
			output[channel][frame] *= lite.SampleValue(inst.gain.valueAt(time))
		}
	}
	return output
}

type malgoAudioDestinationNode struct {
	malgoAudioNode
}

func newMalgoAudioDestinationNode(ctx *MalgoAudioContext) *malgoAudioDestinationNode {
	inst := new(malgoAudioDestinationNode)
	inst.init(ctx, inst)
	return inst
}

// GetMaxChannelCount implements [lite.AudioDestinationNode].
func (inst *malgoAudioDestinationNode) GetMaxChannelCount() int {
	return malgoDefaultChannelCount
}

func (inst *malgoAudioDestinationNode) renderSignal(
	startTime lite.Seconds,
	frameCount int,
	depth int,
	visiting map[lite.AudioNode]bool,
) [][]lite.SampleValue {

	inputs := inst.ctx.renderInputs(inst.owner, startTime, frameCount, depth+1, visiting)
	return mixSignals(inputs, frameCount, malgoDefaultChannelCount)
}

func signalChannelCount(signals [][][]lite.SampleValue) int {
	result := 0
	for _, signal := range signals {
		if len(signal) > result {
			result = len(signal)
		}
	}
	return result
}

func mixSignals(signals [][][]lite.SampleValue, frameCount int, channels int) [][]lite.SampleValue {
	output := make([][]lite.SampleValue, channels)
	for channel := range output {
		output[channel] = make([]lite.SampleValue, frameCount)
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
	ctx *MalgoAudioContext

	mu sync.Mutex

	startScheduled bool
	startTime      lite.Seconds
	started        bool

	stopScheduled bool
	stopTime      lite.Seconds
	ended         bool

	onEnded func()
}

// Start implements [lite.AudioScheduledSourceNode].
func (inst *malgoSourceSchedule) Start(when lite.Seconds) {
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

// Stop implements [lite.AudioScheduledSourceNode].
func (inst *malgoSourceSchedule) Stop(when lite.Seconds) {
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

// SetOnEnded implements [lite.AudioScheduledSourceNode].
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
	nodeType  lite.OscillatorType
	phase     float64
}

// GetFrequency implements [lite.OscillatorNode].
func (inst *malgoOscillatorNode) GetFrequency() lite.AudioParam {
	return inst.frequency
}

// GetDetune implements [lite.OscillatorNode].
func (inst *malgoOscillatorNode) GetDetune() lite.AudioParam {
	return inst.detune
}

// GetType implements [lite.OscillatorNode].
func (inst *malgoOscillatorNode) GetType() lite.OscillatorType {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.nodeType
}

// SetType implements [lite.OscillatorNode].
func (inst *malgoOscillatorNode) SetType(nodeType lite.OscillatorType) {
	inst.mu.Lock()
	inst.nodeType = nodeType
	inst.mu.Unlock()
}

// Start implements [lite.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) Start(when lite.Seconds) {
	inst.schedule.Start(when)
}

// Stop implements [lite.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) Stop(when lite.Seconds) {
	inst.schedule.Stop(when)
}

// SetOnEnded implements [lite.AudioScheduledSourceNode].
func (inst *malgoOscillatorNode) SetOnEnded(handler func()) {
	inst.schedule.SetOnEnded(handler)
}

func (inst *malgoOscillatorNode) renderSignal(
	startTime lite.Seconds,
	frameCount int,
	_ int,
	_ map[lite.AudioNode]bool,
) [][]lite.SampleValue {

	output := [][]lite.SampleValue{make([]lite.SampleValue, frameCount)}
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
		time := startTime + lite.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
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

func oscillatorSample(nodeType lite.OscillatorType, phase float64) lite.SampleValue {
	switch nodeType {
	case lite.OscillatorTypeSquare:
		if phase < 0.5 {
			return 1
		}
		return -1
	case lite.OscillatorTypeSawtooth:
		return lite.SampleValue(2*phase - 1)
	case lite.OscillatorTypeTriangle:
		return lite.SampleValue(1 - 4*math.Abs(phase-0.5))
	case lite.OscillatorTypeSine, lite.OscillatorTypeCustom:
		fallthrough
	default:
		return lite.SampleValue(math.Sin(2 * math.Pi * phase))
	}
}

type malgoAudioBufferSourceNode struct {
	malgoAudioNode
	schedule malgoSourceSchedule

	mu           sync.Mutex
	buffer       lite.AudioBuffer
	playbackRate *malgoAudioParam
	detune       *malgoAudioParam
	position     float64
}

// GetBuffer implements [lite.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetBuffer() lite.AudioBuffer {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.buffer
}

// SetBuffer implements [lite.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) SetBuffer(buffer lite.AudioBuffer) {
	inst.mu.Lock()
	inst.buffer = buffer
	inst.mu.Unlock()
}

// GetPlaybackRate implements [lite.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetPlaybackRate() lite.AudioParam {
	return inst.playbackRate
}

// GetDetune implements [lite.AudioBufferSourceNode].
func (inst *malgoAudioBufferSourceNode) GetDetune() lite.AudioParam {
	return inst.detune
}

// Start implements [lite.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) Start(when lite.Seconds) {
	inst.schedule.Start(when)
}

// Stop implements [lite.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) Stop(when lite.Seconds) {
	inst.schedule.Stop(when)
}

// SetOnEnded implements [lite.AudioScheduledSourceNode].
func (inst *malgoAudioBufferSourceNode) SetOnEnded(handler func()) {
	inst.schedule.SetOnEnded(handler)
}

func (inst *malgoAudioBufferSourceNode) renderSignal(
	startTime lite.Seconds,
	frameCount int,
	_ int,
	_ map[lite.AudioNode]bool,
) [][]lite.SampleValue {

	inst.mu.Lock()
	buffer := inst.buffer
	position := inst.position
	inst.mu.Unlock()

	channels := 1
	var length lite.SampleCount
	var bufferSampleRate lite.SampleRate
	var channelData [][]lite.SampleValue
	if buffer != nil {
		channels = buffer.GetNumberOfChannels()
		if channels <= 0 {
			channels = 0
		}
		length = buffer.GetLength()
		bufferSampleRate = buffer.GetSampleRate()
		channelData = make([][]lite.SampleValue, channels)
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

	output := make([][]lite.SampleValue, channels)
	for channel := range output {
		output[channel] = make([]lite.SampleValue, frameCount)
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
		time := startTime + lite.Seconds(float64(frame)/float64(inst.ctx.sampleRate))
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
			output[channel][frame] = lite.SampleValue(value)
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
	sampleRate lite.SampleRate
	channels   [][]lite.SampleValue
}

// GetSampleRate implements [lite.AudioBuffer].
func (inst *malgoAudioBuffer) GetSampleRate() lite.SampleRate {
	return inst.sampleRate
}

// GetLength implements [lite.AudioBuffer].
func (inst *malgoAudioBuffer) GetLength() lite.SampleCount {
	if len(inst.channels) == 0 {
		return 0
	}
	return lite.SampleCount(len(inst.channels[0]))
}

// GetNumberOfChannels implements [lite.AudioBuffer].
func (inst *malgoAudioBuffer) GetNumberOfChannels() int {
	return len(inst.channels)
}

// GetChannelData implements [lite.AudioBuffer].
func (inst *malgoAudioBuffer) GetChannelData(channel int) []lite.SampleValue {
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
	time         lite.Seconds
	timeConstant lite.Seconds
}

type malgoAudioParam struct {
	mu sync.Mutex

	baseValue float64
	baseTime  lite.Seconds
	events    []malgoAudioParamEvent
	now       func() lite.Seconds
}

func newMalgoAudioParam(
	value float64,
	baseTime lite.Seconds,
	now func() lite.Seconds,
) *malgoAudioParam {
	if now == nil {
		now = func() lite.Seconds { return 0 }
	}
	return &malgoAudioParam{
		baseValue: value,
		baseTime:  baseTime,
		now:       now,
	}
}

// GetValue implements [lite.AudioParam].
func (inst *malgoAudioParam) GetValue() float64 {
	return inst.valueAt(inst.now())
}

// SetValue implements [lite.AudioParam].
func (inst *malgoAudioParam) SetValue(value float64) {
	now := inst.now()
	inst.mu.Lock()
	inst.baseValue = value
	inst.baseTime = now
	inst.events = nil
	inst.mu.Unlock()
}

// SetValueAtTime implements [lite.AudioParam].
func (inst *malgoAudioParam) SetValueAtTime(value float64, startTime lite.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventSet,
		value: value,
		time:  normalizeMalgoTime(startTime),
	})
}

// LinearRampToValueAtTime implements [lite.AudioParam].
func (inst *malgoAudioParam) LinearRampToValueAtTime(value float64, endTime lite.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventLinearRamp,
		value: value,
		time:  normalizeMalgoTime(endTime),
	})
}

// ExponentialRampToValueAtTime implements [lite.AudioParam].
func (inst *malgoAudioParam) ExponentialRampToValueAtTime(value float64, endTime lite.Seconds) {
	inst.addEvent(malgoAudioParamEvent{
		kind:  malgoAudioParamEventExponentialRamp,
		value: value,
		time:  normalizeMalgoTime(endTime),
	})
}

// SetTargetAtTime implements [lite.AudioParam].
func (inst *malgoAudioParam) SetTargetAtTime(
	target float64,
	startTime lite.Seconds,
	timeConstant lite.Seconds,
) {
	inst.addEvent(malgoAudioParamEvent{
		kind:         malgoAudioParamEventTarget,
		value:        target,
		time:         normalizeMalgoTime(startTime),
		timeConstant: normalizeMalgoTime(timeConstant),
	})
}

// CancelScheduledValues implements [lite.AudioParam].
func (inst *malgoAudioParam) CancelScheduledValues(startTime lite.Seconds) {
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

func (inst *malgoAudioParam) valueAt(time lite.Seconds) float64 {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.valueAtLocked(time)
}

func (inst *malgoAudioParam) valueAtLocked(time lite.Seconds) float64 {
	value := inst.baseValue
	cursor := inst.baseTime

	mode := malgoAudioParamEventSet
	target := value
	timeConstant := lite.Seconds(0)

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
	startTime lite.Seconds,
	startValue float64,
	time lite.Seconds,
	mode malgoAudioParamEventKind,
	target float64,
	timeConstant lite.Seconds,
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
	startTime lite.Seconds,
	startValue float64,
	endTime lite.Seconds,
	endValue float64,
	time lite.Seconds,
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
	startTime lite.Seconds,
	startValue float64,
	endTime lite.Seconds,
	endValue float64,
	time lite.Seconds,
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

func normalizeMalgoTime(value lite.Seconds) lite.Seconds {
	if value < 0 || math.IsNaN(float64(value)) {
		return 0
	}
	return value
}

////////////////////////////////////////////////////////////////////////////////
// EOF
