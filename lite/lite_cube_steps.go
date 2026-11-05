package lite

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// step (配件)

// 物理下标
type StepIndex uint

// 逻辑下标
type StepLocation uint

////////////////////////////////////////////////////////////////////////////////

/**********************/
/* Step (position)    */
/**********************/

type StepPosition struct {

	// Position Seconds
	// At       *Timeline

	RelativePosition

	Index    StepIndex
	Location StepLocation
}

////////////////////////////////////////////////////////////////////////////////

/**********************/
/* Step (List)        */
/**********************/

// type StepList struct {
// 	list     []*Step // items
// 	size     uint    // 节拍(步奏)数量(有效值)
// 	capacity uint    // 节拍(步奏)数量(最大值)
// }

// func (inst *StepList) SetCapacity(cap uint) uint {
// 	const max = 16
// 	if cap > max {
// 		cap = max
// 	}
// 	inst.capacity = cap
// 	return cap
// }

// func (inst *StepList) GetSize() uint {
// 	return inst.size
// }

// func (inst *StepList) GetStep(index StepIndex) *Step {
// 	all := inst.list
// 	cnt := len(all)
// 	idx := int(index)
// 	if idx < cnt {
// 		return all[idx]
// 	}
// 	return nil
// }

// func (inst *StepList) SetSize(size uint) uint {
// 	max := inst.capacity
// 	if size > max {
// 		size = max
// 	}
// 	inst.size = size
// 	inst.updateList()
// 	return size
// }

// func (inst *StepList) Init(size, capacity uint) {
// 	if inst.capacity == 0 {
// 		inst.SetCapacity(capacity)
// 		inst.SetSize(size)
// 	}
// }

// func (inst *StepList) Items() []*Step {

// 	if inst == nil {
// 		return nil
// 	}

// 	all := inst.list
// 	want := int(inst.size)
// 	have := len(all)
// 	if want < have {
// 		return all[0:want]
// 	}
// 	return all
// }

// func (inst *StepList) updateList() {

// 	have := len(inst.list)
// 	want := int(inst.size)

// 	for ; have < want; have++ {

// 		item := new(Step)

// 		item.Index = StepIndex(have)
// 		item.Weight = 60

// 		inst.list = append(inst.list, item)
// 	}

// }

////////////////////////////////////////////////////////////////////////////////
// StepArray

type StepArray []*Step

func (array StepArray) Capacity() uint {
	n := len(array)
	return uint(n)
}

func (array StepArray) Init(capacity uint) StepArray {

	if capacity < 1 {
		capacity = 1
	}

	oldcapa := array.Capacity()
	if capacity > oldcapa {
		array = make(StepArray, capacity)
	}

	for idx, item := range array {
		if item == nil {
			item = new(Step)
			array[idx] = item
		}
		item.Index = StepIndex(idx)
	}

	return array
}

func (array StepArray) GetSlice(length uint) StepSlice {
	var sl StepSlice
	return sl.Init(array, length)
}

func (array StepArray) GetItem(index StepIndex) *Step {
	i := uint(index)
	size := array.Capacity()
	if i < size {
		return array[i]
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// StepSlice

type StepSlice []*Step

func (slice StepSlice) Length() uint {
	n := len(slice)
	return uint(n)
}

func (slice StepSlice) Init(array StepArray, length uint) StepSlice {
	if length < 1 {
		length = 1
	}
	capa := array.Capacity()
	if length < capa {
		tmp := array[0:length]
		slice = StepSlice(tmp)
	} else {
		slice = StepSlice(array)
	}
	return slice
}

func (slice StepSlice) Get(index StepIndex) *Step {
	count := slice.Length()
	idx := uint(index)
	if idx < count {
		return slice[idx]
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// step  (主体结构)

type Step struct {
	cube *Cube

	// Head

	Index StepIndex

	Weight byte

	// Body

	Cells CellCache
}

// OnStep implements [StepTickHandler].
func (inst *Step) OnStep(stc *StepTickContext) error {

	if inst == nil {
		return fmt.Errorf("Step-instance is nil")
	}

	var err error
	clist := inst.Cells.list

	for _, cell := range clist {
		err = cell.OnStep(stc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *Step) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
