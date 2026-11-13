package lite

import (
	"fmt"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

// EventType 表示事件的分类，用于区分不同种类的事件。
type EventType int

// EventID 表示事件类型下的具体事件标识。
type EventID int

////////////////////////////////////////////////////////////////////////////////

const (
	// EventTypeMin 是内置事件类型范围的下界。
	EventTypeMin EventType = iota

	// EventTypeMax 是内置事件类型范围的上界。
	EventTypeMax
)

////////////////////////////////////////////////////////////////////////////////

// Event 描述一条待分发的事件。
type Event struct {
	Type    EventType // 事件分类
	ID      EventID   // 事件在分类中的具体标识
	Sender  any       // 触发事件的发送者
	Context *Context  // 事件关联的引擎上下文
}

// EventCallbackFunc 处理匹配的事件，并允许通过错误中止后续分发。
type EventCallbackFunc func(e *Event) error

// EventCallback 将一个回调函数与匹配的事件类型和标识绑定。
type EventCallback struct {
	Type EventType         // 回调匹配的事件分类
	ID   EventID           // 回调匹配的事件标识
	Fn   EventCallbackFunc // 事件匹配时执行的回调
}

// EventDispatcher 负责接收事件，并管理事件回调的注册和移除。
type EventDispatcher interface {
	// Dispatch 将一条事件加入待处理队列，由 EventRunner 后续处理。
	Dispatch(e *Event) error

	// AddCallback 注册一个事件回调。
	AddCallback(c *EventCallback) error

	// RemoveCallback 移除一个已注册的事件回调。
	RemoveCallback(c *EventCallback) error
}

// EventRunner 负责按投递顺序处理事件队列，并调用匹配的回调。
type EventRunner interface {
	// Run 使用运行器关联的事件上下文处理全部待处理事件。
	Run() error

	// RunContext 使用指定的事件上下文处理全部待处理事件。
	RunContext(c *EventContext) error
}

////////////////////////////////////////////////////////////////////////////////

// EventContext 汇集事件分发所需的调度器、回调、事件缓冲区和同步状态。
type EventContext struct {
	dispatcher EventDispatcher // 负责分发事件

	runner EventRunner // 负责驱动事件处理

	callbacks []*EventCallback // 已注册的事件回调

	buffer []*Event // 等待处理的事件缓冲区

	mutex sync.Mutex // 保护上下文中的可变状态
}

// GetDispatcher 返回事件上下文使用的分发器。
// 对于零值 EventContext，返回值为 nil。
func (inst *EventContext) GetDispatcher() EventDispatcher {
	return inst.dispatcher
}

// GetRunner 返回事件上下文使用的事件运行器。
// 对于零值 EventContext，返回值为 nil。
func (inst *EventContext) GetRunner() EventRunner {
	return inst.runner
}

////////////////////////////////////////////////////////////////////////////////

// innerEventDispatcher 是事件分发器的内部实现，并关联所属的事件上下文。
type innerEventDispatcher struct {
	context *EventContext // 分发器所属的事件上下文
}

// Dispatch 将事件副本追加到队列尾部；nil 上下文会返回错误。
// Dispatch implements [EventDispatcher].
func (inst *innerEventDispatcher) Dispatch(e *Event) error {

	if e == nil {
		return nil
	}

	if inst == nil || inst.context == nil {
		return fmt.Errorf("event dispatcher context is nil")
	}

	event := new(Event)
	*event = *e

	ctx := inst.context
	ctx.mutex.Lock()
	ctx.buffer = append(ctx.buffer, event)
	ctx.mutex.Unlock()
	return nil
}

// AddCallback 注册一个事件回调；nil 回调或 nil 处理函数会返回错误。
// AddCallback implements [EventDispatcher].
func (inst *innerEventDispatcher) AddCallback(c *EventCallback) error {
	if inst == nil || inst.context == nil {
		return fmt.Errorf("event dispatcher context is nil")
	}
	if c == nil {
		return fmt.Errorf("event callback is nil")
	}
	if c.Fn == nil {
		return fmt.Errorf("event callback function is nil")
	}

	ctx := inst.context
	ctx.mutex.Lock()
	ctx.callbacks = append(ctx.callbacks, c)
	ctx.mutex.Unlock()
	return nil
}

// RemoveCallback 移除通过同一指针注册的事件回调。
// RemoveCallback implements [EventDispatcher].
func (inst *innerEventDispatcher) RemoveCallback(c *EventCallback) error {
	if inst == nil || inst.context == nil {
		return fmt.Errorf("event dispatcher context is nil")
	}
	if c == nil {
		return fmt.Errorf("event callback is nil")
	}

	ctx := inst.context
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	for index, callback := range ctx.callbacks {
		if callback != c {
			continue
		}
		copy(ctx.callbacks[index:], ctx.callbacks[index+1:])
		ctx.callbacks[len(ctx.callbacks)-1] = nil
		ctx.callbacks = ctx.callbacks[:len(ctx.callbacks)-1]
		return nil
	}
	return fmt.Errorf("event callback is not registered")
}

////////////////////////////////////////////////////////////////////////////////

// innerEventRunner 是事件运行器的内部实现，并关联所属的事件上下文。
type innerEventRunner struct {
	context *EventContext // 运行器所属的事件上下文
}

// Run 处理运行器关联上下文中的全部待处理事件。
// Run implements [EventRunner].
func (inst *innerEventRunner) Run() error {
	if inst == nil || inst.context == nil {
		return fmt.Errorf("event runner context is nil")
	}
	return inst.RunContext(inst.context)
}

// RunContext 按 FIFO 顺序处理指定上下文中的事件，并在回调失败时停止。
// RunContext implements [EventRunner].
func (inst *innerEventRunner) RunContext(c *EventContext) error {
	if inst == nil {
		return fmt.Errorf("event runner is nil")
	}
	if c == nil {
		return fmt.Errorf("event context is nil")
	}

	for {
		event := c.innerTakeEvent()
		if event == nil {
			return nil
		}

		callbacks := c.innerMatchingCallbacks(event)
		for _, callback := range callbacks {
			if err := callback.Fn(event); err != nil {
				return fmt.Errorf(
					"dispatch event type %d id %d: %w",
					event.Type,
					event.ID,
					err,
				)
			}
		}
	}
}

// innerTakeEvent 从事件缓冲区头部取出一个事件；缓冲区为空时返回 nil。
func (inst *EventContext) innerTakeEvent() *Event {
	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	if len(inst.buffer) == 0 {
		return nil
	}
	event := inst.buffer[0]
	inst.buffer[0] = nil
	inst.buffer = inst.buffer[1:]
	return event
}

// innerMatchingCallbacks 返回与事件类型和标识匹配的回调快照。
func (inst *EventContext) innerMatchingCallbacks(event *Event) []*EventCallback {
	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	callbacks := make([]*EventCallback, 0)
	for _, callback := range inst.callbacks {
		if callback == nil || callback.Fn == nil {
			continue
		}
		if callback.Type == event.Type && callback.ID == event.ID {
			callbacks = append(callbacks, callback)
		}
	}
	return callbacks
}

////////////////////////////////////////////////////////////////////////////////

// NewEventContext 创建一个 EventContext，并安装相互关联的分发器和运行器实例。
func NewEventContext() *EventContext {

	ctx := new(EventContext)
	ctx.dispatcher = &innerEventDispatcher{context: ctx}
	ctx.runner = &innerEventRunner{context: ctx}

	return ctx
}

////////////////////////////////////////////////////////////////////////////////
// EOF
