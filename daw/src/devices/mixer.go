package devices

import "github.com/ak47less/step1go/step1core"

////////////////////////////////////////////////////////////////////////////////
// mixer - device

type innerMixer struct {
}

// Clock implements [step1core.Device].
func (i *innerMixer) Clock(bus *step1core.Bus) error {
	panic("unimplemented")
}

// Configure implements [step1core.Device].
func (i *innerMixer) Configure(ec *step1core.EngineContext, cfg *step1core.EngineConfiguration) error {
	panic("unimplemented")
}

// GetEngine implements [step1core.Device].
func (i *innerMixer) GetEngine() step1core.Engine {
	panic("unimplemented")
}

// GetInfo implements [step1core.Device].
func (i *innerMixer) GetInfo(ei *step1core.DeviceInfo) *step1core.DeviceInfo {
	panic("unimplemented")
}

// GetPorts implements [step1core.Device].
func (i *innerMixer) GetPorts() []step1core.Port {
	panic("unimplemented")
}

// GetProperties implements [step1core.Device].
func (i *innerMixer) GetProperties() []*step1core.Property {
	panic("unimplemented")
}

// Route implements [step1core.Device].
func (i *innerMixer) Route(bus *step1core.Bus) error {
	panic("unimplemented")
}

////////////////////////////////////////////////////////////////////////////////
// mixer - factory

type MixerFactory struct {
}

// CreateDevice implements [step1core.DeviceFactory].
func (inst *MixerFactory) CreateDevice(ei *step1core.DeviceInfo) (step1core.Device, error) {
	dev := new(innerMixer)
	return dev, nil
}

// GetFactory implements [step1core.DeviceProvider].
func (inst *MixerFactory) GetFactory() step1core.DeviceFactory {
	return inst
}

// GetRegistration implements [step1core.DeviceProvider].
func (inst *MixerFactory) GetRegistration() *step1core.DeviceRegistration {

	reg := new(step1core.DeviceRegistration)

	reg.Class = step1core.DeviceClassMixer
	reg.Factory = inst
	reg.Provider = inst

	return reg
}

func (inst *MixerFactory) _impl() (step1core.DeviceFactory, step1core.DeviceProvider) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
