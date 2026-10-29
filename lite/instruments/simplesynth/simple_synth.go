package simplesynth

import "github.com/ak47less/step1go/lite"

type SimpleSynthFactory struct {

	// SimpleSynthFactory  :  这个工厂提供一个简单的合成器实现, 能生成: 正弦波 ,三角波,锯齿波,方波, 噪声波, 这些简单的波形

}

// ListInfo implements [lite.InstrumentFactory].
func (inst *SimpleSynthFactory) ListInfo() []*lite.InstrumentInfo {
	return []*lite.InstrumentInfo{
		{
			Name:   "default",
			Label:  "Simple Synth",
			Preset: "default",
		},
	}
}

// CreateInstrument implements [lite.InstrumentFactory].
func (inst *SimpleSynthFactory) CreateInstrument(
	ctx *lite.Context,
	info *lite.InstrumentInfo,
) (lite.Instrument, error) {
	panic("unimplemented")
}

func (inst *SimpleSynthFactory) _impl() lite.InstrumentFactory {
	return inst
}
