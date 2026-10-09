package step1core

////////////////////////////////////////////////////////////////////////////////

type ElementID string

type ElementName string

type ElementClass string

////////////////////////////////////////////////////////////////////////////////

const (
	ElementClassClock       ElementClass = "step1/classes/clock"
	ElementClassEngine      ElementClass = "step1/classes/engine"
	ElementClassWaveMonitor ElementClass = "step1/classes/wave_monitor"
	ElementClassLevelMeter  ElementClass = "step1/classes/level_meter"
	ElementClassOscillator  ElementClass = "step1/classes/oscillator"
	ElementClassGain        ElementClass = "step1/classes/gain"

	ElementClassAudioIn  ElementClass = "step1/classes/audio_in"
	ElementClassAudioOut ElementClass = "step1/classes/audio_out"
	ElementClassFileIn   ElementClass = "step1/classes/file_in"
	ElementClassFileOut  ElementClass = "step1/classes/file_out"

	ElementClassExample ElementClass = "step1/classes/example"
)

////////////////////////////////////////////////////////////////////////////////

type ElementInfo struct {
	Class ElementClass

	ID ElementID

	Name ElementName

	Engine Engine

	Element Element
}

////////////////////////////////////////////////////////////////////////////////
// API

type Element interface {
	GetPorts() []Port

	GetProperties() []*Property

	GetInfo(ei *ElementInfo) *ElementInfo

	GetEngine() Engine

	// 更新 element 的时钟, 使之与 bus 的时钟同步
	Clock(lc *Bus) error

	IO(lc *Bus) error

	Configure(ec *EngineContext, cfg *EngineConfiguration) error
}

type ElementManager interface {
	Count() int

	GetElement(index int) Element

	FindElement(ei *ElementInfo) (Element, error)

	LoadElement(ei *ElementInfo) (Element, error)

	ListAll() []Element
}

type ElementFactory interface {
	CreateElement(ei *ElementInfo) (Element, error)
}

type ElementProvider interface {
	GetRegistration() *ElementRegistration

	GetFactory() ElementFactory
}

type ElementRegistration struct {
	Class ElementClass

	Provider ElementProvider

	Factory ElementFactory
}

type ElementRegistry interface {
	Register(provider ElementProvider)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
