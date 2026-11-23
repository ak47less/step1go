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

	// CurrentBankIndex  BankIndex
	// CurrentStepIndex  StepIndex
	// CurrentTrackIndex TrackIndex

	Cells   CellManager
	Buffers CubeBuffers

	Banks  BankSlice
	Steps  StepSlice
	Tracks TrackSlice
}

func (inst *Cube) Init(cfg *CubeConfig) error {

	cm, err := NewCellManager(inst, cfg)
	if err != nil {
		return err
	}

	inst.Cells = cm
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type CubeBuffers struct {
	BankBuffer  BankArray
	StepBuffer  StepArray
	TrackBuffer TrackArray
}

////////////////////////////////////////////////////////////////////////////////
// EOF
