package core

import (
	"fmt"
	"sort"

	"github.com/ak47less/step1go/step1core"
)

// DefaultLifeManager 是引擎中默认的 life-cycle 管理器.
//
// 它负责:
//   - 收集各个 Device 与 API 注册进来的 [step1core.Life];
//   - 通过 ListAll() 给出排重并按 Order 递增排序之后的列表;
//   - 通过 GetMain() 给出一个"主生命周期": 它把注册进来的各个 Life 合成一个
//     整体, 由引擎统一驱动.
//
// 与 [step1core.DeviceManager] 不同, 这里的顺序由 Life.Order 决定,
// 而不是注册顺序.
type DefaultLifeManager struct {
	_main step1core.LifeAPI

	buffer []*step1core.Life // 未经加工的 items
	cached []*step1core.Life // 经过排序 & 排重的 items (按照 item.order 递增 )
}

// 确保实现相关接口
var _ step1core.LifeManager = (*DefaultLifeManager)(nil)

// AddAPI implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddAPI(api step1core.LifeAPI) {
	if inst == nil || api == nil {
		return
	}
	life := api.GetLife()
	inst.innerAdd(life)
}

// AddDevice implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddDevice(el step1core.Device) {

	api, ok := el.(step1core.LifeAPI)
	if !ok {
		return
	}
	if api == nil {
		return
	}
	life := api.GetLife()
	inst.innerAdd(life)
}

// AddLife implements [step1core.LifeManager].
func (inst *DefaultLifeManager) AddLife(life *step1core.Life) {
	inst.innerAdd(life)
}

func (inst *DefaultLifeManager) innerAdd(life *step1core.Life) {
	if life == nil || inst == nil {
		return
	}
	inst.buffer = append(inst.buffer, life)

	// 内容变了, 缓存(排序表 & 主生命周期)全部失效
	inst.cached = nil
	inst._main = nil
}

// GetMain implements [step1core.LifeManager].
func (inst *DefaultLifeManager) GetMain() step1core.LifeAPI {
	m := inst._main
	if m == nil {
		m = inst.innerLoadMain()
		inst._main = m
	}
	return m
}

// innerLoadMain 合成"主生命周期" (main-life).
//
// 参与合成的 items 是 ListAll() 给出的有序列表; 打开(open)类的回调按 Order
// 递增依次调用, 关闭(close)类的回调按 Order 递减依次调用, 详见 [mainLife].
func (inst *DefaultLifeManager) innerLoadMain() step1core.LifeAPI {
	return newMainLife(inst.ListAll())
}

// ListAll implements [step1core.LifeManager].
//
// 返回全部 life 的快照, 具有以下保证:
//   - 已经排除 nil, 并按指针排重 (同一个 *Life 只出现一次);
//   - 按照 [step1core.Life].Order 递增排序, Order 相同时保持注册顺序;
//   - 返回的是副本, 调用方修改它不会影响管理器内部状态.
func (inst *DefaultLifeManager) ListAll() []*step1core.Life {

	if inst == nil {
		return nil
	}

	cached := inst.cached
	if cached == nil {
		// buffer 有变化时才重新整理, 结果缓存在 cached 里
		cached = inst.innerLoadCache()
		inst.cached = cached
	}

	list := make([]*step1core.Life, len(cached))
	copy(list, cached)
	return list
}

// innerLoadCache 排重 & 排序 buffer 中的 items.
func (inst *DefaultLifeManager) innerLoadCache() []*step1core.Life {

	list := make([]*step1core.Life, 0, len(inst.buffer))
	seen := make(map[*step1core.Life]bool)

	for _, life := range inst.buffer {
		if life == nil || seen[life] {
			continue
		}
		seen[life] = true
		list = append(list, life)
	}

	// 用稳定排序, 让 Order 相同的 items 保持注册顺序
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Order < list[j].Order
	})

	return list
}

func (inst *DefaultLifeManager) _impl() step1core.LifeManager {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// main-life : 主生命周期

// mainLife 是 [DefaultLifeManager] 合成的"主生命周期".
//
// 它把注册进来的各个 [step1core.Life] 串成一个整体, 供引擎统一驱动:
//   - 打开(open)类的回调按照 Order 递增调用, 一旦出错立即返回;
//   - 关闭(close)类的回调按照 Order 递减调用, 出错时记录日志后继续,
//     保证每个元素都有机会做清理, 最后返回第一个错误.
//
// 回调执行期间发生的 panic 会被捕获, 并转换成 error.
type mainLife struct {
	life  step1core.Life
	items []*step1core.Life
}

// 确保实现 [step1core.LifeAPI]
var _ step1core.LifeAPI = (*mainLife)(nil)

// newMainLife 用一组已经排好序的 Life 合成一个 mainLife.
func newMainLife(items []*step1core.Life) *mainLife {

	inst := new(mainLife)
	inst.items = items
	inst.life.Order = 0
	inst.life.Label = "main"

	// on_engine
	inst.life.OnEngineCreate = inst.onEngineCreate
	inst.life.OnEngineLoad = inst.onEngineLoad
	inst.life.OnEngineReload = inst.onEngineReload
	inst.life.OnEngineStart = inst.onEngineStart
	inst.life.OnEngineUpdate = inst.onEngineUpdate
	inst.life.OnEngineStop = inst.onEngineStop
	inst.life.OnEngineDestroy = inst.onEngineDestroy

	// on_transport
	inst.life.OnTransportStart = inst.onTransportStart
	inst.life.OnTransportResume = inst.onTransportResume
	inst.life.OnTransportPause = inst.onTransportPause
	inst.life.OnTransportStop = inst.onTransportStop

	return inst
}

// GetLife implements [step1core.LifeAPI].
func (inst *mainLife) GetLife() *step1core.Life {
	return &inst.life
}

////////////////////////////////////////////////////////////////////////////////
// main-life : on_engine

func (inst *mainLife) onEngineCreate(e step1core.Engine) error {
	return inst.invokeEngineOpen(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineCreate
	})
}

func (inst *mainLife) onEngineLoad(e step1core.Engine) error {
	return inst.invokeEngineOpen(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineLoad
	})
}

func (inst *mainLife) onEngineReload(e step1core.Engine) error {
	return inst.invokeEngineOpen(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineReload
	})
}

func (inst *mainLife) onEngineStart(e step1core.Engine) error {
	return inst.invokeEngineOpen(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineStart
	})
}

// Update 是在循环中反复触发的回调, 所以和"打开"类回调一样按 Order 递增调用
func (inst *mainLife) onEngineUpdate(e step1core.Engine) error {
	return inst.invokeEngineOpen(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineUpdate
	})
}

// Stop 标志着循环结束, 按 Order 递减调用
func (inst *mainLife) onEngineStop(e step1core.Engine) error {
	return inst.invokeEngineClose(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineStop
	})
}

// Destroy 在引擎销毁时触发, 按 Order 递减调用
func (inst *mainLife) onEngineDestroy(e step1core.Engine) error {
	return inst.invokeEngineClose(e, func(l *step1core.Life) step1core.EngineLifeFunction {
		return l.OnEngineDestroy
	})
}

////////////////////////////////////////////////////////////////////////////////
// main-life : on_transport

func (inst *mainLife) onTransportStart(t step1core.Transport) error {
	return inst.invokeTransportOpen(t, func(l *step1core.Life) step1core.TransportLifeFunction {
		return l.OnTransportStart
	})
}

// Resume 与 Start 同向, 属于"打开" transport 的动作
func (inst *mainLife) onTransportResume(t step1core.Transport) error {
	return inst.invokeTransportOpen(t, func(l *step1core.Life) step1core.TransportLifeFunction {
		return l.OnTransportResume
	})
}

// Pause 与 Stop 同向, 属于"关闭" transport 的动作
func (inst *mainLife) onTransportPause(t step1core.Transport) error {
	return inst.invokeTransportClose(t, func(l *step1core.Life) step1core.TransportLifeFunction {
		return l.OnTransportPause
	})
}

func (inst *mainLife) onTransportStop(t step1core.Transport) error {
	return inst.invokeTransportClose(t, func(l *step1core.Life) step1core.TransportLifeFunction {
		return l.OnTransportStop
	})
}

////////////////////////////////////////////////////////////////////////////////
// main-life : 内部方法

// invokeEngineOpen 按 Order 递增调用各个 Life 的 engine 回调, 遇到错误立即返回.
func (inst *mainLife) invokeEngineOpen(e step1core.Engine, pick func(*step1core.Life) step1core.EngineLifeFunction) error {

	for _, item := range inst.items {
		if item == nil {
			continue
		}
		fn := pick(item)
		if fn == nil {
			continue
		}
		err := innerInvokeEngine(fn, e)
		if err != nil {
			return innerWrapLifeError(item, err)
		}
	}

	return nil
}

// invokeEngineClose 按 Order 递减调用各个 Life 的 engine 回调;
// 出错时记录日志并继续, 最后返回第一个错误.
func (inst *mainLife) invokeEngineClose(e step1core.Engine, pick func(*step1core.Life) step1core.EngineLifeFunction) error {

	logger := step1core.Log()
	var firstErr error

	for i := len(inst.items) - 1; i >= 0; i-- {

		item := inst.items[i]
		if item == nil {
			continue
		}
		fn := pick(item)
		if fn == nil {
			continue
		}

		err := innerInvokeEngine(fn, e)
		if err != nil {
			err = innerWrapLifeError(item, err)
			logger.Warn("life close failed: %v", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// invokeTransportOpen 与 invokeEngineOpen 相同, 只是回调的类型不同.
func (inst *mainLife) invokeTransportOpen(t step1core.Transport, pick func(*step1core.Life) step1core.TransportLifeFunction) error {

	for _, item := range inst.items {
		if item == nil {
			continue
		}
		fn := pick(item)
		if fn == nil {
			continue
		}
		err := innerInvokeTransport(fn, t)
		if err != nil {
			return innerWrapLifeError(item, err)
		}
	}

	return nil
}

// invokeTransportClose 与 invokeEngineClose 相同, 只是回调的类型不同.
func (inst *mainLife) invokeTransportClose(t step1core.Transport, pick func(*step1core.Life) step1core.TransportLifeFunction) error {

	logger := step1core.Log()
	var firstErr error

	for i := len(inst.items) - 1; i >= 0; i-- {

		item := inst.items[i]
		if item == nil {
			continue
		}
		fn := pick(item)
		if fn == nil {
			continue
		}

		err := innerInvokeTransport(fn, t)
		if err != nil {
			err = innerWrapLifeError(item, err)
			logger.Warn("life close failed: %v", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// innerWrapLifeError 给回调返回的错误附上 life 的信息.
func innerWrapLifeError(item *step1core.Life, err error) error {

	if item == nil || err == nil {
		return err
	}

	return fmt.Errorf("life error, order: %d, label: %s: %w", item.Order, item.Label, err)
}

// innerInvokeEngine 调用一个 engine 回调, 并把 panic 转换成 error.
func innerInvokeEngine(fn step1core.EngineLifeFunction, e step1core.Engine) (err error) {

	defer func() {
		x := recover()
		if x != nil {
			err = fmt.Errorf("panic: %v", x)
		}
	}()

	return fn(e)
}

// innerInvokeTransport 调用一个 transport 回调, 并把 panic 转换成 error.
func innerInvokeTransport(fn step1core.TransportLifeFunction, t step1core.Transport) (err error) {

	defer func() {
		x := recover()
		if x != nil {
			err = fmt.Errorf("panic: %v", x)
		}
	}()

	return fn(t)
}
