package lib

import (
	"fmt"
	"sort"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type LifeStageCallbackFunc func() error

type LifeInitCallbackFunc func(ctx *ApplicationContext) error

////////////////////////////////////////////////////////////////////////////////

type LifeCycle struct {
	OnInit LifeInitCallbackFunc

	OnLoop LifeStageCallbackFunc

	OnCreate  LifeStageCallbackFunc
	OnStart   LifeStageCallbackFunc
	OnResume  LifeStageCallbackFunc
	OnPause   LifeStageCallbackFunc
	OnStop    LifeStageCallbackFunc
	OnDestroy LifeStageCallbackFunc

	Order int
}

////////////////////////////////////////////////////////////////////////////////

type Life interface {
	GetLifeCycle() *LifeCycle
}

type LifeManager interface {
	AddCycle(lc *LifeCycle) error

	AddLife(l Life) error

	GetMain() Life
}

////////////////////////////////////////////////////////////////////////////////

func NewLifeManager() LifeManager {
	return new(innerLifeManager)
}

////////////////////////////////////////////////////////////////////////////////

type innerLifeManager struct {
	mu sync.Mutex

	cycles []*LifeCycle
	cache  []*LifeCycle
	main   *innerMainLife

	context      *ApplicationContext
	onInitDone   bool
	onCreateDone bool
	onStartDone  bool
	paused       bool
}

// AddCycle implements [LifeManager].
func (inst *innerLifeManager) AddCycle(lc *LifeCycle) error {
	if inst == nil {
		return fmt.Errorf("life manager is nil")
	}
	if lc == nil {
		return fmt.Errorf("life cycle is nil")
	}

	inst.mu.Lock()
	inst.cycles = append(inst.cycles, lc)
	inst.cache = nil
	state := inst.innerReplayStateLocked()
	inst.mu.Unlock()

	return inst.innerReplay(lc, state)
}

// AddLife implements [LifeManager].
func (inst *innerLifeManager) AddLife(l Life) error {
	if inst == nil {
		return fmt.Errorf("life manager is nil")
	}
	if innerIsNil(l) {
		return fmt.Errorf("life is nil")
	}

	lc := l.GetLifeCycle()
	if lc == nil {
		return fmt.Errorf("life cycle is nil")
	}
	return inst.AddCycle(lc)
}

// GetMain implements [LifeManager].
func (inst *innerLifeManager) GetMain() Life {
	if inst == nil {
		return nil
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.main == nil {
		inst.main = &innerMainLife{manager: inst}
	}
	return inst.main
}

type innerMainLife struct {
	manager *innerLifeManager
}

// GetLifeCycle implements [Life].
func (inst *innerMainLife) GetLifeCycle() *LifeCycle {
	if inst == nil || inst.manager == nil {
		return nil
	}

	return &LifeCycle{
		OnInit:    inst.manager.innerOnInit,
		OnCreate:  func() error { return inst.manager.innerRunCreate() },
		OnStart:   func() error { return inst.manager.innerRunStart() },
		OnResume:  func() error { return inst.manager.innerRunResume() },
		OnLoop:    func() error { return inst.manager.innerRunLoop() },
		OnPause:   func() error { return inst.manager.innerRunPause() },
		OnStop:    func() error { return inst.manager.innerRunStop() },
		OnDestroy: func() error { return inst.manager.innerRunDestroy() },
	}
}

type innerReplayState struct {
	context      *ApplicationContext
	onInitDone   bool
	onCreateDone bool
	onStartDone  bool
	paused       bool
}

func (inst *innerLifeManager) innerReplayStateLocked() innerReplayState {
	return innerReplayState{
		context:      inst.context,
		onInitDone:   inst.onInitDone,
		onCreateDone: inst.onCreateDone,
		onStartDone:  inst.onStartDone,
		paused:       inst.paused,
	}
}

func (inst *innerLifeManager) innerReplay(lc *LifeCycle, state innerReplayState) error {
	if state.onInitDone && state.context != nil && lc.OnInit != nil {
		if err := lc.OnInit(state.context); err != nil {
			return err
		}
	}
	if state.onCreateDone && lc.OnCreate != nil {
		if err := lc.OnCreate(); err != nil {
			return err
		}
	}
	if state.onStartDone {
		if lc.OnStart != nil {
			if err := lc.OnStart(); err != nil {
				return err
			}
		}
		if !state.paused && lc.OnResume != nil {
			return lc.OnResume()
		}
	}
	return nil
}

func (inst *innerLifeManager) innerOnInit(ctx *ApplicationContext) error {
	if ctx == nil {
		return fmt.Errorf("application context is nil")
	}

	inst.mu.Lock()
	inst.context = ctx
	inst.onInitDone = true
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()

	for _, lc := range cycles {
		if lc.OnInit == nil {
			continue
		}
		if err := lc.OnInit(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (inst *innerLifeManager) innerRunCreate() error {
	inst.mu.Lock()
	inst.onCreateDone = true
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, false, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnCreate
	})
}

func (inst *innerLifeManager) innerRunStart() error {
	inst.mu.Lock()
	inst.onStartDone = true
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, false, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnStart
	})
}

func (inst *innerLifeManager) innerRunResume() error {
	inst.mu.Lock()
	inst.paused = false
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, false, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnResume
	})
}

func (inst *innerLifeManager) innerRunPause() error {
	inst.mu.Lock()
	inst.paused = true
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, true, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnPause
	})
}

func (inst *innerLifeManager) innerRunLoop() error {
	inst.mu.Lock()
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, false, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnLoop
	})
}

func (inst *innerLifeManager) innerRunStop() error {
	inst.mu.Lock()
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, true, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnStop
	})
}

func (inst *innerLifeManager) innerRunDestroy() error {
	inst.mu.Lock()
	cycles := inst.innerSortedCyclesLocked()
	inst.mu.Unlock()
	return inst.innerRunStage(cycles, true, func(lc *LifeCycle) LifeStageCallbackFunc {
		return lc.OnDestroy
	})
}

func (inst *innerLifeManager) innerRunStage(
	cycles []*LifeCycle,
	reverse bool,
	callback func(*LifeCycle) LifeStageCallbackFunc,
) error {
	for idx := range cycles {
		index := idx
		if reverse {
			index = len(cycles) - 1 - idx
		}
		lc := cycles[index]
		fn := callback(lc)
		if fn == nil {
			continue
		}
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (inst *innerLifeManager) innerSortedCyclesLocked() []*LifeCycle {
	if inst.cache == nil {
		cycles := make([]*LifeCycle, 0, len(inst.cycles))
		for _, lc := range inst.cycles {
			if lc != nil {
				cycles = append(cycles, lc)
			}
		}
		sort.SliceStable(cycles, func(i, j int) bool {
			return cycles[i].Order < cycles[j].Order
		})
		inst.cache = cycles
	}

	result := make([]*LifeCycle, len(inst.cache))
	copy(result, inst.cache)
	return result
}

////////////////////////////////////////////////////////////////////////////////
// EOF
