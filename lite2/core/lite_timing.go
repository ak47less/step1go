package core

////////////////////////////////////////////////////////////////////////////////

type TimePoint Seconds

func (tp TimePoint) Add(length TimeLength) TimePoint {
	return tp + TimePoint(length)
}

func (tp TimePoint) Sub(p2 TimePoint) TimeLength {
	value := tp - p2
	return TimeLength(value)
}

////////////////////////////////////////////////////////////////////////////////

type TimeLength Seconds

func (l TimeLength) Absolute() TimeLength {
	if l < 0 {
		return -l
	}
	return l
}

////////////////////////////////////////////////////////////////////////////////

type TimelineType int

const (
	TimelineTypeMin TimelineType = iota

	TimelineTypeRoot
	TimelineTypeDocument
	TimelineTypeEngine
	TimelineTypeLoop
	TimelineTypeFrame
	TimelineTypeClip
	TimelineTypeTransport
	TimelineTypeStep

	TimelineTypeMax
)

////////////////////////////////////////////////////////////////////////////////

type Timeline struct {
	Name string

	Type TimelineType

	Parent *Timeline

	Offset Seconds

	Length TimeLength

	Loop bool
}

func (inst *Timeline) GetTotalOffset() Seconds {
	p := inst
	var total Seconds = 0
	for ttl := 32; ttl > 0; ttl-- {
		if p == nil {
			break
		} else {
			total += p.Offset
			p = p.Parent
		}
	}
	return total
}

////////////////////////////////////////////////////////////////////////////////

type AbsolutePosition Seconds

// 转换为相对值
func (ap AbsolutePosition) Relative(base *Timeline, dst *RelativePosition) *RelativePosition {

	if dst == nil {
		dst = new(RelativePosition)
	}

	if base == nil {
		base = new(Timeline)
	}

	off := base.GetTotalOffset()

	dst.Ref = base
	dst.Pos = Seconds(ap) - off // the rel = ( abs - off )

	return dst
}

////////////////////////////////////////////////////////////////////////////////

type RelativePosition struct {
	Ref *Timeline
	Pos Seconds
}

// 转换为绝对值
func (inst *RelativePosition) Absolute() AbsolutePosition {
	if inst == nil {
		return 0
	}
	at := inst.Ref
	off := at.GetTotalOffset()
	rel := inst.Pos
	return AbsolutePosition(off + rel) // = abs
}

func (inst *RelativePosition) Clone() *RelativePosition {
	dst := new(RelativePosition)
	if inst != nil {
		*dst = *inst
	}
	return dst
}

////////////////////////////////////////////////////////////////////////////////
