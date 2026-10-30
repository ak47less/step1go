package lite

import "github.com/ak47less/step1go/common"

////////////////////////////////////////////////////////////////////////////////

type SimpleProfile struct {
	scale common.ScaleMode
	bpm   common.BPM
	tempo common.Tempo
	tone  common.ToneName
}

// GetScale implements [common.MusicProfile].
func (inst *SimpleProfile) GetScale() common.ScaleMode {
	return inst.scale
}

// GetSpeed implements [common.MusicProfile].
func (inst *SimpleProfile) GetSpeed() common.BPM {
	return inst.bpm
}

// GetTempo implements [common.MusicProfile].
func (inst *SimpleProfile) GetTempo() common.Tempo {
	return inst.tempo
}

// GetTone implements [common.MusicProfile].
func (inst *SimpleProfile) GetTone() common.ToneName {
	return inst.tone
}

func (inst *SimpleProfile) SetScale(scale common.ScaleMode) *SimpleProfile {
	inst.scale = scale
	return inst
}

func (inst *SimpleProfile) SetSpeed(bpm common.BPM) *SimpleProfile {
	inst.bpm = bpm
	return inst
}

func (inst *SimpleProfile) SetTempo(tempo common.Tempo) *SimpleProfile {
	inst.tempo = tempo
	return inst
}

func (inst *SimpleProfile) SetTone(tone common.ToneName) *SimpleProfile {
	inst.tone = tone
	return inst
}

func (inst *SimpleProfile) Init() *SimpleProfile {
	return inst.InitWith(nil)
}

func (inst *SimpleProfile) InitWith(base MusicProfile) *SimpleProfile {

	if inst == nil {
		inst = new(SimpleProfile)
	}

	if base != nil {
		inst.bpm = base.GetSpeed()
		inst.scale = base.GetScale()
		inst.tone = base.GetTone()
		inst.tempo = base.GetTempo()
	} else {
		inst.bpm = 100
		inst.scale = common.ScaleModeMajor
		inst.tone = common.ToneNameC
		inst.tempo = NewTempo(4, 4)
	}

	return inst
}

func (inst *SimpleProfile) _impl() MusicProfile {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
