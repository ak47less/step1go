package lib

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type BankIndex uint

type Bank struct {
	Cube *Cube

	Index BankIndex

	Label string

	Description string

	// StepBuffer StepArray
	// Steps      StepSlice
	// Tracks     TrackSlice
}

////////////////////////////////////////////////////////////////////////////////

type BankArray []*Bank

func (array BankArray) Init(cap uint) BankArray {

	if array.Capacity() < 1 {
		var i uint
		array = make(BankArray, cap)
		for i = 0; i < cap; i++ {
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

func (array BankArray) Slice(length uint) BankSlice {
	dst := array
	cap := array.Capacity()
	if length <= cap {
		dst = array[0:length]
	}
	return BankSlice(dst)
}

func (array BankArray) Capacity() uint {
	n := len(array)
	return uint(n)
}

////////////////////////////////////////////////////////////////////////////////

type BankSlice []*Bank

func (sl BankSlice) Length() uint {
	n := len(sl)
	return uint(n)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
