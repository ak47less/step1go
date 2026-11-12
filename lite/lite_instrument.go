package lite

import (
	"fmt"
	"sort"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

// Instrument 表示能够接收 MIDI 消息并生成声音的乐器。
type Instrument interface {
	// Dispatch 分发一条 MIDI 消息，消息可包含音符按下或音符释放事件。
	Dispatch(msg *MidiMessage) error

	// GetInfo 返回乐器的名称、标签和预设信息。
	GetInfo() *InstrumentInfo
}

// InstrumentFactory 负责声明可创建的乐器，并根据配置创建乐器实例。
type InstrumentFactory interface {
	// ListInfo 返回工厂能够创建的乐器信息列表。
	ListInfo() []*InstrumentInfo

	// CreateInstrument 在给定引擎上下文中创建 info 描述的乐器。
	CreateInstrument(ctx *Context, info *InstrumentInfo) (Instrument, error)
}

// InstrumentManager 负责查询乐器信息，并为指定乐器选择对应的工厂。
type InstrumentManager interface {
	Part

	// ListInfo 返回管理器已注册的全部乐器信息。
	ListInfo() []*InstrumentInfo

	// GetFactory 返回能够创建 info 描述乐器的工厂。
	GetFactory(info *InstrumentInfo) (InstrumentFactory, error)
}

// InstrumentHolder 负责提供当前可用的乐器，并在需要时创建默认乐器。
type InstrumentHolder interface {
	// GetInstrument 返回上下文中的当前乐器，或创建一个默认乐器。
	GetInstrument(ctx *Context) (Instrument, error)
}

// InstrumentProvider 表示能够提供乐器注册信息的对象。
type InstrumentProvider interface {
	// ListRegistrations 返回该对象提供的全部乐器注册项。
	ListRegistrations() []*InstrumentRegistration
}

// InstrumentInfo 描述乐器在注册、查询和创建过程中使用的元数据。
type InstrumentInfo struct {
	Name   string // 乐器的唯一名称，也是注册表中的查找键
	Label  string // 面向用户显示的乐器名称
	Preset string // 创建乐器时采用的预设
}

////////////////////////////////////////////////////////////////////////////////

// InstrumentRegistration 将乐器元数据与其创建工厂绑定为一条注册项。
type InstrumentRegistration struct {
	InstrumentInfo

	Factory InstrumentFactory // 创建该乐器实例的工厂
}

////////////////////////////////////////////////////////////////////////////////

// InstrumentFactoryRegistry 维护按名称索引的乐器工厂，并支持默认工厂回退。
type InstrumentFactoryRegistry struct {
	table          map[string]*InstrumentRegistration // 按乐器名称索引的注册项
	factoryDefault InstrumentFactory                  // 未找到名称匹配项时使用的工厂
	mu             sync.RWMutex                       // 保护 table 和 factoryDefault
}

// GetLifeCycle 返回注册表作为零部件参与引擎生命周期时使用的钩子。
// GetLifeCycle implements [Life].
func (inst *InstrumentFactoryRegistry) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit:    inst.innerOnInit,
		OnCreate:  inst.innerOnCreate,
		OnDestroy: inst.innerOnDestroy,
	}
	return lc
}

// GetInfo 返回注册表作为 [Part] 所需的元数据。
// GetInfo implements [Part].
func (inst *InstrumentFactoryRegistry) GetInfo() *PartInfo {

	info := &PartInfo{}

	info.Name = "InstrumentFactoryRegistry"
	info.Classes = []PartClass{"InstrumentFactoryRegistry"}
	info.Part = inst

	return info
}

// CreateInstrument 根据乐器名称选择工厂；未注册的名称使用默认工厂创建。
// CreateInstrument implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) CreateInstrument(
	ctx *Context,
	info *InstrumentInfo,
) (Instrument, error) {
	if inst == nil {
		return nil, fmt.Errorf("instrument factory registry is nil")
	}
	if info == nil {
		return nil, fmt.Errorf("instrument info is nil")
	}

	inst.mu.RLock()
	ir := inst.table[info.Name]
	factory := inst.factoryDefault
	if ir != nil && ir.Factory != nil {
		factory = ir.Factory
	}
	inst.mu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("instrument %q is not registered and no default factory is set", info.Name)
	}

	return factory.CreateInstrument(ctx, info)
}

// SetDefault 设置名称未注册时使用的默认工厂；nil 参数会被忽略。
func (inst *InstrumentFactoryRegistry) SetDefault(f InstrumentFactory) {
	if inst == nil || f == nil {
		return
	}
	inst.mu.Lock()
	inst.factoryDefault = f
	inst.mu.Unlock()
}

// ListInfo 返回已注册乐器信息的副本，并按名称、标签和预设排序。
// ListInfo implements [InstrumentFactory].
func (inst *InstrumentFactoryRegistry) ListInfo() []*InstrumentInfo {
	if inst == nil {
		return nil
	}

	inst.mu.RLock()
	list := make([]*InstrumentInfo, 0, len(inst.table))
	for _, ir := range inst.table {
		if ir == nil {
			continue
		}
		item := ir.InstrumentInfo
		list = append(list, &item)
	}
	inst.mu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		if list[i].Label != list[j].Label {
			return list[i].Label < list[j].Label
		}
		return list[i].Preset < list[j].Preset
	})
	return list
}

// Register 将工厂声明的全部乐器注册到表中。
func (inst *InstrumentFactoryRegistry) Register(f InstrumentFactory) {
	if inst == nil || f == nil {
		return
	}
	for _, info := range f.ListInfo() {
		if info == nil {
			continue
		}
		inst.add(&InstrumentRegistration{
			InstrumentInfo: *info,
			Factory:        f,
		})
	}
}

// add 添加一条乐器注册项；相同名称的已有注册项会被替换。
func (inst *InstrumentFactoryRegistry) add(ir *InstrumentRegistration) {
	if inst == nil || ir == nil || ir.Factory == nil || ir.Name == "" {
		return
	}

	inst.mu.Lock()
	inst.initLocked()
	inst.table[ir.Name] = ir
	inst.mu.Unlock()
}

// init 在并发安全的前提下初始化内部注册表。
func (inst *InstrumentFactoryRegistry) init() {
	if inst == nil {
		return
	}

	inst.mu.Lock()
	inst.initLocked()
	inst.mu.Unlock()
}

// initLocked 初始化内部注册表；调用方必须持有 inst.mu 的写锁。
func (inst *InstrumentFactoryRegistry) initLocked() {
	if inst.table == nil {
		inst.table = make(map[string]*InstrumentRegistration)
	}
}

// innerOnInit 是注册表的初始化生命周期钩子，当前没有额外操作。
func (inst *InstrumentFactoryRegistry) innerOnInit(ec *Context) error {

	return nil
}

// innerOnCreate 是注册表的创建生命周期钩子，当前没有额外操作。
func (inst *InstrumentFactoryRegistry) innerOnCreate() error {

	return nil
}

// innerOnDestroy 是注册表的销毁生命周期钩子，当前没有额外操作。
func (inst *InstrumentFactoryRegistry) innerOnDestroy() error {

	return nil
}

// _impl 声明注册表实现了 [InstrumentFactory]、[Part] 和 [Life] 接口。
func (inst *InstrumentFactoryRegistry) _impl() (InstrumentFactory, Part, Life) {
	return inst, inst, inst
}

////////////////////////////////////////////////////////////////////////////////

// DefaultInstrumentHolder 按需创建并缓存上下文中的默认乐器。
type DefaultInstrumentHolder struct {
}

// GetInstrument 返回当前乐器；若不存在，则使用主乐器工厂创建默认三角波乐器。
// GetInstrument implements [InstrumentHolder].
func (inst *DefaultInstrumentHolder) GetInstrument(ctx *Context) (Instrument, error) {

	if ctx == nil {
		return nil, fmt.Errorf("lite.Context is nil")
	}

	i := ctx.CurrentInstrument
	if i != nil {
		return i, nil
	}

	factory := ctx.MainInstrumentFactory

	// const (
	// 	OscillatorTypeSine     OscillatorType = "sine"
	// 	OscillatorTypeSquare   OscillatorType = "square"
	// 	OscillatorTypeSawtooth OscillatorType = "sawtooth"
	// 	OscillatorTypeTriangle OscillatorType = "triangle"
	// 	OscillatorTypeCustom   OscillatorType = "custom"
	// )

	info := &InstrumentInfo{
		Name:   "",
		Preset: "triangle",
	}

	if factory == nil {
		return nil, fmt.Errorf("context.MainInstrumentFactory is nil")
	}

	i, err := factory.CreateInstrument(ctx, info)
	if err != nil {
		return nil, err
	}

	ctx.CurrentInstrument = i
	return i, nil
}

// _impl 声明默认乐器持有器实现了 [InstrumentHolder] 接口。
func (inst *DefaultInstrumentHolder) _impl() InstrumentHolder {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

// InstrumentRegistry 从引擎零部件中收集乐器注册信息。
type InstrumentRegistry struct {
	table map[string]*InstrumentRegistration // 按乐器名称索引的注册项
	mu    sync.RWMutex                       // 保护 table 和注册过程
}

// registerAllInstruments 扫描上下文中的全部零部件，并注册它们提供的乐器。
func (inst *InstrumentRegistry) registerAllInstruments(c *Context) error {
	if inst == nil {
		return fmt.Errorf("instrument registry is nil")
	}
	if c == nil {
		return fmt.Errorf("lite.Context is nil")
	}
	if c.PartManager == nil {
		return fmt.Errorf("context.PartManager is nil")
	}

	mu := &inst.mu
	mu.Lock()
	defer mu.Unlock()

	pm := c.PartManager
	namelist := pm.ListNames()
	table := inst.table

	if table == nil {
		table = make(map[string]*InstrumentRegistration)
	}

	for _, name := range namelist {

		part, err := pm.FindPartByName(name)
		if err != nil {
			return err
		}

		err = inst.registerInstrument(part, table)
		if err != nil {
			return err
		}
	}

	inst.table = table
	return nil
}

// registerInstrument 将单个零件提供的乐器写入 table。
func (inst *InstrumentRegistry) registerInstrument(part Part, table map[string]*InstrumentRegistration) error {
	if part == nil {
		return fmt.Errorf("instrument provider part is nil")
	}

	provider, ok := part.(InstrumentProvider)
	if !ok {
		return nil
	}

	registrations := provider.ListRegistrations()
	for _, registration := range registrations {
		if registration == nil {
			return fmt.Errorf("instrument registration is nil")
		}
		if registration.Name == "" {
			return fmt.Errorf("instrument registration name is empty")
		}
		if registration.Factory == nil {
			return fmt.Errorf("instrument %q has no factory", registration.Name)
		}
		if table == nil {
			return fmt.Errorf("instrument registration table is nil")
		}
		table[registration.Name] = registration
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// innerInstrumentManagerImpl 是 [InstrumentManager] 的内部实现。
type innerInstrumentManagerImpl struct {
	ir InstrumentRegistry // 保存从引擎零部件中收集的乐器注册信息
}

// GetFactory 返回指定乐器对应的工厂。
// GetFactory implements [InstrumentManager].
func (inst *innerInstrumentManagerImpl) GetFactory(info *InstrumentInfo) (InstrumentFactory, error) {
	if inst == nil {
		return nil, fmt.Errorf("instrument manager is nil")
	}
	if info == nil {
		return nil, fmt.Errorf("instrument info is nil")
	}

	inst.ir.mu.RLock()
	registration := inst.ir.table[info.Name]
	inst.ir.mu.RUnlock()
	if registration == nil {
		return nil, fmt.Errorf("instrument %q is not registered", info.Name)
	}
	if registration.Factory == nil {
		return nil, fmt.Errorf("instrument %q has no factory", info.Name)
	}
	return registration.Factory, nil
}

// ListInfo 返回已注册乐器的信息副本，并按名称、标签和预设排序。
// ListInfo implements [InstrumentManager].
func (inst *innerInstrumentManagerImpl) ListInfo() []*InstrumentInfo {
	if inst == nil {
		return nil
	}

	inst.ir.mu.RLock()
	list := make([]*InstrumentInfo, 0, len(inst.ir.table))
	for _, registration := range inst.ir.table {
		if registration == nil {
			continue
		}
		info := registration.InstrumentInfo
		list = append(list, &info)
	}
	inst.ir.mu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		if list[i].Label != list[j].Label {
			return list[i].Label < list[j].Label
		}
		return list[i].Preset < list[j].Preset
	})
	return list
}

// GetInfo 返回乐器管理器作为 [Part] 所需的元数据。
// GetInfo implements [Part].
func (inst *innerInstrumentManagerImpl) GetInfo() *PartInfo {
	b := new(PartInfoBuilder)
	b.SetName("InstrumentManager")
	return b.Build()
}

// GetLifeCycle 返回乐器管理器参与引擎生命周期时使用的钩子。
// GetLifeCycle implements [Life].
func (inst *innerInstrumentManagerImpl) GetLifeCycle() *LifeCycle {
	return (&LifeCycle{
		OnInit: inst.oninit,
	})
}

// oninit 在引擎初始化时收集并注册上下文中的全部乐器。
func (inst *innerInstrumentManagerImpl) oninit(c *Context) error {

	return inst.ir.registerAllInstruments(c)

}

// _impl 声明内部管理器实现了 [InstrumentManager]、[Part] 和 [Life] 接口。
func (inst *innerInstrumentManagerImpl) _impl() (InstrumentManager, Part, Life) {
	return inst, inst, inst
}

////////////////////////////////////////////////////////////////////////////////

// NewInstrumentManager 创建一个基于引擎零部件注册信息的乐器管理器。
func NewInstrumentManager() InstrumentManager {
	return new(innerInstrumentManagerImpl)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
