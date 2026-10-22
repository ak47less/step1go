package step1core

////////////////////////////////////////////////////////////////////////////////

type TransportState int

const (
	TransportStateInit TransportState = iota

	TransportStateStopped
	TransportStatePlaying
	TransportStateRecording
	TransportStatePaused
)

////////////////////////////////////////////////////////////////////////////////

type TransportContext struct {
	State TransportState

	Transport Transport

	Recording bool

	Playing bool

	Paused bool

	RootTimeline *Timeline
	WorkTimeline *Timeline
}

////////////////////////////////////////////////////////////////////////////////

type Transport interface {
	State() TransportState

	Play() error

	Stop() error

	Record() error

	Pause() error

	Resume() error
}

////////////////////////////////////////////////////////////////////////////////
// EOF
