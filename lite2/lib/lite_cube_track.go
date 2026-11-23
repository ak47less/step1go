package lib

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type TrackIndex uint

type TrackName string

////////////////////////////////////////////////////////////////////////////////

type Track struct {
	Cube *Cube

	Enabled bool
	Solo    bool
	Mute    bool

	Index TrackIndex

	Name TrackName

	Label string

	Description string

	Order int
}

////////////////////////////////////////////////////////////////////////////////

type TrackArray []*Track

func (array TrackArray) Init(cap int) TrackArray {

	count := array.Capacity()

	if count < 1 {
		array = make(TrackArray, cap)
		for i := 0; i < count; i++ {
			label := fmt.Sprintf("bank_%d", i)
			item := &Track{
				Index: BankIndex(i),
				Label: label,
			}
			array[i] = item
		}
	}

	return array
}

func (array TrackArray) Slice(length int) TrackSlice {
	dst := array
	cap := array.Capacity()
	if 0 <= length && length <= cap {
		dst = array[0:length]
	}
	return TrackSlice(dst)
}

func (array TrackArray) Capacity() int {
	return len(array)
}

////////////////////////////////////////////////////////////////////////////////

type TrackSlice []*Track

func (sl TrackSlice) Length() int {
	return len(sl)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
