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

	b := &inst.Buffers
	bank_cap := cfg.BankListCapacity
	step_cap := cfg.StepListCapacity
	track_cap := cfg.TrackListCapacity

	// init banks

	banks := b.BankBuffer.Init(bank_cap)

	// init steps

	steps := b.StepBuffer.Init(step_cap)

	// init tracks

	tracks := b.TrackBuffer.Init(track_cap)

	// init cells

	cm, err := NewCellManager(inst, cfg)
	if err != nil {
		return err
	}

	// bind

	err = inst.innerBindCells(cm, banks, steps, tracks)
	if err != nil {
		return err
	}

	// save

	b.BankBuffer = banks
	b.StepBuffer = steps
	b.TrackBuffer = tracks

	inst.Cells = cm
	inst.Banks = banks.Slice(cfg.BankListLength)
	inst.Steps = steps.Slice(cfg.StepListLength)
	inst.Tracks = tracks.Slice(cfg.TrackListLength)

	return nil
}

func (inst *Cube) innerBindCells(cells CellManager, banks BankArray, steps StepArray, tracks TrackArray) error {

	count_banks := banks.Capacity()
	count_steps := steps.Capacity()
	count_tracks := tracks.Capacity()
	var coo CellCoordinate

	for ibank := uint(0); ibank < count_banks; ibank++ {

		bank := banks[ibank]
		bank.Cube = inst

		for istep := uint(0); istep < count_steps; istep++ {

			step := steps[istep]
			step.Cube = inst

			for itrack := uint(0); itrack < count_tracks; itrack++ {

				track := tracks[itrack]
				track.Cube = inst

				coo.BankIndex = BankIndex(ibank)
				coo.StepIndex = StepIndex(istep)
				coo.TrackIndex = TrackIndex(itrack)

				cell, err := cells.GetCell(coo)
				if err != nil {
					return err
				}

				cell.OwnerBank = bank
				cell.OwnerStep = step
				cell.OwnerTrack = track
			}
		}
	}

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
