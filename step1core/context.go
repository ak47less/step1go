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

	Wires []*Wire

	State EngineState

	DeviceFactory DeviceFactory // the main Device factory

	DeviceManager DeviceManager

	DeviceRegistry DeviceRegistry

	LifeManager LifeManager

	TrackManager TrackManager

	Logger Logger
}

type Bus struct {
	Context *EngineContext

	Current TimePoint

	Router Router

	Audio *AudioBuffer

	MIDI *MidiBuffer
}
