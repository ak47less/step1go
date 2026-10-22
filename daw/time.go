package step1core

type SampleTime int64 // 基于样本计数的时间戳

type UnixTime int64 // in ms, from 1970-01-01

////////////////////////////////////////////////////////////////////////////////
// time-line

type Timeline struct {
	Parent *Timeline

	Audio *AudioConfig

	MIDI *MidiConfig

	Offset SampleTime // offset in samples

	Length SampleTime // length in samples

	Loop bool
}

func (inst *Timeline) Clone() *Timeline {
	src := inst
	dst := new(Timeline)
	if src != nil {
		*dst = *src
	}
	return dst.Normalize()
}

func (inst *Timeline) Normalize() *Timeline {

	if inst == nil {
		inst = new(Timeline)
	}

	if inst.Length < 1 {
		inst.Length = 10
	}

	return inst
}

////////////////////////////////////////////////////////////////////////////////
// time-point

type TimePoint struct {
	Offset SampleTime
	At     *Timeline
}

////////////////////////////////////////////////////////////////////////////////
// EOF
