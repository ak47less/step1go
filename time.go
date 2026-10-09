package step1go

type SampleTime int64 // 基于样本计数的时间戳

type UnixTime int64 // in ms, from 1970-01-01

////////////////////////////////////////////////////////////////////////////////
// time-line

type Timeline struct {
	Parent *Timeline

	Audio *AudioConfig

	MIDI *MidiConfig

	Offset uint // in samples

	Length uint // in samples

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
// time-stamp

type TimeStamp struct {
	Position SampleTime

	At *Timeline
}

////////////////////////////////////////////////////////////////////////////////
// EOF
