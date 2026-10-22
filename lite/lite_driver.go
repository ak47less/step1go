package lite

// todo : 在这个 go 文件中,  按照 Web Audio API, 创建对应的抽象接口

type Driver interface {
	GetAudioContextFactory() AudioContextFactory
}

type AudioNode interface {
	Connect(destination AudioNode)
}

type OscillatorNode interface {
	AudioNode
}

// AudioContext 接口是对 Web Audio API (js) 中的 AudioContext 接口的抽象
type AudioContext interface {
	Close()

	Resume()

	Suspend()

	CreateOscillator() OscillatorNode

	// getter

	GetCurrentTime() Seconds
}

type AudioContextFactory interface {
	NewAudioContext() AudioContext
}
