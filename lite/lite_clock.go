package lite

////////////////////////////////////////////////////////////////////////////////
// Sequencer - clock

type ClockContext struct {
	Current TimePoint // at work-timeline

	CurrentStepIndex StepIndex
	CurrentStepHead  *StepHead
	CurrentStepBody  *StepBody
}

type ClockHandler interface {
	OnClock(cc *ClockContext) error
}
