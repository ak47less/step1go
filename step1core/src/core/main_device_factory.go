package core

import (
	"fmt"
	"sync"

	"github.com/ak47less/step1go/step1core"
)

// MainDeviceFactory 是引擎中默认的 Device 工厂.
//
// 它同时实现 [step1core.DeviceRegistry] 与 [step1core.DeviceFactory]:
//   - 各个 Device 模块在初始化时, 通过 Register() 把自己的
//     [step1core.DeviceProvider] 注册进来;
//   - 引擎在加载配置时, 通过 CreateDevice() 按 class 创建 Device 实例.
//
// 本工厂只负责按 class 做分发, 不关心具体 Device 的创建细节,
// 创建动作由注册进来的 [step1core.DeviceFactory] 完成.
type MainDeviceFactory struct {
	mutex sync.RWMutex
	table map[step1core.DeviceClass]*step1core.DeviceRegistration
}

// 确保实现相关接口
var _ step1core.DeviceFactory = (*MainDeviceFactory)(nil)
var _ step1core.DeviceRegistry = (*MainDeviceFactory)(nil)

// Register implements [step1core.DeviceRegistry].
//
// 参数非法(provider 为空)或者 class 已经被注册时, 只记录错误日志, 不中断调用方,
// 这样单个 Device 模块的问题不会影响整个引擎的启动.
func (inst *MainDeviceFactory) Register(provider step1core.DeviceProvider) {

	logger := step1core.Log()

	if provider == nil {
		logger.Error("Device provider is nil")
		return
	}

	registration := provider.GetRegistration()
	if registration == nil {
		logger.Error("Device registration is nil")
		return
	}

	class := registration.Class
	if class == "" {
		logger.Error("Device class is empty")
		return
	}

	factory := registration.Factory
	if factory == nil {
		// registration 里没有带上 factory 时, 退回到 provider 自己提供的 factory
		factory = provider.GetFactory()
	}
	if factory == nil {
		logger.Error("Device factory is nil, class: %s", class)
		return
	}

	// 复制一份 registration, 避免调用方在注册之后改动它影响工厂内部状态
	item := new(step1core.DeviceRegistration)
	*item = *registration
	item.Class = class
	item.Provider = provider
	item.Factory = factory

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	if inst.table == nil {
		inst.table = make(map[step1core.DeviceClass]*step1core.DeviceRegistration)
	}

	_, exists := inst.table[class]
	if exists {
		logger.Error("Device class already registered: %s", class)
		return
	}

	inst.table[class] = item
	logger.Info("Device class registered: %s", class)
}

// CreateDevice implements [step1core.DeviceFactory].
//
// 根据 ei.Class 找到对应的注册项, 再委托给注册的 factory 创建 Device.
func (inst *MainDeviceFactory) CreateDevice(ei *step1core.DeviceInfo) (step1core.Device, error) {

	if ei == nil {
		return nil, fmt.Errorf("Device info is nil")
	}

	class := ei.Class
	if class == "" {
		return nil, fmt.Errorf("Device class is empty")
	}

	registration, err := inst.FindRegistration(class)
	if err != nil {
		return nil, err
	}

	factory := registration.Factory
	if factory == nil {
		return nil, fmt.Errorf("Device factory is nil, class: %s", class)
	}

	// 传给具体 factory 的信息里, class 必须与注册时的一致
	inner := new(step1core.DeviceInfo)
	*inner = *ei
	inner.Class = class

	Device, err := factory.CreateDevice(inner)
	if err != nil {
		return nil, err
	}
	if Device == nil {
		return nil, fmt.Errorf("Device is nil, class: %s", class)
	}

	return Device, nil
}

////////////////////////////////////////////////////////////////////////////////
// 内部方法

// FindRegistration 按 class 查找注册项, 找不到时返回错误.
func (inst *MainDeviceFactory) FindRegistration(class step1core.DeviceClass) (*step1core.DeviceRegistration, error) {

	if class == "" {
		return nil, fmt.Errorf("Device class is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	registration := inst.table[class]
	if registration == nil {
		return nil, fmt.Errorf("Device class not registered: %s", class)
	}

	return registration, nil
}

// ContainsClass 判断某个 class 是否已经注册.
func (inst *MainDeviceFactory) ContainsClass(class step1core.DeviceClass) bool {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	return inst.table[class] != nil
}

// ListClasses 列出所有已经注册的 class.
func (inst *MainDeviceFactory) ListClasses() []step1core.DeviceClass {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	classes := make([]step1core.DeviceClass, 0, len(inst.table))
	for class := range inst.table {
		classes = append(classes, class)
	}

	return classes
}

func (inst *MainDeviceFactory) _impl() (step1core.DeviceFactory, step1core.DeviceRegistry) {
	return inst, inst
}
