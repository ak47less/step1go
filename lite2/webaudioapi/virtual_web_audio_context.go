package webaudioapi

type BaseAudioContext struct {
	Implementation AudioDriverImplementation
}

// AudioContext 是对 Web Audio API 中的 AudioContext 对象的一比一模仿;
// 在 Javascript 版本的代码中, 这个结构将被原生的 AudioContext 对象 直接替换.
type AudioContext struct {
	BaseAudioContext
}

func NewAudioContext(im AudioDriverImplementation) *AudioContext {
	ac := new(AudioContext)
	ac.Implementation = im
	return ac
}
