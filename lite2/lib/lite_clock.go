package lib

import (
	"fmt"
	"sync"

	"github.com/ak47less/step1go/lite2/core"
)

type Clock interface {
	AddListener(l ClockListener) []*ClockObserver

	AddObserver(ob *ClockObserver) *ClockObserver

	RemoveObserver(id ClockObserverID)
}

func NewClock() Clock {
	return new(innerMainClock)
}

////////////////////////////////////////////////////////////////////////////////

type ClockUnit byte

const (
	ClockUnitNone ClockUnit = iota

	ClockUnitStep
	ClockUnitFrame
)

////////////////////////////////////////////////////////////////////////////////

type ClockEvent struct {
	Clock Clock

	Index uint

	Unit ClockUnit // ( step|frame )

	UnitLength TimeLength // ( step_length | frame_length )

	Timeline *Timeline

	Realtime AbsolutePosition

	Thistime AbsolutePosition

	Preload bool //  ( 1=preload, 0=now )
}

type ClockEventHandlerFunc func(c *ClockEvent) error

////////////////////////////////////////////////////////////////////////////////

type ClockObserverID int

type ClockObserver struct {
	ID ClockObserverID

	Clock Clock

	Priority int

	Preload bool

	Unit ClockUnit

	Handler ClockEventHandlerFunc
}

////////////////////////////////////////////////////////////////////////////////

type ClockListener interface {
	Listen(c Clock) []*ClockObserver
}

////////////////////////////////////////////////////////////////////////////////

type innerMainClock struct {
	cc *innerClockContext
}

// AddObserver implements [Clock].
func (inst *innerMainClock) AddObserver(ob *ClockObserver) *ClockObserver {

	if inst == nil || ob == nil {
		return nil
	}

	cc := inst.cc
	mu := &cc.mutex
	mu.Lock()
	defer mu.Unlock()

	if ob.ID == 0 {
		ob.ID = cc.nextObserverID()
	}

	cc.observers[ob.ID] = ob
	cc.cache = nil

	return ob
}

// RemoveObserver implements [Clock].
func (inst *innerMainClock) RemoveObserver(id ClockObserverID) {

	cc := inst.cc
	mu := &cc.mutex
	mu.Lock()
	defer mu.Unlock()

	older := cc.observers[id]
	if older != nil {
		cc.observers[id] = nil
		cc.cache = nil
	}

}

func (inst *innerMainClock) init(ac *ApplicationContext) {

	if inst == nil || ac == nil {
		return
	}

	cc := inst.cc
	if cc != nil {
		return
	}

	cc = new(innerClockContext)
	cc.observers = make(map[ClockObserverID]*ClockObserver)
	cc.channels = make(map[ClockChannelID]*ClockChannel)
	cc.observerIDGen = 0x123456
	cc.preloadLength = 1
	cc.bufferLength = 3
	cc.context = ac
	cc.clock = inst
	cc.timelineForEngine = ac.EngineContext.EngineTimeline
	cc.timelineForDocument = ac.EngineContext.DocumentTimeline

	inst.innerAddChannel(cc, ClockUnitStep, true, 4)
	inst.innerAddChannel(cc, ClockUnitStep, false, 3)
	inst.innerAddChannel(cc, ClockUnitFrame, true, 2)
	inst.innerAddChannel(cc, ClockUnitFrame, false, 1)

	inst.cc = cc
}

func (inst *innerMainClock) innerAddChannel(cc *innerClockContext, unit ClockUnit, preload bool, priority int) {

	chl := &ClockChannel{
		unit:     unit,
		preload:  preload,
		priority: priority,
		interval: 0.5,
	}

	ec := cc.context.EngineContext

	switch unit {
	case ClockUnitFrame:
		if preload {
			chl.ccid = CCID_frame_preload
		} else {
			chl.ccid = CCID_frame_now
		}
		chl.timeline = ec.FrameTimeline
	case ClockUnitStep:
		if preload {
			chl.ccid = CCID_step_preload
		} else {
			chl.ccid = CCID_step_now
		}
		chl.timeline = ec.StepTimeline
	}

	cc.channels[chl.ccid] = chl
}

// AddListener implements [Clock].
func (inst *innerMainClock) AddListener(l ClockListener) []*ClockObserver {

	if inst == nil || l == nil {
		return nil
	}

	src := l.Listen(inst)
	dst := make([]*ClockObserver, 0)

	for _, ob := range src {
		ob2 := inst.AddObserver(ob)
		if ob2 == nil {
			continue
		}
		dst = append(dst, ob2)
	}

	return dst
}

func (inst *innerMainClock) innerFetchNow() AbsolutePosition {

	cc := inst.cc
	ac := cc.context
	stop_at := ac.TransportContext.StopAt

	if stop_at != nil {
		return stop_at.Absolute()
	}

	ctime := ac.EngineContext.AudioContext.GetCurrentTime()
	var current RelativePosition
	current.Pos = ctime
	current.Ref = ac.EngineContext.EngineTimeline

	return current.Absolute()
}

func (inst *innerMainClock) innerOnInit(ac *ApplicationContext) error {

	inst.init(ac)
	ac.Clock = inst

	return nil
}

func (inst *innerMainClock) innerOnTimer(t *Timer) error {

	cc := inst.cc
	cc.countOnTimer++
	now := inst.innerFetchNow()

	// appctx := cc.context
	// cnt := cc.countOnTimer

	// logger := appctx.Logger
	// logger.Info(" innerMainClock: onTimer (%d) ", cnt)

	return cc.dispatcher.dispatch(cc, now)
}

func (inst *innerMainClock) innerOnCreate() error {

	timer := &core.Timer{
		Delay:    0.1,
		Interval: (1.0 / 20),
		Handler:  inst.innerOnTimer,
	}

	inst.cc.timer = timer

	return nil
}

func (inst *innerMainClock) innerOnStart() error {
	inst.cc.timer.Start()
	return nil
}

func (inst *innerMainClock) innerOnStop() error {
	inst.cc.timer.Stop()
	return nil
}

// GetLifeCycle implements [Life].
func (inst *innerMainClock) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		Order:    0,
		OnInit:   inst.innerOnInit,
		OnCreate: inst.innerOnCreate,
		OnStart:  inst.innerOnStart,
		OnStop:   inst.innerOnStop,
	}
	return lc
}

func (inst *innerMainClock) _impl() (Clock, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////

type innerClockEventDispatcher struct {
	procedures []func(cc *innerClockContext) error
}

func (inst *innerClockEventDispatcher) dispatch(cc *innerClockContext, now AbsolutePosition) error {

	var err error
	plist := inst.innerGetProcedures()
	pre := now + AbsolutePosition(cc.preloadLength)

	cc.now = now
	cc.preload = pre
	cc.not_before = now - AbsolutePosition(cc.bufferLength)
	cc.not_after = pre + 0.1

	for _, pfn := range plist {
		err = pfn(cc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerClockEventDispatcher) innerDoPrepare(cc *innerClockContext) error {

	cc.getCache()

	return nil
}

func (inst *innerClockEventDispatcher) innerDoCheck(cc *innerClockContext) error {

	cc.getCache()

	return nil
}

func (inst *innerClockEventDispatcher) innerDoX(c *innerClockContext) error {

	return nil
}

// func (inst *innerClockEventDispatcher) innerInitEventForChannel(cc *innerClockContext, chl *ClockChannel, event *ClockEvent) error {

// 	tl := chl.timeline
// 	unit_len := tl.Length
// 	const min_unit_len = 0.001

// 	if unit_len < min_unit_len {
// 		unit_len = min_unit_len
// 	}

// 	event.Unit = chl.unit
// 	event.Preload = chl.preload
// 	event.UnitLength = unit_len
// 	event.Realtime = cc.now
// 	event.Clock = cc.clock

// 	// set later (on-fast-forward)
// 	event.Index = 0
// 	event.Thistime = 0

// 	return nil
// }

func (inst *innerClockEventDispatcher) innerDoDispatchToAllChannels(cc *innerClockContext) error {

	cache, err := cc.getCache()
	if err != nil {
		return err
	}

	channels := cache.channels
	var pos2abs AbsolutePosition

	for _, chl_holder := range channels {

		if chl_holder == nil {
			continue
		}
		chl := chl_holder.channel
		if chl == nil {
			continue
		}

		if chl.preload {
			pos2abs = cc.preload
		} else {
			pos2abs = cc.now
		}

		chl.innerTryUpdateInterval()

		err = chl.runFastForward(cc, pos2abs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerClockEventDispatcher) innerGetProcedures() []func(c *innerClockContext) error {
	list := inst.procedures
	if len(list) == 0 {

		list = append(list, inst.innerDoCheck)
		list = append(list, inst.innerDoPrepare)
		list = append(list, inst.innerDoX)
		list = append(list, inst.innerDoDispatchToAllChannels)

		inst.procedures = list
	}
	return list
}

////////////////////////////////////////////////////////////////////////////////

type ClockObserverHolder struct {
	ccid     ClockChannelID
	observer *ClockObserver
	channel  *ClockChannel
}

type ClockChannelHolder struct {
	ccid    ClockChannelID
	channel *ClockChannel
	obs     []*ClockObserverHolder
}

type innerClockObserverCache struct {
	channels []*ClockChannelHolder
}

func (inst *innerClockObserverCache) findByUnitPre(unit ClockUnit, preload bool) *ClockChannelHolder {
	all := inst.channels
	for _, item := range all {
		if item == nil {
			continue
		}
		chl := item.channel
		if chl == nil {
			continue
		}
		if chl.unit == unit && chl.preload == preload {
			return item
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type innerClockObserverCacheLoader struct {
}

func (inst *innerClockObserverCacheLoader) load(cc *innerClockContext) (*innerClockObserverCache, error) {

	if cc == nil {
		return nil, fmt.Errorf("innerClockContext is nil")
	}

	logger := cc.context.Logger
	logger.Info("load (reload) clock-obs-cache")

	var err error
	dst := new(innerClockObserverCache)

	err = inst.innerLoadChannels(cc, dst)
	if err != nil {
		return nil, err
	}

	err = inst.innerLoadObservers(cc, dst)
	if err != nil {
		return nil, err
	}

	err = inst.innerUpdateChannels(dst)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (inst *innerClockObserverCacheLoader) innerLoadChannels(cc *innerClockContext, cache *innerClockObserverCache) error {

	src1 := cc.channels

	for _, chl := range src1 {

		if chl == nil {
			continue
		}

		holder := &ClockChannelHolder{
			ccid:    chl.ccid,
			channel: chl,
			obs:     nil,
		}

		cache.channels = append(cache.channels, holder)
	}

	return nil
}

func (inst *innerClockObserverCacheLoader) innerLoadObservers(cc *innerClockContext, cache *innerClockObserverCache) error {

	src2 := cc.observers

	for _, ob := range src2 {

		if ob == nil {
			continue
		}

		h1 := cache.findByUnitPre(ob.Unit, ob.Preload)
		if h1 == nil {
			continue
		}

		h2 := &ClockObserverHolder{
			ccid:     h1.ccid,
			observer: ob,
			channel:  h1.channel,
		}
		h1.obs = append(h1.obs, h2)
	}

	return nil
}

func (inst *innerClockObserverCacheLoader) innerUpdateChannels(cache *innerClockObserverCache) error {
	list := cache.channels
	for _, chl_holder := range list {
		if chl_holder == nil {
			continue
		}
		if chl_holder.channel == nil {
			continue
		}
		chl_holder.channel.observers = chl_holder.obs
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type innerClockContext struct {
	context *ApplicationContext

	clock         Clock
	timer         *core.Timer
	mutex         sync.Mutex
	observerIDGen ClockObserverID

	preloadLength TimeLength // 预加载时长(提前量)
	bufferLength  TimeLength // 缓冲区时长, 超过这个值会触发 xrun

	timelineForDocument *Timeline
	timelineForEngine   *Timeline

	countOnTimer uint
	countXrun    uint

	now        AbsolutePosition // 当前的实时位置
	preload    AbsolutePosition // 当前的预加载位置
	not_before AbsolutePosition // 用于判断 xrun (min)
	not_after  AbsolutePosition // 用于判断 xrun (max)

	dispatcher innerClockEventDispatcher

	cache     *innerClockObserverCache
	observers map[ClockObserverID]*ClockObserver
	channels  map[ClockChannelID]*ClockChannel
}

func (inst *innerClockContext) nextObserverID() ClockObserverID {
	inst.observerIDGen++
	return inst.observerIDGen
}

func (inst *innerClockContext) getCache() (*innerClockObserverCache, error) {
	cache := inst.cache
	if cache == nil {
		loader := new(innerClockObserverCacheLoader)
		c2, err := loader.load(inst)
		if err != nil {
			return nil, err
		}
		cache = c2
		inst.cache = c2
	}
	return cache, nil
}

////////////////////////////////////////////////////////////////////////////////

type ClockChannelID byte

const (
	CCID_none ClockChannelID = iota

	CCID_frame_now
	CCID_frame_preload
	CCID_step_now
	CCID_step_preload
)

////////////////////////////////////////////////////////////////////////////////

type ClockChannel struct {
	ccid     ClockChannelID // [step_now|step_pre|frame_now|frame_pre]
	interval TimeLength     // 周期 (间隔)
	unit     ClockUnit
	preload  bool
	priority int

	index1 int // point_idx(a)
	index2 int // point_idx(b)

	timeline *Timeline // step | frame

	observers []*ClockObserverHolder
}

func (inst *ClockChannel) checkXrun(ctx *innerClockContext, pos AbsolutePosition) error {

	min := ctx.not_before
	max := ctx.not_after

	if min <= pos && pos <= max {
		return nil
	}

	ctx.countXrun++
	return fmt.Errorf("xrun")
}

func (inst *ClockChannel) innerTryUpdateInterval() {
	tline := inst.timeline
	if tline == nil {
		return
	}
	switch tline.Type {
	case core.TimelineTypeStep:
		inst.setInterval(tline.Length)
	case core.TimelineTypeFrame:
		inst.setInterval(tline.Length)
	}
}

func (inst *ClockChannel) setInterval(interval TimeLength) TimeLength {
	const (
		min = 0.001
		max = 60
	)
	if interval < min {
		interval = min
	}
	if interval > max {
		interval = max
	}
	inst.interval = interval
	return interval
}

func (inst *ClockChannel) runFastForward(ctx *innerClockContext, dest AbsolutePosition) error {

	if inst == nil || ctx == nil {
		return fmt.Errorf("param(s) is nil")
	}

	var err error
	var point RelativePosition

	interval := inst.interval
	from := inst.index1
	to := inst.index2
	tline := inst.timeline

	if interval < 0.001 {
		return fmt.Errorf("interval is too small, value=%f", interval)
	}

	dest.Relative(tline, &point)
	to_idx := float64(point.Pos) / float64(interval)
	to = int(to_idx)

	defer func() {
		inst.index1 = to
		inst.index2 = to
	}()

	point.Pos = Seconds(from) * Seconds(interval)
	point.Ref = tline
	err = inst.checkXrun(ctx, point.Absolute())
	if err != nil {
		return err
	}

	for i := from; i < to; i++ {

		point.Pos = Seconds(interval) * Seconds(i)

		event := &ClockEvent{
			Index:      uint(i),
			Clock:      ctx.clock,
			Unit:       inst.unit,
			UnitLength: interval,
			Preload:    inst.preload,
			Timeline:   inst.timeline,
			Realtime:   ctx.now,
			Thistime:   point.Absolute(),
		}

		err := inst.dispatch(event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *ClockChannel) dispatchWithHolder(event *ClockEvent, holder *ClockObserverHolder) error {

	if inst == nil || event == nil || holder == nil {
		return nil
	}

	ob := holder.observer
	if ob == nil {
		return nil
	}

	handler := ob.Handler
	if handler == nil {
		return nil
	}

	return handler(event)
}

func (inst *ClockChannel) dispatch(event *ClockEvent) error {

	var err error
	holders := inst.observers

	for _, holder := range holders {
		err = inst.dispatchWithHolder(event, holder)
		if err != nil {
			core.HandleError(err)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// EOF
