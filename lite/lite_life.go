package lite

import (
	"fmt"
	"sort"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type LifeCycleStageFunc func() error

type LifeCycleInitFunc func(ec *EngineContext) error

type LifeCycle struct {
	Order int

	OnInit LifeCycleInitFunc

	OnCreate  LifeCycleStageFunc
	OnStart   LifeCycleStageFunc
	OnResume  LifeCycleStageFunc
	OnRun     LifeCycleStageFunc
	OnPause   LifeCycleStageFunc
	OnStop    LifeCycleStageFunc
	OnDestroy LifeCycleStageFunc
}

type Life interface {
	GetLifeCycle() *LifeCycle
}

type LifeManager interface {
	GetMain() Life

	Add(l Life)

	Remove(l Life)
}

////////////////////////////////////////////////////////////////////////////////

type innerLifeCycleContext struct {
	cache *innerLifeCycleCache

	main *innerMainLife

	ec *EngineContext

	all []Life

	facade LifeManager

	mu sync.Mutex

	onInitDone   bool
	onCreateDone bool
	onStartDone  bool
	paused       bool
}

////////////////////////////////////////////////////////////////////////////////

func NewLifeManager() LifeManager {

	ctx := new(innerLifeCycleContext)
	facade := new(innerLifeManager)
	main := new(innerMainLife)

	ctx.facade = facade
	ctx.main = main

	facade.context = ctx
	main.context = ctx

	return facade
}

////////////////////////////////////////////////////////////////////////////////

type innerLifeCycleCache struct {
	list []*LifeCycle // 经过排序的 LifeCycle 列表
}

func (inst *innerLifeCycleCache) innerForEach(reverse bool, fn func(lc *LifeCycle) error) error {

	if inst == nil || fn == nil {
		return fmt.Errorf("param is nil")
	}

	all := inst.list
	count := len(all)
	idx := 0
	var err error

	if reverse {
		for idx = count - 1; idx >= 0; idx-- {
			item := all[idx]
			if item == nil {
				continue
			}
			err = fn(item)
			if err != nil {
				return err
			}
		}
	} else {
		for idx = 0; idx < count; idx++ {
			item := all[idx]
			if item == nil {
				continue
			}
			err = fn(item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type innerMainLife struct {
	context *innerLifeCycleContext
}

// GetLifeCycle implements [Life].
func (inst *innerMainLife) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit:    inst.onInit,
		OnCreate:  inst.onCreate,
		OnStart:   inst.onStart,
		OnResume:  inst.onResume,
		OnRun:     inst.onRun,
		OnPause:   inst.onPause,
		OnStop:    inst.onStop,
		OnDestroy: inst.onDestroy,
	}
	return lc
}

func (inst *innerMainLife) innerGetCache() *innerLifeCycleCache {
	c := inst.context.cache
	if c == nil {
		c = new(innerLifeCycleCache)
		inst.innerLoadCache(c)
		inst.context.cache = c
	}
	return c
}

func (inst *innerMainLife) innerLoadCache(c *innerLifeCycleCache) error {

	c.list = nil
	src := inst.context.all
	dst := c.list

	for _, item1 := range src {
		if item1 == nil {
			continue
		}
		lc := item1.GetLifeCycle()
		if lc == nil {
			continue
		}
		dst = append(dst, lc)
	}

	sort.Slice(dst, func(i, j int) bool {
		o1 := dst[i]
		o2 := dst[j]
		return (o1.Order < o2.Order)
	})

	c.list = dst
	return nil
}

func (inst *innerMainLife) innerInvokeLifeCycleStageFunc(fn LifeCycleStageFunc) error {
	if fn == nil {
		return nil
	}
	return fn()
}

func (inst *innerMainLife) onInit(ec *EngineContext) error {

	if ec == nil {
		return fmt.Errorf("innerMainLife.onInit() : param 'ec:EngineContext' is nil")
	}

	inst.context.ec = ec
	inst.context.onInitDone = true
	cache := inst.innerGetCache()
	const reverse = false

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnInit
		if fn == nil {
			return nil
		}
		return fn(ec)
	})

	return err
}

func (inst *innerMainLife) onCreate() error {

	inst.context.onCreateDone = true
	cache := inst.innerGetCache()
	const reverse = false

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnCreate
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onStart() error {

	inst.context.onStartDone = true
	cache := inst.innerGetCache()
	const reverse = false

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnStart
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onPause() error {

	inst.context.paused = true
	cache := inst.innerGetCache()
	const reverse = true

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnPause
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onResume() error {

	inst.context.paused = false
	cache := inst.innerGetCache()
	const reverse = false

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnResume
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onRun() error {

	cache := inst.innerGetCache()
	const reverse = false

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnRun
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onStop() error {

	cache := inst.innerGetCache()
	const reverse = true

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnStop
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) onDestroy() error {

	cache := inst.innerGetCache()
	const reverse = true

	err := cache.innerForEach(reverse, func(lc *LifeCycle) error {
		fn := lc.OnDestroy
		return inst.innerInvokeLifeCycleStageFunc(fn)
	})
	return err
}

func (inst *innerMainLife) _impl() Life {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerItemLifeCaller struct {
	context *innerLifeCycleContext
}

func (inst *innerItemLifeCaller) callInitFn(ec *EngineContext, fn LifeCycleInitFunc) {
	if ec == nil || fn == nil {
		return
	}
	err := fn(ec)
	handleError(err)
}

func (inst *innerItemLifeCaller) callStageFn(fn LifeCycleStageFunc) {
	if fn == nil {
		return
	}
	err := fn()
	handleError(err)
}

func (inst *innerItemLifeCaller) call(l Life) {

	if l == nil || inst == nil {
		return
	}

	lc := l.GetLifeCycle()
	ctx := inst.context

	if lc == nil || ctx == nil {
		return
	}

	ec := ctx.ec

	if ctx.onInitDone && (ec != nil) {
		inst.callInitFn(ec, lc.OnInit)
	}

	if ctx.onCreateDone {
		inst.callStageFn(lc.OnCreate)
	}

	if ctx.onStartDone {
		inst.callStageFn(lc.OnStart)

		if !ctx.paused {
			inst.callStageFn(lc.OnResume)
			inst.callStageFn(lc.OnRun)
		}

	}

}

////////////////////////////////////////////////////////////////////////////////

type innerLifeManager struct {
	context *innerLifeCycleContext
}

// Add implements [LifeManager].
func (inst *innerLifeManager) Add(l Life) {

	if l == nil {
		return
	}

	mu := &inst.context.mu
	mu.Lock()
	defer mu.Unlock()

	all := inst.context.all
	all = append(all, l)

	inst.context.all = all
	inst.context.cache = nil

	var caller innerItemLifeCaller
	caller.context = inst.context
	caller.call(l)
}

// GetMain implements [LifeManager].
func (inst *innerLifeManager) GetMain() Life {
	return inst.context.main
}

// Remove implements [LifeManager].
func (inst *innerLifeManager) Remove(l Life) {

	if l == nil {
		return
	}

	mu := &inst.context.mu
	mu.Lock()
	defer mu.Unlock()

	all := inst.context.all
	for idx, item := range all {
		if item == l {
			all[idx] = nil
		}
	}
	inst.context.cache = nil
}

func (inst *innerLifeManager) _impl() LifeManager {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type LifeCycleRunner struct {
	ec *EngineContext
}

func (inst *LifeCycleRunner) Init(ec *EngineContext) *LifeCycleRunner {
	inst.ec = ec
	return inst
}

func (inst *LifeCycleRunner) innerRunLC(lc *LifeCycle) error {

	var err error
	stagelist := make([]func() error, 0)

	err = lc.OnInit(inst.ec)
	if err != nil {
		return err
	}

	stagelist = append(stagelist, lc.OnCreate)
	stagelist = append(stagelist, lc.OnStart)
	stagelist = append(stagelist, lc.OnResume)
	stagelist = append(stagelist, lc.OnRun)
	stagelist = append(stagelist, lc.OnPause)
	stagelist = append(stagelist, lc.OnStop)
	stagelist = append(stagelist, lc.OnDestroy)

	for _, stage := range stagelist {
		if stage == nil {
			continue
		}
		err = stage()
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *LifeCycleRunner) RunLC(lc *LifeCycle) error {
	return inst.innerRunLC(lc)
}

func (inst *LifeCycleRunner) RunL(l Life) error {
	lc := l.GetLifeCycle()
	return inst.innerRunLC(lc)
}

func (inst *LifeCycleRunner) RunLM(lm LifeManager) error {
	life := lm.GetMain()
	lc := life.GetLifeCycle()
	return inst.innerRunLC(lc)
}

////////////////////////////////////////////////////////////////////////////////

type LifeCycleStageLogger struct {
	logger Logger
}

// GetLifeCycle implements [Life].
func (inst *LifeCycleStageLogger) GetLifeCycle() *LifeCycle {
	lc := new(LifeCycle)

	lc.OnInit = inst.innerOnInit

	lc.OnCreate = inst.innerOnCreate
	lc.OnStart = inst.innerOnStart
	lc.OnResume = inst.innerOnResume
	lc.OnRun = inst.innerOnRun
	lc.OnPause = inst.innerOnPause
	lc.OnStop = inst.innerOnStop
	lc.OnDestroy = inst.innerOnDestroy

	return lc
}

// GetInfo implements [Part].
func (inst *LifeCycleStageLogger) GetInfo() *PartInfo {
	info := new(PartInfo)
	info.Name = "LifeCycleStageLogger"
	info.Part = inst
	return info
}

func (inst *LifeCycleStageLogger) innerOnInit(ec *Context) error {
	inst.logger = ec.Logger
	return inst.innerLogStage("OnInit")
}

func (inst *LifeCycleStageLogger) innerLogStage(stage string) error {
	inst.logger.Info("engine.lifecycle.%s()", stage)
	return nil
}

func (inst *LifeCycleStageLogger) innerOnCreate() error  { return inst.innerLogStage("OnCreate") }
func (inst *LifeCycleStageLogger) innerOnStart() error   { return inst.innerLogStage("OnStart") }
func (inst *LifeCycleStageLogger) innerOnResume() error  { return inst.innerLogStage("OnResume") }
func (inst *LifeCycleStageLogger) innerOnRun() error     { return inst.innerLogStage("OnRun") }
func (inst *LifeCycleStageLogger) innerOnPause() error   { return inst.innerLogStage("OnPause") }
func (inst *LifeCycleStageLogger) innerOnStop() error    { return inst.innerLogStage("OnStop") }
func (inst *LifeCycleStageLogger) innerOnDestroy() error { return inst.innerLogStage("OnDestroy") }

func (inst *LifeCycleStageLogger) _impl() (Life, Part) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
