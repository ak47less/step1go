package lib

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ak47less/step1go/lite2/core"
)

// Clock 定义应用主时钟的观察者管理接口。
//
// 主时钟根据音频引擎的当前时间驱动 step/frame 两类逻辑时钟，
// 并按预加载、实时两个时间窗口调用观察者。Clock 的具体实现由
// NewClock 创建，并作为 Part 注册到 ApplicationContext 中；实际
// 状态在应用生命周期进入 OnInit 阶段后才会初始化。
type Clock interface {
	// AddListener 通过 ClockListener 一次性取得一组观察者，并将它们
	// 注册到时钟中。返回实际注册成功的观察者列表。
	AddListener(l ClockListener) []*ClockObserver

	// AddObserver 注册一个观察者。传入观察者的 ID 为 0 时，时钟会
	// 自动分配 ID；返回被注册的观察者指针。
	AddObserver(ob *ClockObserver) *ClockObserver

	// RemoveObserver 移除指定 ID 的观察者。移除不存在的 ID 不会报错。
	RemoveObserver(id ClockObserverID)
}

// NewClock 创建一个尚未初始化的主时钟。
//
// 返回的 Clock 只是主时钟的外壳，应用上下文、定时器和观察者缓存等
// 运行状态会在其生命周期 OnInit 阶段建立。因此，在时钟完成初始化
// 之前不应调用观察者注册相关方法。
func NewClock() Clock {
	return new(innerMainClock)
}

////////////////////////////////////////////////////////////////////////////////

// ClockUnit 标识时钟事件使用的逻辑时间单位。
//
// step 对应 StepTimeline 上的步进事件，frame 对应 FrameTimeline 上的
// 帧事件。每个时钟单位分别拥有预加载和实时两条分发通道。
type ClockUnit byte

const (
	// ClockUnitNone 表示未指定时钟单位。
	ClockUnitNone ClockUnit = iota

	// ClockUnitStep 表示按步进时间线产生的事件。
	ClockUnitStep

	// ClockUnitFrame 表示按帧时间线产生的事件。
	ClockUnitFrame
)

////////////////////////////////////////////////////////////////////////////////

// ClockEvent 描述一次逻辑时钟推进产生的单个刻度事件。
//
// 同一个事件只会分发给与 Unit、Preload 匹配的观察者。预加载事件用于
// 提前准备后续内容，实时事件则表示该刻度已经到了应当执行的位置。
type ClockEvent struct {
	// Clock 是产生该事件的主时钟。
	Clock Clock

	// Index 是该事件在当前通道时间线上的刻度序号，从 0 开始。
	Index uint

	// Unit 是事件所属的逻辑时间单位，取值为 step 或 frame。
	Unit ClockUnit // ( step|frame )

	// UnitLength 是单个逻辑单位对应的秒数，也是相邻事件的间隔。
	UnitLength TimeLength // ( step_length | frame_length )

	// Timeline 是事件所属的 step/frame 时间线。
	Timeline *Timeline

	// Realtime 是分发这一批事件时使用的实时绝对位置。预加载事件与
	// 实时事件可能具有不同的 Realtime。
	Realtime AbsolutePosition

	// Thistime 是当前刻度对应的逻辑绝对位置，即
	// Timeline.GetTotalOffset() + Index*UnitLength。
	Thistime AbsolutePosition

	// Preload 为 true 时表示预加载事件，为 false 时表示实时事件。
	Preload bool //  ( 1=preload, 0=now )
}

// ClockEventHandlerFunc 是观察者处理单个时钟事件的回调。
//
// ClockChannel 会为每个观察者调用该回调。当前分发实现会记录回调返回的
// 错误但继续处理其他观察者，不会把错误返回给 ClockChannel.dispatch。
type ClockEventHandlerFunc func(c *ClockEvent) error

////////////////////////////////////////////////////////////////////////////////

// ClockObserverID 是时钟观察者在一次时钟实例中的唯一标识。
//
// 零值保留为“尚未分配 ID”，AddObserver 会为其生成非零 ID。
type ClockObserverID int

// ClockObserver 描述一个时钟事件订阅。
type ClockObserver struct {
	// ID 是观察者标识；注册时若为 0，会自动分配新 ID。
	ID ClockObserverID

	// Clock 是观察者注册到的主时钟，由调用方按需填写。
	Clock Clock

	// Priority 是保留的优先级字段。当前缓存加载与事件分发不会依据它
	// 排序，实际调用顺序由缓存中的通道和观察者顺序决定。
	Priority int

	// Preload 选择预加载通道（true）或实时通道（false）。
	Preload bool

	// Unit 选择 step 或 frame 事件通道。
	Unit ClockUnit

	// Handler 是收到匹配时钟事件时调用的处理函数。
	Handler ClockEventHandlerFunc
}

////////////////////////////////////////////////////////////////////////////////

// ClockListener 用于批量提供时钟观察者。
//
// AddListener 会先调用 Listen，再把返回的每个非空观察者交给 AddObserver。
type ClockListener interface {
	// Listen 返回需要注册到 c 的观察者；返回 nil 表示没有观察者。
	Listen(c Clock) []*ClockObserver
}

////////////////////////////////////////////////////////////////////////////////

// innerMainClock 是 Clock 的默认实现，也是主时钟状态的外部入口。
//
// cc 在生命周期 OnInit 前为 nil；初始化后由它保存全部运行状态。
type innerMainClock struct {
	cc *innerClockContext
}

// AddObserver 注册或替换一个观察者，并在需要时分配观察者 ID。
//
// 如果 ID 已存在，新观察者会覆盖旧观察者。观察者指针由调用方提供，
// 时钟只保存该指针，不会复制它。注册表的修改受互斥锁保护，同时会使
// 观察者缓存失效，以便下一次分发重新构建。
//
// inst 或 ob 为 nil 时返回 nil。inst 必须已经完成初始化，否则没有
// 可用的时钟上下文。
//
// AddObserver implements [Clock].
func (inst *innerMainClock) AddObserver(ob *ClockObserver) *ClockObserver {

	if inst == nil || ob == nil {
		return nil
	}

	if ob.Handler == nil {
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

// RemoveObserver 从注册表中移除观察者，并立即使观察者缓存失效。
//
// 实现只把映射中的对应项置为 nil，不删除映射键。移除不存在的 ID 不会
// 产生错误；inst 必须已经完成初始化。
//
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

// init 使用应用上下文初始化主时钟的内部状态。
//
// 这是幂等操作：同一时钟已有内部上下文时直接返回。初始化会设置
// 预加载和缓冲区默认值，绑定引擎与文档时间线，并创建 step/frame 的
// 预加载、实时共四条标准分发通道。该方法由生命周期回调调用，不是
// 并发安全的公开初始化入口。
func (inst *innerMainClock) init(ac *ApplicationContext) {

	if inst == nil || ac == nil {
		return
	}

	cc := inst.cc
	if cc != nil {
		return
	}

	tls := &ac.EngineContext.Timelines

	cc = new(innerClockContext)
	cc.observers = make(map[ClockObserverID]*ClockObserver)
	cc.channels = make(map[ClockChannelID]*ClockChannel)
	cc.observerIDGen = 0x123456
	cc.preloadLength = 1
	cc.bufferLength = 3
	cc.context = ac
	cc.clock = inst
	cc.timelineForEngine = tls.Engine
	cc.timelineForDocument = tls.Document

	inst.innerAddChannel(cc, ClockUnitStep, true, 4)
	inst.innerAddChannel(cc, ClockUnitStep, false, 3)
	inst.innerAddChannel(cc, ClockUnitFrame, true, 2)
	inst.innerAddChannel(cc, ClockUnitFrame, false, 1)

	inst.cc = cc
}

// innerAddChannel 创建并登记一条标准时钟通道。
//
// unit 决定使用 StepTimeline 还是 FrameTimeline，preload 决定通道
// 对应的时间窗口，priority 作为通道元数据保存。初始间隔为 0.5 秒，
// 后续分发时会根据绑定时间线的 Length 更新。
func (inst *innerMainClock) innerAddChannel(cc *innerClockContext, unit ClockUnit, preload bool, priority int) {

	chl := &ClockChannel{
		unit:     unit,
		preload:  preload,
		priority: priority,
		interval: 0.5,
	}

	ec := cc.context.EngineContext
	tls := &ec.Timelines

	switch unit {
	case ClockUnitFrame:
		if preload {
			chl.ccid = CCID_frame_preload
		} else {
			chl.ccid = CCID_frame_now
		}
		chl.timeline = tls.Frame
	case ClockUnitStep:
		if preload {
			chl.ccid = CCID_step_preload
		} else {
			chl.ccid = CCID_step_now
		}
		chl.timeline = tls.Step
	}

	cc.channels[chl.ccid] = chl
}

// AddListener 通过监听器批量注册观察者。
//
// 如果时钟或监听器为 nil，返回 nil。监听器返回列表中的 nil 项会被
// 跳过；注册失败的观察者也不会出现在结果中。返回值按监听器给出的
// 原始顺序排列。
//
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

// innerFetchNow 计算当前批次应当使用的实时绝对位置。
//
// 当 TransportContext.StopAt 已设置时，该目标位置优先于音频时钟，用于
// 把时钟稳定地停在指定位置。否则读取音频上下文当前时间，并将其从
// EngineTimeline 的相对位置换算成绝对位置。
func (inst *innerMainClock) innerFetchNow() AbsolutePosition {

	cc := inst.cc
	ac := cc.context
	ec := ac.EngineContext
	stop_at := ec.Positions.StopAt

	if stop_at != nil {
		return stop_at.Absolute()
	}

	ctime := ec.AudioContext.GetCurrentTime()
	tls := &ec.Timelines

	var current RelativePosition
	current.Pos = ctime
	current.Ref = tls.Engine

	return current.Absolute()
}

// innerOnInit 把主时钟绑定到应用上下文并完成内部初始化。
//
// 初始化后，应用上下文可通过 Clock 字段访问该实例。
func (inst *innerMainClock) innerOnInit(ac *ApplicationContext) error {

	inst.init(ac)
	ac.Clock = inst

	return nil
}

// innerOnTimer 是核心定时器的周期回调，用于推进一次时钟分发。
//
// 每次调用先累计回调次数，再采样实时位置并交给事件分发器。分发器会
// 依次刷新观察者缓存、计算预加载时间窗口，并推进每个通道。返回的错误
// 会由底层 Timer 的错误处理流程记录。
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

// innerOnCreate 创建驱动主时钟的核心定时器。
//
// 定时器首次延迟 0.1 秒，之后每 0.05 秒（20 Hz）调用一次
// innerOnTimer。创建阶段只保存定时器，不启动它。
func (inst *innerMainClock) innerOnCreate() error {

	timer := &core.Timer{
		Delay:    0.1,
		Interval: (1.0 / 20),
		Handler:  inst.innerOnTimer,
	}

	inst.cc.timer = timer

	return nil
}

// innerOnStart 启动主时钟定时器，使应用运行时开始接收时钟事件。
func (inst *innerMainClock) innerOnStart() error {
	inst.cc.timer.Start()
	return nil
}

// innerOnStop 停止主时钟定时器，应用停止后不再产生新的时钟事件。
func (inst *innerMainClock) innerOnStop() error {
	inst.cc.timer.Stop()
	return nil
}

// GetLifeCycle 返回主时钟接入应用生命周期所需的回调集合。
//
// 时钟使用 Order 0，依次绑定初始化、创建定时器、启动和停止定时器；
// 暂停与恢复由底层音频时间是否推进间接体现。
//
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

// _impl 是编译期接口实现标记，确保主时钟同时满足 Clock 和 Life。
// 该方法不应由业务代码调用。
func (inst *innerMainClock) _impl() (Clock, Life) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////

// innerClockEventDispatcher 按固定阶段推进一次时钟批次。
//
// procedures 缓存阶段函数列表。首次分发时按检查、准备、扩展点和通道
// 分发顺序建立列表，后续直接复用。
type innerClockEventDispatcher struct {
	procedures []func(cc *innerClockContext) error
}

// dispatch 使用实时位置 now 推进一次完整时钟批次。
//
// 方法先计算预加载位置 now+preloadLength，并更新 xrun 判定窗口，然后
// 依次执行已经缓存的阶段函数。任一阶段返回错误时会立即中止本批次，
// 后续阶段不会执行。
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

// innerDoPrepare 确保观察者缓存已建立，为后续通道分发准备数据。
//
// 当前实现只加载缓存；缓存未失效时不会重复重建。
func (inst *innerClockEventDispatcher) innerDoPrepare(cc *innerClockContext) error {

	cc.getCache()

	return nil
}

// innerDoCheck 在准备阶段之前检查并刷新观察者缓存。
//
// 当前不执行额外的一致性检查。首次调用会触发缓存加载，后续调用直接
// 命中 prepare 阶段已经建立的缓存。
func (inst *innerClockEventDispatcher) innerDoCheck(cc *innerClockContext) error {

	cc.getCache()

	return nil
}

// innerDoX 是分发流程中的扩展点，当前不执行任何操作。
func (inst *innerClockEventDispatcher) innerDoX(c *innerClockContext) error {

	return nil
}

// innerDoUpdateContextPositions 更新上下文中的实时位置
func (inst *innerClockEventDispatcher) innerDoUpdateContextPositions(c *innerClockContext) error {

	appctx := c.context
	posset := &appctx.EngineContext.Positions
	now := c.now
	pre := c.preload

	tline := appctx.EngineContext.Timelines.Engine
	raw := now.Relative(tline, nil)

	posset.Raw = raw
	posset.Preload = pre
	posset.Current = now

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

// innerDoDispatchToAllChannels 推进观察者缓存中的全部时钟通道。
//
// 预加载通道以 cc.preload 为目标位置，实时通道以 cc.now 为目标位置。
// 每个通道会先按时间线刷新事件间隔，再从当前索引快速推进到目标索引。
// 某个通道返回错误时会停止后续通道的处理。
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

// innerGetProcedures 返回分发阶段函数列表，并在首次调用时完成初始化。
//
// 当前顺序为：缓存检查、缓存准备、扩展点、所有通道分发。该列表会
// 缓存在 dispatcher 中，受主时钟单一定时器调用模型约束。
func (inst *innerClockEventDispatcher) innerGetProcedures() []func(c *innerClockContext) error {
	list := inst.procedures
	if len(list) == 0 {

		list = append(list, inst.innerDoCheck)
		list = append(list, inst.innerDoPrepare)
		list = append(list, inst.innerDoUpdateContextPositions)
		list = append(list, inst.innerDoDispatchToAllChannels)

		// list = append(list, inst.innerDoX)

		inst.procedures = list
	}
	return list
}

////////////////////////////////////////////////////////////////////////////////

// ClockObserverHolder 保存观察者与其所属通道的关联。
//
// 它避免每次分发时重新判断观察者的 Unit/Preload，也是缓存加载阶段的
// 中间结果。
type ClockObserverHolder struct {
	// ccid 是观察者匹配到的通道 ID。
	ccid ClockChannelID
	// observer 是原始观察者；其 Handler 保存实际回调。
	observer *ClockObserver
	// channel 是该观察者将要使用的通道。
	channel *ClockChannel
}

// ClockChannelHolder 保存一条通道及其在当前缓存中的观察者集合。
type ClockChannelHolder struct {
	// ccid 是通道标识。
	ccid ClockChannelID
	// channel 是原始通道对象。
	channel *ClockChannel
	// obs 是按 Unit/Preload 匹配到该通道的观察者。
	obs []*ClockObserverHolder
}

// innerClockObserverCache 是观察者注册表按通道分组后的分发快照。
//
// 注册或移除观察者会把 innerClockContext.cache 置空；下一次分发再根据
// 当前 observers 和 channels 生成新的快照。
type innerClockObserverCache struct {
	channels []*ClockChannelHolder
}

// findByUnitPre 按时间单位和预加载标志查找对应的通道缓存。
//
// 查找采用线性扫描；找不到匹配通道时返回 nil。
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

// innerClockObserverCacheLoader 从一个时钟上下文构建观察者缓存快照。
//
// 加载分成通道复制、观察者归类、通道回写三个步骤，便于分别维护和
// 测试每个阶段。
type innerClockObserverCacheLoader struct {
}

// load 构建完整的通道与观察者缓存。
//
// 方法要求 cc 非 nil，并记录一次缓存加载日志。它不会修改注册表，
// 但在最后一步会把归类后的观察者列表写回各个 ClockChannel。
func (inst *innerClockObserverCacheLoader) load(cc *innerClockContext) (*innerClockObserverCache, error) {

	if cc == nil {
		return nil, fmt.Errorf("innerClockContext is nil")
	}

	mu := &cc.mutex
	logger := cc.context.Logger

	mu.Lock()
	defer mu.Unlock()

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

	inst.innerSortChannels(dst)
	inst.innerSortObersevers(dst)

	err = inst.innerUpdateChannels(dst)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// innerLoadChannels 把时钟上下文中的标准通道复制到缓存。
//
// nil 通道会被跳过。当前通道保存在 map 中，因此遍历产生的缓存顺序
// 不保证与 map 的插入顺序一致。
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

// innerLoadObservers 按 Unit 和 Preload 把观察者分到对应通道。
//
// nil 观察者或找不到匹配通道的观察者会被忽略。这里只建立观察者与
// 通道持有者的关联，不改动 ClockChannel.observers。观察者保存在 map
// 中，因此同一次加载得到的观察者顺序不保证稳定。
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

func (inst *innerClockObserverCacheLoader) innerSortChannels(cache *innerClockObserverCache) {

	list := cache.channels

	sort.Slice(list, func(i, j int) bool {
		o1 := list[i].channel
		o2 := list[j].channel
		return (o1.priority > o2.priority)
	})
}

func (inst *innerClockObserverCacheLoader) innerSortObersevers(cache *innerClockObserverCache) {

	list1 := cache.channels

	for _, it1 := range list1 {
		chl := it1.channel
		list2 := chl.observers
		sort.Slice(list2, func(i, j int) bool {
			o1 := list2[i].observer
			o2 := list2[j].observer
			return (o1.Priority > o2.Priority)
		})
	}
}

// innerUpdateChannels 把缓存中归类好的观察者列表回写到各通道。
//
// 通道会在后续 runFastForward 中遍历该列表分发事件。nil 持有者或
// nil 通道会被跳过。
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

// innerClockContext 保存主时钟运行期间共享的全部状态。
//
// 它把应用上下文、定时器、时间窗口、观察者注册表和分发器集中在一起，
// 再传给各个内部辅助方法。除观察者注册表的 mutex 外，该结构本身不是
// 完整的并发安全容器，主要由主时钟定时器串行访问。
type innerClockContext struct {
	// context 是主时钟绑定的应用上下文。
	context *ApplicationContext

	// clock 指回暴露给调用方的主时钟。
	clock Clock
	// timer 是驱动 20 Hz 分发的核心定时器。
	timer *core.Timer
	// mutex 保护 observers、observerIDGen 和 cache 的注册表修改。
	mutex sync.Mutex
	// observerIDGen 是最近一次分配的观察者 ID，新 ID 在此基础上递增。
	observerIDGen ClockObserverID

	// preloadLength 是预加载窗口相对实时位置的前瞻时长，默认 1 秒。
	preloadLength TimeLength // 预加载时长(提前量)
	// bufferLength 是实时位置允许回溯的容差，默认 3 秒。
	bufferLength TimeLength // 缓冲区时长, 超过这个值会触发 xrun

	// timelineForDocument 保存文档时间线，当前主时钟分发逻辑未直接使用。
	timelineForDocument *Timeline
	// timelineForEngine 保存引擎时间线，音频当前时间以它为参考。
	timelineForEngine *Timeline

	// countOnTimer 记录主时钟定时器回调累计次数。
	countOnTimer uint
	// countXrun 记录通道位置越出允许时间窗口的累计次数。
	countXrun uint

	// now 是本批次采样的实时绝对位置。
	now AbsolutePosition // 当前的实时位置
	// preload 是本批次预加载窗口的绝对位置，即 now+preloadLength。
	preload AbsolutePosition // 当前的预加载位置
	// not_before 是 xrun 判定窗口允许的最小绝对位置。
	not_before AbsolutePosition // 用于判断 xrun (min)
	// not_after 是 xrun 判定窗口允许的最大绝对位置。
	not_after AbsolutePosition // 用于判断 xrun (max)

	// dispatcher 保存本批次的阶段流程和缓存。
	dispatcher innerClockEventDispatcher

	// cache 是观察者按通道分组后的最新快照；nil 表示需要重建。
	cache *innerClockObserverCache
	// observers 保存全部非空观察者注册项，键为观察者 ID。
	observers map[ClockObserverID]*ClockObserver
	// channels 保存四条标准 step/frame 分发通道，键为通道 ID。
	channels map[ClockChannelID]*ClockChannel
}

// nextObserverID 生成下一个非零观察者 ID。
//
// 该方法不自行加锁，调用方必须持有 innerClockContext.mutex。
func (inst *innerClockContext) nextObserverID() ClockObserverID {
	inst.observerIDGen++
	return inst.observerIDGen
}

// getCache 返回观察者缓存，并在缓存缺失或失效时重新加载。
//
// 加载失败会保留 nil 缓存并返回错误。该方法本身不加锁；注册表修改与
// 缓存读取的协调依赖主时钟的调用时序。
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

// ClockChannelID 标识主时钟预建的四种标准分发通道。
type ClockChannelID byte

const (
	// CCID_none 表示未指定通道。
	CCID_none ClockChannelID = iota

	// CCID_frame_now 分发实时 frame 事件。
	CCID_frame_now
	// CCID_frame_preload 分发预加载 frame 事件。
	CCID_frame_preload
	// CCID_step_now 分发实时 step 事件。
	CCID_step_now
	// CCID_step_preload 分发预加载 step 事件。
	CCID_step_preload
)

////////////////////////////////////////////////////////////////////////////////

// ClockChannel 管理一个逻辑时间单位上的连续刻度分发。
//
// 通道保存上次推进到的索引，并通过 runFastForward 补齐当前批次覆盖的
// 所有刻度。每个通道只处理与其 unit、preload 匹配的观察者。
type ClockChannel struct {
	// ccid 是通道标识。
	ccid ClockChannelID // [step_now|step_pre|frame_now|frame_pre]
	// interval 是相邻刻度之间的秒数，由时间线长度计算并限制在合理范围。
	interval TimeLength // 周期 (间隔)
	// unit 是通道处理的 step/frame 时间单位。
	unit ClockUnit
	// preload 表示该通道使用预加载窗口还是实时窗口。
	preload bool
	// priority 是保留的通道优先级元数据，当前加载流程未据此排序。
	priority int

	// index1 是当前批次开始推进前的时间线索引。
	index1 int // point_idx(a)
	// index2 是当前批次开始推进前记录的目标索引。
	index2 int // point_idx(b)

	// timeline 是 step 或 frame 逻辑时间线。
	timeline *Timeline // step | frame

	// observers 是缓存加载后匹配到该通道的观察者。
	observers []*ClockObserverHolder
}

// checkXrun 检查绝对位置是否落在当前允许的时间窗口内。
//
// 允许范围为 [not_before, not_after]。位置越界时累加 countXrun，并
// 返回带有 "xrun" 文本的错误；位置合法时返回 nil。
func (inst *ClockChannel) checkXrun(ctx *innerClockContext, pos AbsolutePosition) error {

	min := ctx.not_before
	max := ctx.not_after

	if min <= pos && pos <= max {
		return nil
	}

	ctx.countXrun++
	return fmt.Errorf("xrun")
}

// innerTryUpdateInterval 根据绑定时间线的类型和长度刷新通道间隔。
//
// 只有 step 和 frame 时间线会更新间隔；timeline 为 nil 或其他类型时
// 保持原值。实际赋值会经过 setInterval 的上下限约束。
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

// setInterval 设置通道刻度间隔并返回约束后的值。
//
// 间隔被限制在 [0.001, 60] 秒，避免极小间隔造成过密分发，也避免单个
// 刻度跨越过长的时间范围。
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

// runFastForward 将通道从当前索引推进到 dest 对应的目标索引。
//
// 方法先把目标绝对位置转换为该通道时间线上的相对位置，再除以间隔得到
// 目标索引。对区间 [from, to) 中的每个刻度，它创建一个 ClockEvent 并
// 分发给通道观察者。无论成功还是中途失败，退出时都会把 index1 和
// index2 更新为目标索引，避免下一次重复处理已经检查过的区间。
//
// 推进前会以当前通道索引对应的时间位置检查 xrun。观察者回调返回的
// 错误由 dispatch 记录，不会从本方法返回；本方法直接返回的是参数、
// 间隔或 xrun 错误。
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

// dispatchWithHolder 调用单个观察者的处理函数。
//
// 任一参数或观察者 Handler 为 nil 时直接返回 nil。
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

// dispatch 把事件依次发送给通道内的全部观察者。
//
// 单个观察者的错误通过 core.HandleError 记录，分发仍会继续给后续
// 观察者；方法最终始终返回 nil。
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
