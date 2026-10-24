package lite

////////////////////////////////////////////////////////////////////////////////
// Sequencer - clock

type ClockContext struct {
	Current RelativePosition // at work-timeline

	CurrentStepCount StepCount
	CurrentStepIndex StepIndex
	CurrentStepHead  *StepHead
	CurrentStepBody  *StepBody
}

type ClockHandler interface {
	OnClock(cc *ClockContext) error
}
