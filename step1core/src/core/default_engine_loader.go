package core

import (
	"fmt"

	"github.com/ak47less/step1go/step1core"
)

// DefaultEngineLoader 是引擎中默认的加载器.
//
// 它负责在引擎启动之前:
//   - 把 [step1core.EngineConfiguration] 里声明的 element 逐个创建出来,
//     并交给 [step1core.ElementManager] 缓存;
//   - 把配置里声明的连线([step1core.WireConfig])按名字连接到对应的端口上.
//
// 创建 element 的具体动作由 ElementManager 内部的 [step1core.ElementFactory]
// 完成, 本加载器只负责遍历配置, 校验与连线.
type DefaultEngineLoader struct {
}

// 确保实现相关接口
var _ step1core.EngineLoader = (*DefaultEngineLoader)(nil)

// PortTable 是 element 端口的索引表.
//
// 键是 [step1core.ElementName] 或者 [step1core.ElementID] (配置里两者都可能被
// 引用), 值是该 element 的全部端口(按名称索引).
type PortTable struct {
	table map[string]*PortTableItem
}

// PortTableItem 是单个 element 的端口集合.
type PortTableItem struct {
	element step1core.Element

	table map[step1core.PortName][]step1core.Port
}

// Load implements [step1core.EngineLoader].
//
// 根据配置加载相应的 elements, 然后按配置把端口连接(wire)起来.
//
// 加载失败时直接返回错误, 这样调用方不会拿到一个"半可用"的引擎.
func (inst *DefaultEngineLoader) Load(en step1core.Engine) error {

	if en == nil {
		return fmt.Errorf("engine is nil")
	}

	ctx := en.GetContext()
	if ctx == nil {
		return fmt.Errorf("engine context is nil")
	}

	cfg := en.GetConfiguration()
	if cfg == nil {
		return fmt.Errorf("engine configuration is nil")
	}

	manager := ctx.ElementManager
	if manager == nil {
		return fmt.Errorf("element manager is nil")
	}

	logger := step1core.Log()

	// 同一个 ID 出现两次时, ElementManager 会复用第一个实例, 这多半是配置写错了,
	// 所以在这里先检查一遍, 把问题暴露出来.
	ids := make(map[step1core.ElementID]bool)
	count := 0

	for index, item := range cfg.Elements {

		if item == nil {
			return fmt.Errorf("element config is nil, index: %d", index)
		}

		class := item.Class
		if class == "" {
			return fmt.Errorf("element class is empty, index: %d", index)
		}

		id := item.ID
		if id == "" {
			return fmt.Errorf("element id is empty, index: %d", index)
		}

		if ids[id] {
			return fmt.Errorf("duplicated element id: %s", id)
		}
		ids[id] = true

		ei := new(step1core.ElementInfo)
		ei.Class = class
		ei.ID = id
		ei.Name = item.Name
		ei.Engine = en

		_, err := manager.LoadElement(ei)
		if err != nil {
			return fmt.Errorf("load element failed, id: %s, class: %s: %w", id, class, err)
		}

		count++
		logger.Info("element loaded, id: %s, class: %s", id, class)
	}

	logger.Info("elements loaded, count: %d", count)

	// 建立连线
	return inst.innerWireElements(en)
}

// innerWireElements 根据配置, 把已经加载的 elements 连接(wire)起来.
//
// 配置中的 [step1core.PortRef] 用 element 名称 + 端口名称来引用一个端口;
// 为了容错, 名称找不到时会退回到用 ID 匹配.
func (inst *DefaultEngineLoader) innerWireElements(en step1core.Engine) error {

	if en == nil {
		return fmt.Errorf("engine is nil")
	}

	ctx := en.GetContext()
	if ctx == nil {
		return fmt.Errorf("engine context is nil")
	}

	cfg := en.GetConfiguration()
	if cfg == nil {
		return fmt.Errorf("engine configuration is nil")
	}

	manager := ctx.ElementManager
	if manager == nil {
		return fmt.Errorf("element manager is nil")
	}

	logger := step1core.Log()
	result := new(PortTable)
	count := 0

	// 收集所有 element 的端口
	for index := 0; index < manager.Count(); index++ {
		element := manager.GetElement(index)
		if element == nil {
			continue
		}
		item := result.innerPut(element)
		if item == nil {
			continue
		}
		logger.Debug("element ports collected, name: %s, count: %d", item.name(), len(item.table))
	}

	// 逐个建立连线
	for index, element := range cfg.Elements {

		if element == nil {
			return fmt.Errorf("element config is nil, index: %d", index)
		}

		for wireIndex, wc := range element.Wires {

			if wc == nil {
				return fmt.Errorf("wire config is nil, element id: %s, index: %d", element.ID, wireIndex)
			}

			_, err := inst.innerWireOne(result, wc)
			if err != nil {
				return fmt.Errorf("wire element failed, element id: %s, wire index: %d: %w", element.ID, wireIndex, err)
			}

			count++
			logger.Info("element wired, source: %s.%s, destination: %s.%s",
				wc.Source.Element, wc.Source.Port, wc.Destination.Element, wc.Destination.Port)
		}
	}

	logger.Info("elements wired, count: %d", count)

	return nil
}

// innerWireOne 建立一条连线.
func (inst *DefaultEngineLoader) innerWireOne(pt *PortTable, wc *step1core.WireConfig) (*step1core.Wire, error) {

	if pt == nil {
		return nil, fmt.Errorf("port table is nil")
	}
	if wc == nil {
		return nil, fmt.Errorf("wire config is nil")
	}

	src, err := pt.innerFindSource(&wc.Source)
	if err != nil {
		return nil, err
	}

	dst, err := pt.innerFindDestination(&wc.Destination)
	if err != nil {
		return nil, err
	}

	wire := new(step1core.Wire)
	wire.Source = src
	wire.Destination = dst

	// 两端都通知一次: Port 接口只提供 Wire(*Wire) 这一个入口,
	// 具体由哪一端记录连线, 交给具体实现决定.
	ok := false
	for _, port := range []step1core.Port{src, dst} {
		if port == nil {
			continue
		}
		ok = true
		err = port.Wire(wire)
		if err != nil {
			return nil, fmt.Errorf("port wire failed, element: %s, port: %s: %w",
				wc.Source.Element, wc.Source.Port, err)
		}
	}
	if !ok {
		return nil, fmt.Errorf("wire has no port")
	}

	return wire, nil
}

////////////////////////////////////////////////////////////////////////////////
// PortTable

// innerPut 把一个 element 的端口加入索引表, element 名称为空时返回 nil.
func (inst *PortTable) innerPut(element step1core.Element) *PortTableItem {

	if element == nil {
		return nil
	}

	info := element.GetInfo(nil)
	if info == nil {
		return nil
	}

	name := string(info.Name)
	if name == "" {
		// 没有名字的 element 无法被连线引用
		return nil
	}

	if inst.table == nil {
		inst.table = make(map[string]*PortTableItem)
	}

	item := inst.table[name]
	if item == nil {
		item = new(PortTableItem)
		item.element = element
		item.table = make(map[step1core.PortName][]step1core.Port)
		inst.table[name] = item
	}

	// 按 ID 也建立一份引用, 便于配置用 ID 引用
	id := string(info.ID)
	if id != "" {
		if _, exists := inst.table[id]; !exists {
			inst.table[id] = item
		}
	}

	for _, port := range element.GetPorts() {
		if port == nil {
			continue
		}
		pi := port.GetInfo(nil)
		if pi == nil {
			continue
		}
		item.innerAdd(&pi.Name, port)
	}

	return item
}

// innerAdd 登记一个端口.
func (inst *PortTableItem) innerAdd(name *step1core.PortName, port step1core.Port) {

	if inst.table == nil || name == nil || port == nil || *name == "" {
		return
	}

	list := inst.table[*name]
	if len(list) > 0 {
		// 同名端口多于一个时只使用第一个, 这里提示一下
		step1core.Log().Warn("duplicated port name: %s", *name)
	}

	inst.table[*name] = append(list, port)
}

// name 返回该 element 的名称(仅用于日志).
func (inst *PortTableItem) name() string {

	if inst == nil || inst.element == nil {
		return ""
	}

	info := inst.element.GetInfo(nil)
	if info == nil {
		return ""
	}

	return string(info.Name)
}

// innerFindSource 按引用查找一个输出端口.
func (inst *PortTable) innerFindSource(ref *step1core.PortRef) (step1core.Output, error) {

	port, err := inst.innerFind(ref)
	if err != nil {
		return nil, err
	}

	output, ok := port.(step1core.Output)
	if !ok {
		return nil, fmt.Errorf("port is not an output, element: %s, port: %s", ref.Element, ref.Port)
	}

	return output, nil
}

// innerFindDestination 按引用查找一个输入端口.
func (inst *PortTable) innerFindDestination(ref *step1core.PortRef) (step1core.Input, error) {

	port, err := inst.innerFind(ref)
	if err != nil {
		return nil, err
	}

	input, ok := port.(step1core.Input)
	if !ok {
		return nil, fmt.Errorf("port is not an input, element: %s, port: %s", ref.Element, ref.Port)
	}

	return input, nil
}

// innerFind 按 [step1core.PortRef] 查找端口.
//
// element 先按名称匹配, 找不到时按 ID 匹配; 端口按名称匹配.
func (inst *PortTable) innerFind(ref *step1core.PortRef) (step1core.Port, error) {

	if ref == nil {
		return nil, fmt.Errorf("port ref is nil")
	}
	if ref.Element == "" {
		return nil, fmt.Errorf("port ref element is empty")
	}

	key := string(ref.Element)
	item := inst.table[key]
	if item == nil {
		return nil, fmt.Errorf("element not found, element: %s", ref.Element)
	}
	if ref.Port == "" {
		return nil, fmt.Errorf("port ref port is empty, element: %s", ref.Element)
	}

	list := item.table[ref.Port]
	if len(list) == 0 {
		return nil, fmt.Errorf("port not found, element: %s, port: %s", ref.Element, ref.Port)
	}

	return list[0], nil
}

////////////////////////////////////////////////////////////////////////////////
// helper

// Unload 移除当前已经加载的所有 element.
//
// 重新加载(Reload)之前可以先调用它, 避免新旧 element 混在一起.
func (inst *DefaultEngineLoader) Unload(en step1core.Engine) error {

	if en == nil {
		return fmt.Errorf("engine is nil")
	}

	ctx := en.GetContext()
	if ctx == nil {
		return fmt.Errorf("engine context is nil")
	}

	manager := ctx.ElementManager
	if manager == nil {
		return fmt.Errorf("element manager is nil")
	}

	if inner, ok := manager.(*DefaultElementManager); ok {
		inner.Clear()
	}

	return nil
}

func (inst *DefaultEngineLoader) _impl() step1core.EngineLoader {
	return inst
}
