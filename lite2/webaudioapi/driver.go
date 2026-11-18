package webaudioapi

type Driver interface {
	GetFactory() AudioContextFactory
}

type AudioContextFactory interface {
	CreateAudioContext() (AudioContext, error)
}
