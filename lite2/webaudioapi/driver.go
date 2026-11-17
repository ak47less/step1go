package webaudioapi

type Driver interface {
	GetFactory() AudioContextFactory
}

type AudioContextFactory interface {
	CreateAudioContext() (*AudioContext, error)
}

// 这个接口 接受模仿对象的委托, 转发给真正的驱动实现
type AudioDriverImplementation interface {

	// todo: methods to implement API
}
