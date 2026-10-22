package step1core

////////////////////////////////////////////////////////////////////////////////

type DeviceID string

type DeviceName string

type DeviceClass string

////////////////////////////////////////////////////////////////////////////////

const (
	DeviceClassClock       DeviceClass = "step1/classes/clock"
	DeviceClassEngine      DeviceClass = "step1/classes/engine"
	DeviceClassWaveMonitor DeviceClass = "step1/classes/wave_monitor"
	DeviceClassLevelMeter  DeviceClass = "step1/classes/level_meter"
	DeviceClassOscillator  DeviceClass = "step1/classes/oscillator"
	DeviceClassGain        DeviceClass = "step1/classes/gain"

	DeviceClassMixer   DeviceClass = "step1/classes/mixer"
	DeviceClassMixer8  DeviceClass = "step1/classes/mixer8"
	DeviceClassMixer16 DeviceClass = "step1/classes/mixer16"
	DeviceClassMixer32 DeviceClass = "step1/classes/mixer32"

	DeviceClassAudioIn  DeviceClass = "step1/classes/audio_in"
	DeviceClassAudioOut DeviceClass = "step1/classes/audio_out"
	DeviceClassFileIn   DeviceClass = "step1/classes/file_in"
	DeviceClassFileOut  DeviceClass = "step1/classes/file_out"

	DeviceClassExample DeviceClass = "step1/classes/example"
)

////////////////////////////////////////////////////////////////////////////////

type DeviceInfo struct {
	Class DeviceClass

	ID DeviceID

	Name DeviceName

	Engine Engine

	Device Device
}

////////////////////////////////////////////////////////////////////////////////
// API

type Device interface {
	GetPorts() []Port

	GetProperties() []*Property

	GetInfo(ei *DeviceInfo) *DeviceInfo

	GetEngine() Engine

	// 更新 Device 的时钟, 使之与 bus 的时钟同步
	Clock(bus *Bus) error

	// 根据设备连接图, 计算全部的信号路由路径,并根据计算结果生成经过排序的 Wire 列表
	Route(bus *Bus) error

	// todo: change to (bus *Bus)
	Configure(ec *EngineContext, cfg *EngineConfiguration) error
}

type DeviceManager interface {
	Count() int

	GetDevice(index int) Device

	FindDevice(ei *DeviceInfo) (Device, error)

	LoadDevice(ei *DeviceInfo) (Device, error)

	ListAll() []Device
}

type DeviceFactory interface {
	CreateDevice(ei *DeviceInfo) (Device, error)
}

type DeviceProvider interface {
	GetRegistration() *DeviceRegistration

	GetFactory() DeviceFactory
}

type DeviceRegistration struct {
	Class DeviceClass

	Provider DeviceProvider

	Factory DeviceFactory
}

type DeviceRegistry interface {
	Register(provider DeviceProvider)
}

type DeviceProviderManager interface {

	// DeviceManager

	DeviceRegistry

	DeviceFactory
}

////////////////////////////////////////////////////////////////////////////////
// EOF
