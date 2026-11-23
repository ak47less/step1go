package lib

////////////////////////////////////////////////////////////////////////////////

type CubeConfig struct {
	BankListCapacity uint
	BankListLength   uint

	StepListCapacity uint
	StepListLength   uint

	TrackListCapacity uint
	TrackListLength   uint
}

////////////////////////////////////////////////////////////////////////////////

type Cube struct {
	CurrentBankIndex  BankIndex
	CurrentStepIndex  StepIndex
	CurrentTrackIndex TrackIndex

	Buffer CubeBuffer

	Banks  BankSlice
	Steps  StepSlice
	Tracks TrackSlice
}

func (inst *Cube) Init(cfg *CubeConfig) error {

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type CubeBuffer struct {
	BankBuffer  BankArray
	StepBuffer  StepArray
	TrackBuffer TrackArray
	CellBuffer  CellManager
}

////////////////////////////////////////////////////////////////////////////////
// EOF
