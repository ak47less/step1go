package lib

import "fmt"

type CubeConfig struct {
	BankListCapacity uint
	BankListLength   uint

	StepListCapacity uint
	StepListLength   uint

	TrackListCapacity uint
	TrackListLength   uint
}

type Cube struct {
	CurrentBankIndex  BankIndex
	CurrentStepIndex  StepIndex
	CurrentTrackIndex TrackIndex

	AllBanks  BankArray
	AllSteps  StepArray
	AllTracks TrackArray
	AllCells  CellManager

	Banks  BankSlice
	Steps  StepSlice
	Tracks TrackSlice
}

func (inst *Cube) Init(cfg *CubeConfig) error {

	return fmt.Errorf("no impl")
}
