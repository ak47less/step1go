package lite

import "sort"

////////////////////////////////////////////////////////////////////////////////
// track

type TrackIndex int

type TrackName string

////////////////////////////////////////////////////////////////////////////////

type TrackMeta struct {
	Index       TrackIndex
	Name        TrackName
	Label       string
	Description string
	Weight      byte
	Channel     byte // midi-channel

	Mute    bool
	Solo    bool
	Locked  bool
	Enabled bool

	Gain float32 //  (unit=dB, 0=100%)

	Order int // 这个轨道在视图中的顺序

	Instrument Instrument
}

////////////////////////////////////////////////////////////////////////////////

type Track struct {

	// Head

	Meta  *TrackMeta
	cube  *Cube
	Index TrackIndex

	// Body

	Cells CellCache
}

////////////////////////////////////////////////////////////////////////////////
// TrackArray

type TrackArray []*Track

func (array TrackArray) Capacity() uint {
	n := len(array)
	return uint(n)
}

func (array TrackArray) Init(capacity uint) TrackArray {

	if capacity < 1 {
		capacity = 1
	}

	oldcapa := array.Capacity()
	if capacity > oldcapa {
		array = make(TrackArray, capacity)
	}

	for idx, item := range array {
		if item == nil {
			item = new(Track)
			array[idx] = item
		}
		item.Index = TrackIndex(idx)
	}

	return array
}

func (array TrackArray) GetSlice(length uint) TrackSlice {
	var sl TrackSlice
	return sl.Init(array, length)
}

func (array TrackArray) GetItem(index TrackIndex) *Track {
	i := uint(index)
	size := array.Capacity()
	if i < size {
		return array[i]
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TrackSlice

type TrackSlice []*Track

func (slice TrackSlice) Length() uint {
	n := len(slice)
	return uint(n)
}

func (slice TrackSlice) Init(array TrackArray, length uint) TrackSlice {
	if length < 1 {
		length = 1
	}
	capa := array.Capacity()
	if length < capa {
		tmp := array[0:length]
		slice = TrackSlice(tmp)
	} else {
		slice = TrackSlice(array)
	}
	return slice
}

func (slice TrackSlice) Get(index uint) *Track {
	count := slice.Length()
	if index < count {
		return slice[index]
	}
	return nil
}

func (slice TrackSlice) Sort(fn func(a, b *Track) bool) {

	if slice == nil || fn == nil {
		return
	}

	sort.Slice(slice, func(i, j int) bool {
		o1 := slice[i]
		o2 := slice[j]
		if o1 == nil || o2 == nil {
			return false
		}
		return fn(o1, o2)
	})

}

func (slice TrackSlice) Filter(fn func(t *Track) bool) TrackSlice {

	src := slice
	if src == nil || fn == nil {
		return src
	}

	size := slice.Length()
	dst := make([]*Track, 0, size)

	for _, item := range src {
		if item == nil {
			continue
		}
		if fn(item) {
			dst = append(dst, item)
		}
	}

	return dst
}

////////////////////////////////////////////////////////////////////////////////

/**********************/
/* TrackList          */
/**********************/

// type TrackList struct {
// 	list     []*Track //items
// 	size     uint     // 轨道数量(有效值)
// 	capacity uint     // 轨道数量(最大值)
// }

// func (inst *TrackList) SetCapacity(cap uint) uint {
// 	const max = 16
// 	if cap > max {
// 		cap = max
// 	}
// 	inst.capacity = cap
// 	return cap
// }

// func (inst *TrackList) SetSize(size uint) uint {
// 	max := inst.capacity
// 	if size > max {
// 		size = max
// 	}
// 	inst.size = size
// 	inst.updateList()
// 	return size
// }

// func (inst *TrackList) GetSize() uint {
// 	return inst.size
// }

// func (inst *TrackList) GetTrack(index TrackIndex) *Track {
// 	all := inst.list
// 	cnt := len(all)
// 	idx := int(index)
// 	if idx < cnt {
// 		return all[idx]
// 	}
// 	return nil
// }

// func (inst *TrackList) Init(size, capacity uint) {
// 	if inst.capacity == 0 {
// 		inst.SetCapacity(capacity)
// 		inst.SetSize(size)
// 	}
// }

// func (inst *TrackList) Items() []*Track {
// 	all := inst.list
// 	want := int(inst.size)
// 	have := len(all)
// 	if want < have {
// 		return all[0:want]
// 	}
// 	return all
// }

// func (inst *TrackList) updateList() {

// 	have := len(inst.list)
// 	want := int(inst.size)

// 	for ; have < want; have++ {

// 		item := new(Track)
// 		nstr := strconv.Itoa(have)

// 		item.Index = TrackIndex(have)
// 		item.Name = TrackName("track" + nstr)
// 		item.Label = "Track-" + nstr

// 		inst.list = append(inst.list, item)
// 	}

// }

////////////////////////////////////////////////////////////////////////////////
// EOF
