package step1core

type EngineContext struct {

	// config

	Configuration *EngineConfiguration
	ConfigRedo    *EngineConfiguration // 需要重新配置的信息 , 否则为 nil

	// time-lines

	RootTimeline    *Timeline // timeline for engine
	ProjectTimeline *Timeline // timeline for document (project)

	// runtime-data

	State EngineState

	TransportContext *TransportContext
	Looper           *Looper
	DOM              *DOM

	Wires []*Wire

	Channels []*Channel

	// 以下接口由 factory 创建

	DeviceFactory  DeviceFactory // the main Device factory
	DeviceManager  DeviceManager
	DeviceRegistry DeviceRegistry
	LifeManager    LifeManager
	Logger         Logger
	LS             EngineLS
	TrackManager   TrackManager

	// 以下接口由 device 提供, 是一些特殊的 device

	Clock     Clock
	Engine    Engine
	Transport Transport
}
