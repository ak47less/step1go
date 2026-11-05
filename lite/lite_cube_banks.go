package lite

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// bank (配件)

type BankIndex uint

type BankName string

////////////////////////////////////////////////////////////////////////////////
// BankHolder

type BankHolder interface {
	GetBank() (*Bank, error)
}

////////////////////////////////////////////////////////////////////////////////

type BankArray []*Bank

func (array BankArray) Capacity() uint {
	n := len(array)
	return uint(n)
}

func (array BankArray) Init(capacity uint) BankArray {

	if capacity < 1 {
		capacity = 1
	}

	oldcapa := array.Capacity()
	if capacity > oldcapa {
		array = make(BankArray, capacity)
	}

	for idx, item := range array {
		if item == nil {
			item = new(Bank)
			array[idx] = item
		}
		item.Index = BankIndex(idx)
	}

	return array
}

func (array BankArray) GetSlice(length uint) BankSlice {
	var sl BankSlice
	return sl.Init(array, length)
}

func (array BankArray) GetItem(index BankIndex) *Bank {
	i := uint(index)
	size := array.Capacity()
	if i < size {
		return array[i]
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type BankSlice []*Bank

func (slice BankSlice) Length() uint {
	n := len(slice)
	return uint(n)
}

func (slice BankSlice) Init(array BankArray, length uint) BankSlice {
	if length < 1 {
		length = 1
	}
	capa := array.Capacity()
	if length < capa {
		tmp := array[0:length]
		slice = BankSlice(tmp)
	} else {
		slice = BankSlice(array)
	}
	return slice
}

func (slice BankSlice) Get(index uint) *Bank {
	count := slice.Length()
	if index < count {
		return slice[index]
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//  BankList

// type BankList struct {
// 	list     []*Bank
// 	size     uint // 存储区数量(有效值)
// 	capacity uint // 存储区数量(最大值)
// }

// func (inst *BankList) SetCapacity(cap uint) uint {
// 	const max = 16
// 	if cap > max {
// 		cap = max
// 	}
// 	inst.capacity = cap
// 	return cap
// }

// func (inst *BankList) SetSize(size uint) uint {
// 	max := inst.capacity
// 	if size > max {
// 		size = max
// 	}
// 	inst.size = size
// 	inst.updateList()
// 	return size
// }

// func (inst *BankList) GetSize() uint {
// 	return inst.size
// }

// func (inst *BankList) GetBank(index BankIndex) *Bank {
// 	all := inst.list
// 	cnt := len(all)
// 	idx := int(index)
// 	if idx < cnt {
// 		return all[idx]
// 	}
// 	return nil
// }

// func (inst *BankList) Init(size, capacity uint) {
// 	if inst.capacity == 0 {
// 		inst.SetCapacity(capacity)
// 		inst.SetSize(size)
// 	}
// }

// func (inst *BankList) Items() []*Bank {
// 	all := inst.list
// 	want := int(inst.size)
// 	have := len(all)
// 	if want < have {
// 		return all[0:want]
// 	}
// 	return all
// }

// func (inst *BankList) updateList() {

// 	have := len(inst.list)
// 	want := int(inst.size)

// 	for ; have < want; have++ {

// 		item := new(Bank)
// 		nstr := strconv.Itoa(have)

// 		item.Index = BankIndex(have)
// 		item.Name = BankName("bank" + nstr)
// 		item.Label = "Bank-" + nstr

// 		inst.list = append(inst.list, item)
// 	}

// }

////////////////////////////////////////////////////////////////////////////////
// bank (存储区, 主体结构)

type Bank struct {
	cube *Cube

	// head

	Index       BankIndex
	Name        BankName
	Label       string
	Description string

	// body

	StepBuffer  StepArray
	TrackBuffer TrackArray

	Steps  StepSlice
	Tracks TrackSlice
}

// OnStep implements [StepTickHandler].
func (inst *Bank) OnStep(stc *StepTickContext) error {

	err := inst.innerAutoInitSelf(stc)
	if err != nil {
		return err
	}

	if inst == nil {
		return fmt.Errorf("bank-instance is nil")
	}

	// if inst.Context == nil {
	// 	return fmt.Errorf("use bank without init it")
	// }

	if stc == nil {
		return fmt.Errorf("param:STC(StepTickContext) is nil")
	}

	stp := &stc.Current
	idx := stp.Index
	step := inst.Steps.Get(idx)

	if step == nil {
		return nil
	}

	return step.OnStep(stc)
}

func (inst *Bank) Init(cube *Cube) error {

	if inst == nil || cube == nil {
		return fmt.Errorf("bank.init() : param is nil")
	}

	step_cap := cube.Buffers.StepBuffer.Capacity()
	step_len := cube.Steps.Length()
	track_cap := cube.Buffers.TrackBuffer.Capacity()
	track_len := cube.Tracks.Length()

	inst.cube = cube

	inst.StepBuffer.Init(step_cap)
	inst.TrackBuffer.Init(track_cap)

	inst.Steps = inst.StepBuffer.GetSlice(step_len)
	inst.Tracks = inst.TrackBuffer.GetSlice(track_len)

	return nil
}

func (inst *Bank) innerAutoInitSelf(stc *StepTickContext) error {
	if inst.cube == nil {
		sc := stc.context
		cube := &sc.Cube
		return inst.Init(cube)
	}
	return nil
}

// 根据当前状态重新加载这个 bank 中的各个列表
func (inst *Bank) Reload() error {

	return fmt.Errorf("todo: no impl")
}

func (inst *Bank) _impl() StepTickHandler {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
