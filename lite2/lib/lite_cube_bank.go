package lib

type BankIndex uint

type Bank struct {
	StepBuffer StepArray

	Steps  StepSlice
	Tracks TrackSlice
}

type BankList []*Bank

type BankArray BankList

type BankSlice BankList
