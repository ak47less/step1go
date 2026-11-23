package lib

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type BankIndex uint

type Bank struct {
	Index BankIndex

	Label string

	Description string

	StepBuffer StepArray

	Steps  StepSlice
	Tracks TrackSlice
}

////////////////////////////////////////////////////////////////////////////////

type BankArray []*Bank

func (array BankArray) Init(cap int) BankArray {

	count := array.Capacity()

	if count < 1 {
		array = make(BankArray, cap)
		for i := 0; i < count; i++ {
			label := fmt.Sprintf("bank_%d", i)
			item := &Bank{
				Index: BankIndex(i),
				Label: label,
			}
			array[i] = item
		}
	}

	return array
}

func (array BankArray) Slice(length int) BankSlice {
	dst := array
	cap := array.Capacity()
	if 0 <= length && length <= cap {
		dst = array[0:length]
	}
	return BankSlice(dst)
}

func (array BankArray) Capacity() int {
	return len(array)
}

////////////////////////////////////////////////////////////////////////////////

type BankSlice []*Bank

func (sl BankSlice) Length() int {
	return len(sl)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
