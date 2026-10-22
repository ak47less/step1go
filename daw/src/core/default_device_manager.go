package core

import (
	"fmt"
	"sync"

	"github.com/ak47less/step1go/step1core"
)

// DefaultDeviceManager 是引擎中默认的 Device 管理器.
//
// 它负责:
//   - 通过 [step1core.DeviceFactory] 创建 Device (缓存的创建过程不重复);
//   - 按 [step1core.DeviceID] 保存已经加载的 Device, 供 FindDevice() 查找;
//   - 用 cache 保存加载顺序, 供 GetDevice() / ListAll() 使用.
type DefaultDeviceManager struct {
	mutex sync.RWMutex

	context *step1core.EngineContext

	table map[step1core.DeviceID]*step1core.DeviceInfo
	cache []step1core.Device
}

// 确保实现相关接口
var _ step1core.DeviceManager = (*DefaultDeviceManager)(nil)

// Count implements [step1core.DeviceManager].
//
// 返回已经加载的 Device 数量.
func (inst *DefaultDeviceManager) Count() int {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	return len(inst.cache)
}

// GetDevice implements [step1core.DeviceManager].
//
// 按加载顺序返回 Device, index 越界时返回 nil.
func (inst *DefaultDeviceManager) GetDevice(index int) step1core.Device {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	if index < 0 || index >= len(inst.cache) {
		return nil
	}

	return inst.cache[index]
}

// ListAll implements [step1core.DeviceManager].
//
// 返回当前所有 Device 的快照, 调用方修改返回值不会影响管理器内部状态.
func (inst *DefaultDeviceManager) ListAll() []step1core.Device {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	size := len(inst.cache)
	list := make([]step1core.Device, size)
	copy(list, inst.cache)

	return list
}

// FindDevice implements [step1core.DeviceManager].
//
// 按 ID 查找已经加载的 Device, 找不到时返回错误.
func (inst *DefaultDeviceManager) FindDevice(ei *step1core.DeviceInfo) (step1core.Device, error) {

	if ei == nil {
		return nil, fmt.Errorf("Device info is nil")
	}

	id := ei.ID
	if id == "" {
		return nil, fmt.Errorf("Device id is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	Device := inst.innerFindByID(id)
	if Device == nil {
		return nil, fmt.Errorf("Device not found, id: %s", id)
	}

	return Device, nil
}

// LoadDevice implements [step1core.DeviceManager].
//
// 加载(创建并缓存)一个 Device:
//   - ei.Device 已经存在时直接使用它, 不再调用 factory;
//   - 同一个 ID 重复加载时返回已经缓存的实例;
//   - 其它情况通过 factory 创建, 成功后按 ID 缓存并追加到 cache.
func (inst *DefaultDeviceManager) LoadDevice(ei *step1core.DeviceInfo) (step1core.Device, error) {

	if ei == nil {
		return nil, fmt.Errorf("Device info is nil")
	}

	id := ei.ID
	if id == "" {
		return nil, fmt.Errorf("Device id is empty")
	}
	if ei.Class == "" {
		return nil, fmt.Errorf("Device class is empty, id: %s", id)
	}

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	// 已经加载过: 直接复用
	old := inst.innerFindByID(id)
	if old != nil {
		return old, nil
	}

	Device := ei.Device
	ctx := inst.context

	if Device == nil {
		// 创建 Device
		factory := ctx.DeviceFactory
		if factory == nil {
			return nil, fmt.Errorf("Device factory is nil, id: %s", id)
		}
		created, err := factory.CreateDevice(ei)
		if err != nil {
			return nil, err
		}
		if created == nil {
			return nil, fmt.Errorf("Device is nil, id: %s", id)
		}
		Device = created
	}

	// 缓存
	info := new(step1core.DeviceInfo)
	*info = *ei
	info.Device = Device

	if inst.table == nil {
		inst.table = make(map[step1core.DeviceID]*step1core.DeviceInfo)
	}
	inst.table[id] = info
	inst.cache = append(inst.cache, Device)

	step1core.Log().Debug("Device loaded, id: %s, class: %s", id, info.Class)

	return Device, nil
}

func (inst *DefaultDeviceManager) Init(ctx *step1core.EngineContext) error {

	if ctx == nil {
		return fmt.Errorf("Device factory is nil, ignore")
	}

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	inst.context = ctx
	return nil
}

// Clear 移除所有已经加载的 Device(工厂设置保持不变).
func (inst *DefaultDeviceManager) Clear() {

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	inst.table = nil
	inst.cache = nil
}

////////////////////////////////////////////////////////////////////////////////
// 内部方法

// innerFindByID 按 ID 查找 Device, 不加锁, 调用方必须自己持有锁.
func (inst *DefaultDeviceManager) innerFindByID(id step1core.DeviceID) step1core.Device {

	info := inst.table[id]
	if info == nil {
		return nil
	}

	return info.Device
}

// FindInfo 按 ID 查找 Device 的注册信息, 找不到时返回错误.
func (inst *DefaultDeviceManager) FindInfo(id step1core.DeviceID) (*step1core.DeviceInfo, error) {

	if id == "" {
		return nil, fmt.Errorf("Device id is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	info := inst.table[id]
	if info == nil {
		return nil, fmt.Errorf("Device not found, id: %s", id)
	}

	return info, nil
}

func (inst *DefaultDeviceManager) _impl() step1core.DeviceManager {
	return inst
}
