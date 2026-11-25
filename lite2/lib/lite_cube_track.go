package lib

import (
	"fmt"

	"github.com/ak47less/step1go/lite2/core"
)

////////////////////////////////////////////////////////////////////////////////

type TrackIndex uint

type TrackName string

////////////////////////////////////////////////////////////////////////////////

type Track struct {
	Cube *Cube

	Enabled bool

	// 使用  MixerChannel 代替:
	// Solo    bool
	// Mute    bool
	// Gain    GainDB

	Index TrackIndex

	Name TrackName

	Label string

	Description string

	Order int

	MidiChannel core.MidiChannelNum

	Instrument       Instrument
	InstrumentHolder InstrumentHolder
	InstrumentInfo   InstrumentInfo

	MixerChannel       MixerChannel
	MixerChannelHolder MixerChannelHolder
}

////////////////////////////////////////////////////////////////////////////////

type TrackArray []*Track

func (array TrackArray) Init(cap uint) TrackArray {

	if array.Capacity() < 1 {
		var i uint
		array = make(TrackArray, cap)
		for i = 0; i < cap; i++ {
			label := fmt.Sprintf("track_%d", i)
			item := &Track{
				Index: TrackIndex(i),
				Label: label,
			}
			array[i] = item
		}
	}

	return array
}

func (array TrackArray) Slice(length uint) TrackSlice {
	dst := array
	cap := array.Capacity()
	if length <= cap {
		dst = array[0:length]
	}
	return TrackSlice(dst)
}

func (array TrackArray) Capacity() uint {
	return uint(len(array))
}

////////////////////////////////////////////////////////////////////////////////

type TrackSlice []*Track

func (sl TrackSlice) Length() uint {
	return uint(len(sl))
}

////////////////////////////////////////////////////////////////////////////////
// EOF
