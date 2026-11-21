package core

import "time"

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
