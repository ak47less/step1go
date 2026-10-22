package lite

var _ Driver = new(innerMockDriver)

type innerMockDriver struct {
}

// GetAudioContextFactory implements [Driver].
func (inst *innerMockDriver) GetAudioContextFactory() AudioContextFactory {
	panic("unimplemented")
}
