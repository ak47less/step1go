package lib

////////////////////////////////////////////////////////////////////////////////

type StepIndex int

type StepCount uint

type Step struct {
	Cube *Cube

	Index StepIndex
}

////////////////////////////////////////////////////////////////////////////////

type StepArray []*Step

func (array StepArray) Init(cap uint) StepArray {

	if array.Capacity() < 1 {
		var i uint
		array = make(StepArray, cap)
		for i = 0; i < cap; i++ {

			item := &Step{
				Index: StepIndex(i),
			}
			array[i] = item
		}
	}

	return array
}

func (array StepArray) Slice(length uint) StepSlice {
	dst := array
	cap := array.Capacity()
	if length <= cap {
		dst = array[0:length]
	}
	return StepSlice(dst)
}

func (array StepArray) Capacity() uint {
	return uint(len(array))
}

////////////////////////////////////////////////////////////////////////////////

type StepSlice []*Step

func (sl StepSlice) Length() uint {
	return uint(len(sl))
}

////////////////////////////////////////////////////////////////////////////////
// EOF
