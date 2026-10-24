package lite

import "time"

////////////////////////////////////////////////////////////////////////////////
// Timeline

type Timeline struct {
	Parent *Timeline

	Label string

	Offset Seconds

	Length Seconds

	BPM BPM

	Tempo Tempo

	Loop bool
}

func (inst *Timeline) HasSameRoot(other *Timeline) bool {
	r1 := inst.GetRoot()
	r2 := other.GetRoot()
	return r1.Same(r2)
}

func (inst *Timeline) Same(other *Timeline) bool {
	return inst == other
}

func (inst *Timeline) Equals(other *Timeline) bool {
	if inst == nil || other == nil {
		return false
	}
	if inst == other {
		return true
	}
	if inst.Label != other.Label ||
		inst.Offset != other.Offset ||
		inst.Length != other.Length ||
		inst.BPM != other.BPM ||
		inst.Tempo != other.Tempo ||
		inst.Loop != other.Loop {
		return false
	}
	if inst.Parent == nil || other.Parent == nil {
		return inst.Parent == other.Parent
	}
	return inst.Parent.Equals(other.Parent)
}

func (inst *Timeline) GetRoot() *Timeline {
	p := inst
	for ttl := 32; ttl > 0; ttl-- {
		if p != nil {
			parent := p.Parent
			if parent != nil {
				p = parent
				continue
			}
		}
		break
	}
	return p
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
// Position

/***************************************/
/*  class : AbsolutePosition           */
/***************************************/

// AbsolutePosition 表示全局统一的绝对位置
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

	dst.At = base
	dst.Position = Seconds(ap) - off // the rel = ( abs - off )

	return dst
}

/***************************************/
/*  class : RelativePosition           */
/***************************************/

// RelativePosition  表示相对于 Timeline 原点的位置
type RelativePosition struct {
	At       *Timeline
	Position Seconds
}

// 转换为绝对值
func (inst *RelativePosition) Absolute() AbsolutePosition {
	if inst == nil {
		return 0
	}
	at := inst.At
	off := at.GetTotalOffset()
	rel := inst.Position
	return AbsolutePosition(off + rel) // = abs
}

////////////////////////////////////////////////////////////////////////////////

type Timer struct {
	Delay    Seconds
	Interval Seconds
	Handler  func(t *Timer) error
	workerID int
}

func (inst *Timer) Start() {

	inst.workerID++

	id := inst.workerID
	h := inst.Handler

	go func() {
		inst.work(id, h)
	}()
}

func (inst *Timer) Stop() {
	inst.workerID = -1
}

func (inst *Timer) work(worker int, h func(t *Timer) error) {

	if h == nil {
		return
	}

	var err error
	t := inst.Delay

	for {

		du := inst.sec2du(t)
		time.Sleep(du)

		if worker != inst.workerID {
			break
		}

		err = h(inst)
		if err != nil {
			handleError(err)
		}

		t = inst.Interval
	}

}

func (inst *Timer) sec2du(sec Seconds) time.Duration {
	const min = time.Millisecond
	f := sec * Seconds(time.Second)
	return max(time.Duration(f), min)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
