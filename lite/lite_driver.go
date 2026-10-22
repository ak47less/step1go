package lite

// Driver 提供创建 Web Audio API 上下文所需的底层实现。
type Driver interface {
	GetAudioContextFactory() AudioContextFactory
}

// AudioNode 是 Web Audio API 中所有音频节点的公共抽象。
type AudioNode interface {
	Connect(destination AudioNode)
	Disconnect()
}

// AudioParam 表示音频节点上一个可以被自动化控制的值。
type AudioParam interface {
	GetValue() float64
	SetValue(value float64)

	SetValueAtTime(value float64, startTime Seconds)
	LinearRampToValueAtTime(value float64, endTime Seconds)
	ExponentialRampToValueAtTime(value float64, endTime Seconds)
	SetTargetAtTime(target float64, startTime Seconds, timeConstant Seconds)
	CancelScheduledValues(startTime Seconds)
}

// AudioScheduledSourceNode 表示可以在指定时间开始和停止的音频源节点。
type AudioScheduledSourceNode interface {
	AudioNode

	Start(when Seconds)
	Stop(when Seconds)
	SetOnEnded(handler func())
}

// OscillatorType 表示振荡器产生的波形。
type OscillatorType string

const (
	OscillatorTypeSine     OscillatorType = "sine"
	OscillatorTypeSquare   OscillatorType = "square"
	OscillatorTypeSawtooth OscillatorType = "sawtooth"
	OscillatorTypeTriangle OscillatorType = "triangle"
	OscillatorTypeCustom   OscillatorType = "custom"
)

// OscillatorNode 表示一个周期波形音频源。
type OscillatorNode interface {
	AudioScheduledSourceNode

	GetFrequency() AudioParam
	GetDetune() AudioParam
	GetType() OscillatorType
	SetType(nodeType OscillatorType)
}

// GainNode 表示一个用于调整音频信号增益的节点。
type GainNode interface {
	AudioNode

	GetGain() AudioParam
}

// AudioDestinationNode 表示音频上下文最终输出的目标节点。
type AudioDestinationNode interface {
	AudioNode

	GetMaxChannelCount() int
}

// AudioBuffer 表示内存中的多声道 PCM 音频数据。
type AudioBuffer interface {
	GetSampleRate() SampleRate
	GetLength() SampleCount
	GetNumberOfChannels() int
	GetChannelData(channel int) []SampleValue
}

// AudioBufferSourceNode 表示从 AudioBuffer 读取数据的音频源节点。
type AudioBufferSourceNode interface {
	AudioScheduledSourceNode

	GetBuffer() AudioBuffer
	SetBuffer(buffer AudioBuffer)
	GetPlaybackRate() AudioParam
	GetDetune() AudioParam
}

// AudioContextState 表示 AudioContext 的生命周期状态。
type AudioContextState string

const (
	AudioContextStateSuspended AudioContextState = "suspended"
	AudioContextStateRunning   AudioContextState = "running"
	AudioContextStateClosed    AudioContextState = "closed"
)

// AudioContext 接口是对 Web Audio API (js) 中的 AudioContext 接口的抽象。
type AudioContext interface {
	Close()

	Resume()

	Suspend()

	CreateOscillator() OscillatorNode

	CreateGain() GainNode

	CreateBuffer(numberOfChannels int, length SampleCount, sampleRate SampleRate) AudioBuffer

	CreateBufferSource() AudioBufferSourceNode

	// getter

	GetCurrentTime() Seconds

	GetSampleRate() SampleRate

	GetDestination() AudioDestinationNode

	GetState() AudioContextState
}

// AudioContextFactory 创建由具体音频后端实现的 AudioContext。
type AudioContextFactory interface {
	NewAudioContext() AudioContext
}
