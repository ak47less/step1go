package step1go

////////////////////////////////////////////////////////////////////////////////
// life-cycle

type EngineLifeFunction func(e Engine) error

type TransportLifeFunction func(t Transport) error

type Life struct {
	Order int

	Label string

	// on_engine:

	OnEngineCreate  EngineLifeFunction
	OnEngineStart   EngineLifeFunction
	OnEngineStop    EngineLifeFunction
	OnEngineDestroy EngineLifeFunction

	OnEngineLoad   EngineLifeFunction
	OnEngineReload EngineLifeFunction
	OnEngineUpdate EngineLifeFunction

	// on_transport:

	OnTransportStart  TransportLifeFunction
	OnTransportPause  TransportLifeFunction
	OnTransportResume TransportLifeFunction
	OnTransportStop   TransportLifeFunction
}

type LifeAPI interface {
	GetLife() *Life
}

type LifeManager interface {
	Add(l LifeAPI)

	GetMain() LifeAPI

	ListAll() []*Life
}

////////////////////////////////////////////////////////////////////////////////
// EOF
