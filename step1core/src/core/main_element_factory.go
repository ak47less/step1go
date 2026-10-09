package core

import (
	"fmt"
	"sync"

	"github.com/ak47less/step1go/step1core"
)

// MainElementFactory 是引擎中默认的 element 工厂.
//
// 它同时实现 [step1core.ElementRegistry] 与 [step1core.ElementFactory]:
//   - 各个 element 模块在初始化时, 通过 Register() 把自己的
//     [step1core.ElementProvider] 注册进来;
//   - 引擎在加载配置时, 通过 CreateElement() 按 class 创建 element 实例.
//
// 本工厂只负责按 class 做分发, 不关心具体 element 的创建细节,
// 创建动作由注册进来的 [step1core.ElementFactory] 完成.
type MainElementFactory struct {
	mutex sync.RWMutex
	table map[step1core.ElementClass]*step1core.ElementRegistration
}

// 确保实现相关接口
var _ step1core.ElementFactory = (*MainElementFactory)(nil)
var _ step1core.ElementRegistry = (*MainElementFactory)(nil)

// Register implements [step1core.ElementRegistry].
//
// 参数非法(provider 为空)或者 class 已经被注册时, 只记录错误日志, 不中断调用方,
// 这样单个 element 模块的问题不会影响整个引擎的启动.
func (inst *MainElementFactory) Register(provider step1core.ElementProvider) {

	logger := step1core.Log()

	if provider == nil {
		logger.Error("element provider is nil")
		return
	}

	registration := provider.GetRegistration()
	if registration == nil {
		logger.Error("element registration is nil")
		return
	}

	class := registration.Class
	if class == "" {
		logger.Error("element class is empty")
		return
	}

	factory := registration.Factory
	if factory == nil {
		// registration 里没有带上 factory 时, 退回到 provider 自己提供的 factory
		factory = provider.GetFactory()
	}
	if factory == nil {
		logger.Error("element factory is nil, class: %s", class)
		return
	}

	// 复制一份 registration, 避免调用方在注册之后改动它影响工厂内部状态
	item := new(step1core.ElementRegistration)
	*item = *registration
	item.Class = class
	item.Provider = provider
	item.Factory = factory

	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	if inst.table == nil {
		inst.table = make(map[step1core.ElementClass]*step1core.ElementRegistration)
	}

	_, exists := inst.table[class]
	if exists {
		logger.Error("element class already registered: %s", class)
		return
	}

	inst.table[class] = item
	logger.Info("element class registered: %s", class)
}

// CreateElement implements [step1core.ElementFactory].
//
// 根据 ei.Class 找到对应的注册项, 再委托给注册的 factory 创建 element.
func (inst *MainElementFactory) CreateElement(ei *step1core.ElementInfo) (step1core.Element, error) {

	if ei == nil {
		return nil, fmt.Errorf("element info is nil")
	}

	class := ei.Class
	if class == "" {
		return nil, fmt.Errorf("element class is empty")
	}

	registration, err := inst.FindRegistration(class)
	if err != nil {
		return nil, err
	}

	factory := registration.Factory
	if factory == nil {
		return nil, fmt.Errorf("element factory is nil, class: %s", class)
	}

	// 传给具体 factory 的信息里, class 必须与注册时的一致
	inner := new(step1core.ElementInfo)
	*inner = *ei
	inner.Class = class

	element, err := factory.CreateElement(inner)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, fmt.Errorf("element is nil, class: %s", class)
	}

	return element, nil
}

////////////////////////////////////////////////////////////////////////////////
// 内部方法

// FindRegistration 按 class 查找注册项, 找不到时返回错误.
func (inst *MainElementFactory) FindRegistration(class step1core.ElementClass) (*step1core.ElementRegistration, error) {

	if class == "" {
		return nil, fmt.Errorf("element class is empty")
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	registration := inst.table[class]
	if registration == nil {
		return nil, fmt.Errorf("element class not registered: %s", class)
	}

	return registration, nil
}

// ContainsClass 判断某个 class 是否已经注册.
func (inst *MainElementFactory) ContainsClass(class step1core.ElementClass) bool {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	return inst.table[class] != nil
}

// ListClasses 列出所有已经注册的 class.
func (inst *MainElementFactory) ListClasses() []step1core.ElementClass {

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	classes := make([]step1core.ElementClass, 0, len(inst.table))
	for class := range inst.table {
		classes = append(classes, class)
	}

	return classes
}

func (inst *MainElementFactory) _impl() (step1core.ElementFactory, step1core.ElementRegistry) {
	return inst, inst
}
