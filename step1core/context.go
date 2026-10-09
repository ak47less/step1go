package step1core

type EngineContext struct {
	Configuration *EngineConfiguration
	ConfigRedo    *EngineConfiguration // 需要重新配置的信息 , 否则为 nil

	// time-lines
	RootTimeline    *Timeline // timeline for engine
	ProjectTimeline *Timeline // timeline for document (project)

	Engine Engine

	Loader EngineLoader

	Transport Transport

	Looper *Looper

	Clock Clock

	State EngineState

	ElementFactory ElementFactory // the main element factory

	ElementManager ElementManager

	ElementRegistry ElementRegistry

	LifeManager LifeManager

	TrackManager TrackManager

	Logger *Logger
}

type Bus struct {
	Context *EngineContext

	Current TimePoint

	Audio *AudioBuffer

	MIDI *MidiBuffer
}
