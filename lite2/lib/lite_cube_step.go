package lib

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type StepIndex int

type StepCount uint

type Step struct {
	Cube *Cube
}

////////////////////////////////////////////////////////////////////////////////

type StepArray []*Step

func (array StepArray) Init(cap int) StepArray {

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

func (array StepArray) Slice(length int) StepSlice {
	dst := array
	cap := array.Capacity()
	if 0 <= length && length <= cap {
		dst = array[0:length]
	}
	return StepSlice(dst)
}

func (array StepArray) Capacity() int {
	return len(array)
}

////////////////////////////////////////////////////////////////////////////////

type StepSlice []*Step

func (sl StepSlice) Length() int {
	return len(sl)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
