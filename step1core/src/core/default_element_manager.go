package core

import (
	"fmt"
	"sync"

	"github.com/ak47less/step1go/step1core"
)

// DefaultElementManager 是引擎中默认的 element 管理器.
//
// 它负责:
//   - 通过 [step1core.ElementFactory] 创建 element (缓存的创建过程不重复);
//   - 按 [step1core.ElementID] 保存已经加载的 element, 供 FindElement() 查找;
//   - 用 cache 保存加载顺序, 供 GetElement() / ListAll() 使用.
type DefaultElementManager struct {
	mutex sync.RWMutex

	factory step1core.ElementFactory

	table map[step1core.ElementID]*step1core.ElementInfo
	cache []step1core.Element
}

// 确保实现相关接口
var _ step1core.ElementManager = (*DefaultElementManager)(nil)

// Count implements [step1core.ElementManager].
//
// 返回已经加载的 element 数量.
func (inst *DefaultElementManager) Count() int {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	return len(inst.cache)
}

// GetElement implements [step1core.ElementManager].
//
// 按加载顺序返回 element, index 越界时返回 nil.
func (inst *DefaultElementManager) GetElement(index int) step1core.Element {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	if index < 0 || index >= len(inst.cache) {
		return nil
	}

	return inst.cache[index]
}

// ListAll implements [step1core.ElementManager].
//
// 返回当前所有 element 的快照, 调用方修改返回值不会影响管理器内部状态.
func (inst *DefaultElementManager) ListAll() []step1core.Element {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	size := len(inst.cache)
	list := make([]step1core.Element, size)
	copy(list, inst.cache)

	return list
}

// FindElement implements [step1core.ElementManager].
//
// 按 ID 查找已经加载的 element, 找不到时返回错误.
func (inst *DefaultElementManager) FindElement(ei *step1core.ElementInfo) (step1core.Element, error) {

	if ei == nil {
		return nil, fmt.Errorf("element info is nil")
	}

	id := ei.ID
	if id == "" {
		return nil, fmt.Errorf("element id is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	element := inst.innerFindByID(id)
	if element == nil {
		return nil, fmt.Errorf("element not found, id: %s", id)
	}

	return element, nil
}

// LoadElement implements [step1core.ElementManager].
//
// 加载(创建并缓存)一个 element:
//   - ei.Element 已经存在时直接使用它, 不再调用 factory;
//   - 同一个 ID 重复加载时返回已经缓存的实例;
//   - 其它情况通过 factory 创建, 成功后按 ID 缓存并追加到 cache.
func (inst *DefaultElementManager) LoadElement(ei *step1core.ElementInfo) (step1core.Element, error) {

	if ei == nil {
		return nil, fmt.Errorf("element info is nil")
	}

	id := ei.ID
	if id == "" {
		return nil, fmt.Errorf("element id is empty")
	}
	if ei.Class == "" {
		return nil, fmt.Errorf("element class is empty, id: %s", id)
	}

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	// 已经加载过: 直接复用
	old := inst.innerFindByID(id)
	if old != nil {
		return old, nil
	}

	element := ei.Element
	if element == nil {
		// 创建 element
		factory := inst.factory
		if factory == nil {
			return nil, fmt.Errorf("element factory is nil, id: %s", id)
		}
		created, err := factory.CreateElement(ei)
		if err != nil {
			return nil, err
		}
		if created == nil {
			return nil, fmt.Errorf("element is nil, id: %s", id)
		}
		element = created
	}

	// 缓存
	info := new(step1core.ElementInfo)
	*info = *ei
	info.Element = element

	if inst.table == nil {
		inst.table = make(map[step1core.ElementID]*step1core.ElementInfo)
	}
	inst.table[id] = info
	inst.cache = append(inst.cache, element)

	step1core.Log().Debug("element loaded, id: %s, class: %s", id, info.Class)

	return element, nil
}

// SetFactory 设置用于创建 element 的工厂.
//
// 由引擎装配阶段调用; 参数为 nil 时忽略本次设置.
func (inst *DefaultElementManager) SetFactory(factory step1core.ElementFactory) {

	if factory == nil {
		step1core.Log().Error("element factory is nil, ignore")
		return
	}

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	inst.factory = factory
}

// Clear 移除所有已经加载的 element(工厂设置保持不变).
func (inst *DefaultElementManager) Clear() {

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	inst.table = nil
	inst.cache = nil
}

////////////////////////////////////////////////////////////////////////////////
// 内部方法

// innerFindByID 按 ID 查找 element, 不加锁, 调用方必须自己持有锁.
func (inst *DefaultElementManager) innerFindByID(id step1core.ElementID) step1core.Element {

	info := inst.table[id]
	if info == nil {
		return nil
	}

	return info.Element
}

// FindInfo 按 ID 查找 element 的注册信息, 找不到时返回错误.
func (inst *DefaultElementManager) FindInfo(id step1core.ElementID) (*step1core.ElementInfo, error) {

	if id == "" {
		return nil, fmt.Errorf("element id is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	info := inst.table[id]
	if info == nil {
		return nil, fmt.Errorf("element not found, id: %s", id)
	}

	return info, nil
}

func (inst *DefaultElementManager) _impl() step1core.ElementManager {
	return inst
}
