package core

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/ak47less/step1go/step1core"
)

// dom_impl.go : step1core/dom.go 中 DOM 的默认实现.
//
// 结构:
//
//	innerDOMImplementation    实现入口, 提供 Factory() 与 Codec()
//	├── innerDOMFactory       按 xsd 创建文档与各种 typed element
//	├── innerDOMDocument      work document (持有根元素 work)
//	└── innerDOMDocumentCodec XML 文本 <-> DOM
//
// 内部只有一种节点类型 domElement, 它实现了最基础的 [step1core.DOMElement];
// 与 xsd 中各 complexType 对应的 typed element (WorkElement / TrackElement ...)
// 则是对它的包装 (包装类型里内嵌 *domElement), 由 innerWrapElement() 按标签名
// 自动选择, 因此从子元素拿到的一定是合适的 typed 接口.
//
// 命名空间: 元素保存解析后的 namespace(URI) 与原 prefix; 编解码时以 prefix 写出,
// 并在需要时自动补上 xmlns / xmlns:prefix 声明 (元素自身的命名空间, 以及属性上用到的
// 众所周知的前缀 xml / xsi), 这样既能把文档读成 URI 形式, 又能保持原有的前缀写法.
//
// 已知简化 (与 xsd 的对应关系):
//   - 注释 / 处理指令 / DOCTYPE 在解码时被忽略, DOM 中不保留(接口里没有对应的节点类型);
//   - 不支持混合内容: 元素要么只有文本, 要么只有子元素, 二者同时出现时丢弃文本;
//     文本内容在解码时会去掉首尾空白 (与 xs:token 的折叠语义一致, 也便于忽略 XML 缩进);
//   - 属性按读入/写入顺序保存, 属性名匹配不区分"前缀:名字"与"名字"(两种写法都能命中),
//     解码时若属性前缀有 xmlns 声明, 则 Namespace 记下它的 URI, 否则 Namespace 为空串;
//   - 编码输出带 XML 声明, 缩进 4 个空格; 没有子元素与文本的元素写成自闭合标签.

////////////////////////////////////////////////////////////////////////////////
// 实现入口

// GetDomImplementation 返回 DOM 的默认实现.
func GetDomImplementation() step1core.DOMImplementation {
	return new(innerDOMImplementation)
}

type innerDOMImplementation struct {
	factory step1core.DOMFactory
	codec   step1core.DOMDocumentCodec
}

// 确保实现相关接口
var _ step1core.DOMImplementation = (*innerDOMImplementation)(nil)

// Codec implements [step1core.DOMImplementation].
func (inst *innerDOMImplementation) Codec() step1core.DOMDocumentCodec {
	if inst.codec == nil {
		inst.codec = &innerDOMDocumentCodec{impl: inst}
	}
	return inst.codec
}

// Factory implements [step1core.DOMImplementation].
func (inst *innerDOMImplementation) Factory() step1core.DOMFactory {
	if inst.factory == nil {
		inst.factory = &innerDOMFactory{impl: inst}
	}
	return inst.factory
}

////////////////////////////////////////////////////////////////////////////////
// 内部节点 : domElement

// domElement 是 DOM 中的元素节点.
type domElement struct {
	owner  step1core.DOMDocument // 所属文档
	parent *domElement           // 父元素 (根元素为 nil)

	namespace string // 解析后的命名空间 (URI)
	prefix    string // 前缀 (默认命名空间时为空串)
	tag       string // 标签名

	text       string                   // 文本内容
	attributes []step1core.DOMAttribute // 属性 (保持顺序)
	children   []*domElement            // 子元素 (保持顺序)

	// wrapper 是缓存的包装对象(view), 由 innerWrapElement() 填充;
	// 同一个内部节点始终返回同一个包装对象, 因此可以用 == 比较节点.
	wrapper step1core.DOMElement
}

// 确保实现 [step1core.DOMElement]
var _ step1core.DOMElement = (*domElement)(nil)

// innerNewElement 新建一个内部节点.
func innerNewElement(owner step1core.DOMDocument, namespace string, prefix string, tag string) *domElement {
	node := new(domElement)
	node.owner = owner
	node.namespace = namespace
	node.prefix = prefix
	node.tag = tag
	return node
}

// GetOwner implements [step1core.DOMNode].
func (inst *domElement) GetOwner() step1core.DOMDocument {
	if inst == nil {
		return nil
	}
	return inst.owner
}

// GetParent implements [step1core.DOMElement].
func (inst *domElement) GetParent() step1core.DOMElement {
	if inst == nil || inst.parent == nil {
		return nil
	}
	return innerWrapElement(inst.parent)
}

// GetNamespace implements [step1core.DOMElement].
func (inst *domElement) GetNamespace() string {
	if inst == nil {
		return ""
	}
	return inst.namespace
}

// GetPrefix implements [step1core.DOMElement].
func (inst *domElement) GetPrefix() string {
	if inst == nil {
		return ""
	}
	return inst.prefix
}

// GetTag implements [step1core.DOMElement].
func (inst *domElement) GetTag() string {
	if inst == nil {
		return ""
	}
	return inst.tag
}

// GetText implements [step1core.DOMElement].
func (inst *domElement) GetText() string {
	if inst == nil {
		return ""
	}
	return inst.text
}

// SetText implements [step1core.DOMElement].
func (inst *domElement) SetText(text string) error {
	if inst == nil {
		return fmt.Errorf("element is nil")
	}
	inst.text = text
	return nil
}

// GetAttribute implements [step1core.DOMElement].
func (inst *domElement) GetAttribute(name string) (string, bool) {
	index := inst.innerIndexOfAttribute(name)
	if index < 0 {
		return "", false
	}
	return inst.attributes[index].Value, true
}

// SetAttribute implements [step1core.DOMElement].
//
// 名字里带前缀时 (例如 "xsi:schemaLocation") 会被拆成 prefix 与 name 两部分保存;
// 取值允许为空串 (仍然写出该属性).
func (inst *domElement) SetAttribute(name string, value string) error {
	if inst == nil {
		return fmt.Errorf("element is nil")
	}
	if name == "" {
		return fmt.Errorf("attribute name is empty")
	}

	index := inst.innerIndexOfAttribute(name)
	if index >= 0 {
		inst.attributes[index].Value = value
		return nil
	}

	attr := step1core.DOMAttribute{Value: value}
	pos := strings.IndexByte(name, ':')
	if pos > 0 && pos < (len(name)-1) {
		attr.Prefix = name[:pos]
		attr.Name = name[pos+1:]
	} else {
		attr.Name = name
	}
	inst.attributes = append(inst.attributes, attr)
	return nil
}

// RemoveAttribute implements [step1core.DOMElement].
func (inst *domElement) RemoveAttribute(name string) error {
	if inst == nil {
		return fmt.Errorf("element is nil")
	}
	index := inst.innerIndexOfAttribute(name)
	if index < 0 {
		return nil
	}
	inst.attributes = slices.Delete(inst.attributes, index, index+1)
	return nil
}

// ListAttributes implements [step1core.DOMElement].
func (inst *domElement) ListAttributes() []step1core.DOMAttribute {
	if inst == nil {
		return nil
	}
	list := make([]step1core.DOMAttribute, len(inst.attributes))
	copy(list, inst.attributes)
	return list
}

// AddChild implements [step1core.DOMElement].
//
// 追加一个子元素; 子元素必须还没有父元素(否则返回错误), 加入时会随之改属本文档.
func (inst *domElement) AddChild(child step1core.DOMElement) error {
	node, err := innerCheckChild(child, "")
	if err != nil {
		return err
	}
	inst.innerAppendNode(node)
	return nil
}

// RemoveChild implements [step1core.DOMElement].
func (inst *domElement) RemoveChild(child step1core.DOMElement) error {
	if inst == nil {
		return fmt.Errorf("element is nil")
	}
	node, ok := innerUnwrapElement(child)
	if !ok {
		return fmt.Errorf("child element is nil or unsupported")
	}
	index := slices.Index(inst.children, node)
	if index < 0 {
		return fmt.Errorf("not a child element: %s", node.tag)
	}
	inst.children = slices.Delete(inst.children, index, index+1)
	node.parent = nil
	return nil
}

// ListChildren implements [step1core.DOMElement].
func (inst *domElement) ListChildren() []step1core.DOMElement {
	if inst == nil {
		return nil
	}
	list := make([]step1core.DOMElement, 0, len(inst.children))
	for _, child := range inst.children {
		list = append(list, innerWrapElement(child))
	}
	return list
}

// ListChildrenByTag implements [step1core.DOMElement].
func (inst *domElement) ListChildrenByTag(tag string) []step1core.DOMElement {
	if inst == nil {
		return nil
	}
	return innerWrapElements(inst.innerListChildNodes(tag))
}

////////////////////////////////////////////////////////////////////////////////
// 内部节点的辅助方法

// innerGetAttribute 读取属性, 不存在时返回空串.
func (inst *domElement) innerGetAttribute(name string) string {
	value, _ := inst.GetAttribute(name)
	return value
}

// innerIndexOfAttribute 查找属性, 返回下标 (找不到时返回 -1).
func (inst *domElement) innerIndexOfAttribute(name string) int {
	if inst == nil {
		return -1
	}
	for index := range inst.attributes {
		attr := &inst.attributes[index]
		if attr.Name == name || innerAttributeFullName(attr) == name {
			return index
		}
	}
	return -1
}

// innerFirstChild 查找第一个 tag 匹配的直接子元素.
func (inst *domElement) innerFirstChild(tag string) *domElement {
	if inst == nil {
		return nil
	}
	for _, child := range inst.children {
		if child.tag == tag {
			return child
		}
	}
	return nil
}

// innerListChildNodes 列出全部 tag 匹配的直接子元素; tag 为空串时返回全部.
func (inst *domElement) innerListChildNodes(tag string) []*domElement {
	if inst == nil {
		return nil
	}
	list := make([]*domElement, 0, len(inst.children))
	for _, child := range inst.children {
		if tag == "" || child.tag == tag {
			list = append(list, child)
		}
	}
	return list
}

// innerAppendNode 追加一个子元素(不检查 tag).
func (inst *domElement) innerAppendNode(node *domElement) {
	if inst == nil || node == nil {
		return
	}
	node.parent = inst
	innerSetOwnerRecursive(node, inst.owner)
	inst.children = append(inst.children, node)
}

// innerInsertNode 把 node 插入到第一个 rank 更大的兄弟之前; rank 为 nil 时追加到末尾.
func (inst *domElement) innerInsertNode(node *domElement, rank func(tag string) int) {
	if inst == nil || node == nil {
		return
	}
	if rank != nil {
		value := rank(node.tag)
		for index, old := range inst.children {
			if rank(old.tag) > value {
				node.parent = inst
				innerSetOwnerRecursive(node, inst.owner)
				inst.children = slices.Insert(inst.children, index, node)
				return
			}
		}
	}
	inst.innerAppendNode(node)
}

// innerRemoveChildrenByTag 删除全部 tag 匹配的直接子元素.
func (inst *domElement) innerRemoveChildrenByTag(tag string) {
	if inst == nil {
		return
	}
	keep := make([]*domElement, 0, len(inst.children))
	for _, child := range inst.children {
		if child.tag == tag {
			child.parent = nil
			continue
		}
		keep = append(keep, child)
	}
	inst.children = keep
}

////////////////////////////////////////////////////////////////////////////////
// 内部节点的工具函数

// innerSetOwnerRecursive 把 node 及其子孙的 owner 改成 owner.
func innerSetOwnerRecursive(node *domElement, owner step1core.DOMDocument) {
	if node == nil {
		return
	}
	node.owner = owner
	for _, child := range node.children {
		innerSetOwnerRecursive(child, owner)
	}
}

// innerUnwrapElement 把任意 DOMElement 还原成内部节点.
//
// 这里用类型分支而不是接口方法: typed 包装类型内嵌的是 *domElement,
// 提升出来的方法在接收者为 typed-nil 时会在取字段的那一步 panic,
// 而类型分支可以把 nil 一并处理掉.
func innerUnwrapElement(element step1core.DOMElement) (*domElement, bool) {

	switch item := element.(type) {
	case *domElement:
		if item == nil {
			return nil, false
		}
		return item, true
	case *domWorkElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domHeadElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domMetaElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domTimelineElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domTracksElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domTrackElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domSliceElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domWaveClipElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domMidiClipElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domProgressionElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domChordElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domDevicesElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domDeviceElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	case *domWireElement:
		if item == nil {
			return nil, false
		}
		return item.domElement, true
	}

	return nil, false
}

// innerCheckChild 校验一个待加入的子元素, 返回它的内部节点.
func innerCheckChild(child step1core.DOMElement, tag string) (*domElement, error) {
	node, ok := innerUnwrapElement(child)
	if !ok {
		return nil, fmt.Errorf("child element is nil or unsupported")
	}
	if tag != "" && node.tag != tag {
		return nil, fmt.Errorf("invalid child element: %s, want: %s", node.tag, tag)
	}
	if node.parent != nil {
		return nil, fmt.Errorf("child element already has a parent: %s", node.tag)
	}
	return node, nil
}

// innerAppendTypedChild 校验 tag 之后把子元素追加到末尾.
func innerAppendTypedChild(inst *domElement, child step1core.DOMElement, tag string) error {
	node, err := innerCheckChild(child, tag)
	if err != nil {
		return err
	}
	inst.innerAppendNode(node)
	return nil
}

// innerSetSingleChild 设置"至多一个"的子元素: child 为 nil 表示删除.
func innerSetSingleChild(inst *domElement, child step1core.DOMElement, tag string, rank func(tag string) int) error {
	if inst == nil {
		return fmt.Errorf("element is nil")
	}
	inst.innerRemoveChildrenByTag(tag)
	if child == nil {
		return nil
	}
	node, err := innerCheckChild(child, tag)
	if err != nil {
		return err
	}
	inst.innerInsertNode(node, rank)
	return nil
}

// innerChildAs 查找第一个 tag 匹配的直接子元素, 并转换成接口 T.
func innerChildAs[T any](inst *domElement, tag string) T {
	var zero T
	node := inst.innerFirstChild(tag)
	if node == nil {
		return zero
	}
	item, ok := innerWrapElement(node).(T)
	if !ok {
		return zero
	}
	return item
}

// innerChildrenAs 列出全部 tag 匹配的直接子元素, 并转换成接口 T.
func innerChildrenAs[T any](inst *domElement, tag string) []T {
	nodes := inst.innerListChildNodes(tag)
	list := make([]T, 0, len(nodes))
	for _, node := range nodes {
		item, ok := innerWrapElement(node).(T)
		if ok {
			list = append(list, item)
		}
	}
	return list
}

// innerWrapElements 把内部节点逐个包装成 DOMElement.
func innerWrapElements(nodes []*domElement) []step1core.DOMElement {
	list := make([]step1core.DOMElement, 0, len(nodes))
	for _, node := range nodes {
		list = append(list, innerWrapElement(node))
	}
	return list
}

// innerWrapElement 按标签名把内部节点包装成对应的 typed element.
//
// 只包装本文档命名空间(或没有命名空间)的元素; 其它命名空间的扩展元素一律
// 作为通用 DOMElement 返回. 同一个节点只会有一个包装对象, 会被缓存下来.
func innerWrapElement(node *domElement) step1core.DOMElement {
	if node == nil {
		return nil
	}
	if node.wrapper == nil {
		node.wrapper = innerNewWrapper(node)
	}
	return node.wrapper
}

// innerWrapAs 把内部节点包装成接口 T; 节点类型不匹配时返回错误.
//
// 新建元素时应当统一用它来取得包装对象 (而不是直接 &domXxxElement{node}),
// 这样节点的 wrapper 缓存里存的始终是同一个对象.
func innerWrapAs[T any](node *domElement) (T, error) {
	var zero T
	if node == nil {
		return zero, fmt.Errorf("element is nil")
	}
	item, ok := innerWrapElement(node).(T)
	if !ok {
		return zero, fmt.Errorf("unexpected element type, tag: %s", node.tag)
	}
	return item, nil
}

// innerNewWrapper 新建一个包装对象.
func innerNewWrapper(node *domElement) step1core.DOMElement {

	if node.namespace != "" && node.namespace != step1core.Step1WorkNamespace {
		return node
	}

	switch node.tag {
	case step1core.TagWork:
		return &domWorkElement{node}
	case step1core.TagHead:
		return &domHeadElement{node}
	case step1core.TagMeta:
		return &domMetaElement{node}
	case step1core.TagTimeline:
		return &domTimelineElement{node}
	case step1core.TagTracks:
		return &domTracksElement{node}
	case step1core.TagTrack:
		return &domTrackElement{node}
	case step1core.TagSlice:
		return &domSliceElement{node}
	case step1core.TagWave:
		return &domWaveClipElement{node}
	case step1core.TagMidi:
		return &domMidiClipElement{node}
	case step1core.TagProgression:
		return &domProgressionElement{node}
	case step1core.TagChord:
		return &domChordElement{node}
	case step1core.TagDevices:
		return &domDevicesElement{node}
	case step1core.TagDevice:
		return &domDeviceElement{node}
	case step1core.TagWire:
		return &domWireElement{node}
	default:
		return node
	}
}

// innerAttributeFullName 返回属性的完整名字 (prefix:name 或 name).
func innerAttributeFullName(attr *step1core.DOMAttribute) string {
	if attr == nil {
		return ""
	}
	if attr.Prefix != "" {
		return attr.Prefix + ":" + attr.Name
	}
	return attr.Name
}

// innerRankWork 给出 work 子元素在 xs:sequence 中的位置.
func innerRankWork(tag string) int {
	switch tag {
	case step1core.TagHead:
		return 0
	case step1core.TagTimeline:
		return 1
	case step1core.TagTracks:
		return 2
	case step1core.TagDevices:
		return 3
	default:
		return 4
	}
}

// innerRankHead 给出 head 子元素在 xs:sequence 中的位置.
func innerRankHead(tag string) int {
	switch tag {
	case step1core.TagTitle:
		return 0
	case step1core.TagMeta:
		return 1
	default:
		return 2
	}
}

// innerRankSliceChild 给出 slice 子元素在 xs:sequence 中的位置:
// choice(wave|midi|progression) 在扩展元素之前.
func innerRankSliceChild(tag string) int {
	switch tag {
	case step1core.TagWave, step1core.TagMidi, step1core.TagProgression:
		return 0
	default:
		return 1
	}
}

////////////////////////////////////////////////////////////////////////////////
// work : WorkElement

type domWorkElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.WorkElement = (*domWorkElement)(nil)

// GetOrderID implements [step1core.WorkElement].
func (inst *domWorkElement) GetOrderID() step1core.OrderID {
	return step1core.OrderID(inst.innerGetAttribute(step1core.AttrOrderID))
}

// SetOrderID implements [step1core.WorkElement].
func (inst *domWorkElement) SetOrderID(id step1core.OrderID) error {
	if id == "" {
		return inst.RemoveAttribute(step1core.AttrOrderID)
	}
	if err := step1core.CheckOrderIDText(string(id)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrOrderID, string(id))
}

// GetHead implements [step1core.WorkElement].
func (inst *domWorkElement) GetHead() step1core.HeadElement {
	return innerChildAs[step1core.HeadElement](inst.domElement, step1core.TagHead)
}

// SetHead implements [step1core.WorkElement].
func (inst *domWorkElement) SetHead(head step1core.HeadElement) error {
	return innerSetSingleChild(inst.domElement, head, step1core.TagHead, innerRankWork)
}

// ListTimelines implements [step1core.WorkElement].
func (inst *domWorkElement) ListTimelines() []step1core.TimelineElement {
	return innerChildrenAs[step1core.TimelineElement](inst.domElement, step1core.TagTimeline)
}

// AddTimeline implements [step1core.WorkElement].
func (inst *domWorkElement) AddTimeline(item step1core.TimelineElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagTimeline)
}

// GetTracks implements [step1core.WorkElement].
func (inst *domWorkElement) GetTracks() step1core.TracksElement {
	return innerChildAs[step1core.TracksElement](inst.domElement, step1core.TagTracks)
}

// SetTracks implements [step1core.WorkElement].
func (inst *domWorkElement) SetTracks(tracks step1core.TracksElement) error {
	return innerSetSingleChild(inst.domElement, tracks, step1core.TagTracks, innerRankWork)
}

// GetDevices implements [step1core.WorkElement].
func (inst *domWorkElement) GetDevices() step1core.DevicesElement {
	return innerChildAs[step1core.DevicesElement](inst.domElement, step1core.TagDevices)
}

// SetDevices implements [step1core.WorkElement].
func (inst *domWorkElement) SetDevices(devices step1core.DevicesElement) error {
	return innerSetSingleChild(inst.domElement, devices, step1core.TagDevices, innerRankWork)
}

// FindTimeline implements [step1core.WorkElement].
//
// timeline/@name 在整篇文档中唯一, 因此会递归查找各级子时间轴.
func (inst *domWorkElement) FindTimeline(name step1core.NameType) (step1core.TimelineElement, error) {

	var find func(items []step1core.TimelineElement) step1core.TimelineElement
	find = func(items []step1core.TimelineElement) step1core.TimelineElement {
		for _, item := range items {
			if item == nil {
				continue
			}
			if item.GetName() == name {
				return item
			}
			if got := find(item.ListTimelines()); got != nil {
				return got
			}
		}
		return nil
	}

	got := find(inst.ListTimelines())
	if got == nil {
		return nil, fmt.Errorf("timeline not found: %s", name)
	}
	return got, nil
}

// FindTrack implements [step1core.WorkElement].
func (inst *domWorkElement) FindTrack(name step1core.NameType) (step1core.TrackElement, error) {

	tracks := inst.GetTracks()
	if tracks != nil {
		for _, item := range tracks.ListTracks() {
			if item != nil && item.GetName() == name {
				return item, nil
			}
		}
	}

	return nil, fmt.Errorf("track not found: %s", name)
}

// FindDevice implements [step1core.WorkElement].
func (inst *domWorkElement) FindDevice(name step1core.DeviceName) (step1core.DeviceElement, error) {

	devices := inst.GetDevices()
	if devices != nil {
		for _, item := range devices.ListDevices() {
			if item != nil && item.GetName() == name {
				return item, nil
			}
		}
	}

	return nil, fmt.Errorf("device not found: %s", name)
}

// FindDeviceByID implements [step1core.WorkElement].
func (inst *domWorkElement) FindDeviceByID(id step1core.DeviceID) (step1core.DeviceElement, error) {

	if id == "" {
		return nil, fmt.Errorf("device id is empty")
	}

	devices := inst.GetDevices()
	if devices != nil {
		for _, item := range devices.ListDevices() {
			if item != nil && item.GetID() == id {
				return item, nil
			}
		}
	}

	return nil, fmt.Errorf("device not found, id: %s", id)
}

////////////////////////////////////////////////////////////////////////////////
// head : HeadElement

type domHeadElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.HeadElement = (*domHeadElement)(nil)

// GetTitle implements [step1core.HeadElement].
func (inst *domHeadElement) GetTitle() string {
	node := inst.innerFirstChild(step1core.TagTitle)
	if node == nil {
		return ""
	}
	return node.text
}

// SetTitle implements [step1core.HeadElement].
func (inst *domHeadElement) SetTitle(title string) error {

	if inst == nil || inst.domElement == nil {
		return fmt.Errorf("element is nil")
	}

	node := inst.innerFirstChild(step1core.TagTitle)
	if node == nil {
		// 按 xs:sequence, title 是 head 的第一个子元素
		node = innerNewElement(inst.owner, inst.namespace, inst.prefix, step1core.TagTitle)
		inst.innerInsertNode(node, innerRankHead)
	}

	node.text = title
	return nil
}

// ListMeta implements [step1core.HeadElement].
func (inst *domHeadElement) ListMeta() []step1core.MetaElement {
	return innerChildrenAs[step1core.MetaElement](inst.domElement, step1core.TagMeta)
}

// AddMeta implements [step1core.HeadElement].
func (inst *domHeadElement) AddMeta(item step1core.MetaElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagMeta)
}

// FindMeta implements [step1core.HeadElement].
func (inst *domHeadElement) FindMeta(name step1core.MetaName) (step1core.MetaElement, error) {

	for _, item := range inst.ListMeta() {
		if item != nil && item.GetName() == name {
			return item, nil
		}
	}

	return nil, fmt.Errorf("meta not found: %s", name)
}

////////////////////////////////////////////////////////////////////////////////
// meta : MetaElement

type domMetaElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.MetaElement = (*domMetaElement)(nil)

// GetName implements [step1core.MetaElement].
func (inst *domMetaElement) GetName() step1core.MetaName {
	return step1core.MetaName(inst.innerGetAttribute(step1core.AttrName))
}

// SetName implements [step1core.MetaElement].
func (inst *domMetaElement) SetName(name step1core.MetaName) error {
	if err := step1core.CheckMetaNameText(string(name)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrName, string(name))
}

// GetValue implements [step1core.MetaElement].
func (inst *domMetaElement) GetValue() string {
	return inst.innerGetAttribute(step1core.AttrValue)
}

// SetValue implements [step1core.MetaElement].
func (inst *domMetaElement) SetValue(value string) error {
	return inst.SetAttribute(step1core.AttrValue, value)
}

////////////////////////////////////////////////////////////////////////////////
// timeline : TimelineElement

type domTimelineElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.TimelineElement = (*domTimelineElement)(nil)

// GetName implements [step1core.TimelineElement].
func (inst *domTimelineElement) GetName() step1core.NameType {
	return step1core.NameType(inst.innerGetAttribute(step1core.AttrName))
}

// SetName implements [step1core.TimelineElement].
func (inst *domTimelineElement) SetName(name step1core.NameType) error {
	if err := step1core.CheckNCName(string(name)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrName, string(name))
}

// GetBPM implements [step1core.TimelineElement].
func (inst *domTimelineElement) GetBPM() *step1core.BPM {
	text, ok := inst.GetAttribute(step1core.AttrBPM)
	if !ok {
		return nil
	}
	value, err := step1core.ParseBPMText(text)
	if err != nil {
		return nil
	}
	return &value
}

// SetBPM implements [step1core.TimelineElement].
func (inst *domTimelineElement) SetBPM(value *step1core.BPM) error {
	if value == nil {
		return inst.RemoveAttribute(step1core.AttrBPM)
	}
	if *value < step1core.BPMRangeMin || *value > step1core.BPMRangeMax {
		return fmt.Errorf("bpm out of range [%v,%v]: %v", step1core.BPMRangeMin, step1core.BPMRangeMax, *value)
	}
	return inst.SetAttribute(step1core.AttrBPM, step1core.FormatBPMText(*value))
}

// GetTempo implements [step1core.TimelineElement].
func (inst *domTimelineElement) GetTempo() *step1core.Tempo {
	text, ok := inst.GetAttribute(step1core.AttrTempo)
	if !ok {
		return nil
	}
	value, err := step1core.ParseTempoText(text)
	if err != nil {
		return nil
	}
	return &value
}

// SetTempo implements [step1core.TimelineElement].
func (inst *domTimelineElement) SetTempo(value *step1core.Tempo) error {
	if value == nil {
		return inst.RemoveAttribute(step1core.AttrTempo)
	}
	if value.Upper < 1 || value.Lower < 1 {
		return fmt.Errorf("invalid tempo: %d/%d", value.Upper, value.Lower)
	}
	return inst.SetAttribute(step1core.AttrTempo, step1core.FormatTempoText(*value))
}

// GetOffset implements [step1core.TimelineElement].
func (inst *domTimelineElement) GetOffset() *step1core.SampleTime {
	return innerGetOptionalSampleTime(inst.domElement, step1core.AttrOffset)
}

// SetOffset implements [step1core.TimelineElement].
func (inst *domTimelineElement) SetOffset(value *step1core.SampleTime) error {
	return innerSetOptionalSampleTime(inst.domElement, step1core.AttrOffset, value)
}

// GetLength implements [step1core.TimelineElement].
func (inst *domTimelineElement) GetLength() *step1core.SampleTime {
	return innerGetOptionalSampleTime(inst.domElement, step1core.AttrLength)
}

// SetLength implements [step1core.TimelineElement].
func (inst *domTimelineElement) SetLength(value *step1core.SampleTime) error {
	return innerSetOptionalSampleTime(inst.domElement, step1core.AttrLength, value)
}

// ListTimelines implements [step1core.TimelineElement].
func (inst *domTimelineElement) ListTimelines() []step1core.TimelineElement {
	return innerChildrenAs[step1core.TimelineElement](inst.domElement, step1core.TagTimeline)
}

// AddTimeline implements [step1core.TimelineElement].
func (inst *domTimelineElement) AddTimeline(item step1core.TimelineElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagTimeline)
}

////////////////////////////////////////////////////////////////////////////////
// tracks : TracksElement

type domTracksElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.TracksElement = (*domTracksElement)(nil)

// ListTracks implements [step1core.TracksElement].
func (inst *domTracksElement) ListTracks() []step1core.TrackElement {
	return innerChildrenAs[step1core.TrackElement](inst.domElement, step1core.TagTrack)
}

// AddTrack implements [step1core.TracksElement].
func (inst *domTracksElement) AddTrack(item step1core.TrackElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagTrack)
}

////////////////////////////////////////////////////////////////////////////////
// track : TrackElement

type domTrackElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.TrackElement = (*domTrackElement)(nil)

// GetName implements [step1core.TrackElement].
func (inst *domTrackElement) GetName() step1core.NameType {
	return step1core.NameType(inst.innerGetAttribute(step1core.AttrName))
}

// SetName implements [step1core.TrackElement].
func (inst *domTrackElement) SetName(name step1core.NameType) error {
	if err := step1core.CheckNCName(string(name)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrName, string(name))
}

// GetType implements [step1core.TrackElement].
func (inst *domTrackElement) GetType() step1core.TrackType {
	text := inst.innerGetAttribute(step1core.AttrType)
	value, err := step1core.ParseTrackTypeText(text)
	if err != nil {
		return step1core.TrackTypeUnknown
	}
	return value
}

// SetType implements [step1core.TrackElement].
func (inst *domTrackElement) SetType(t step1core.TrackType) error {
	text := step1core.FormatTrackTypeText(t)
	if text == "" {
		return fmt.Errorf("invalid track type: %d", int(t))
	}
	return inst.SetAttribute(step1core.AttrType, text)
}

// ListSlices implements [step1core.TrackElement].
func (inst *domTrackElement) ListSlices() []step1core.SliceElement {
	return innerChildrenAs[step1core.SliceElement](inst.domElement, step1core.TagSlice)
}

// AddSlice implements [step1core.TrackElement].
func (inst *domTrackElement) AddSlice(item step1core.SliceElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagSlice)
}

////////////////////////////////////////////////////////////////////////////////
// slice : SliceElement

type domSliceElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.SliceElement = (*domSliceElement)(nil)

// GetTimeline implements [step1core.SliceElement].
func (inst *domSliceElement) GetTimeline() step1core.NameType {
	return step1core.NameType(inst.innerGetAttribute(step1core.AttrTimeline))
}

// SetTimeline implements [step1core.SliceElement].
func (inst *domSliceElement) SetTimeline(name step1core.NameType) error {
	if err := step1core.CheckNCName(string(name)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrTimeline, string(name))
}

// GetType implements [step1core.SliceElement].
func (inst *domSliceElement) GetType() *step1core.TrackType {
	text, ok := inst.GetAttribute(step1core.AttrType)
	if !ok {
		return nil
	}
	value, err := step1core.ParseTrackTypeText(text)
	if err != nil {
		return nil
	}
	if err := step1core.CheckSliceType(value); err != nil {
		return nil
	}
	return &value
}

// SetType implements [step1core.SliceElement].
func (inst *domSliceElement) SetType(value *step1core.TrackType) error {
	if value == nil {
		return inst.RemoveAttribute(step1core.AttrType)
	}
	if err := step1core.CheckSliceType(*value); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrType, step1core.FormatTrackTypeText(*value))
}

// GetWaveClip implements [step1core.SliceElement].
func (inst *domSliceElement) GetWaveClip() step1core.WaveClipElement {
	return innerChildAs[step1core.WaveClipElement](inst.domElement, step1core.TagWave)
}

// SetWaveClip implements [step1core.SliceElement].
func (inst *domSliceElement) SetWaveClip(item step1core.WaveClipElement) error {
	return inst.innerSetContent(item, step1core.TagWave)
}

// GetMidiClip implements [step1core.SliceElement].
func (inst *domSliceElement) GetMidiClip() step1core.MidiClipElement {
	return innerChildAs[step1core.MidiClipElement](inst.domElement, step1core.TagMidi)
}

// SetMidiClip implements [step1core.SliceElement].
func (inst *domSliceElement) SetMidiClip(item step1core.MidiClipElement) error {
	return inst.innerSetContent(item, step1core.TagMidi)
}

// GetProgression implements [step1core.SliceElement].
func (inst *domSliceElement) GetProgression() step1core.ProgressionElement {
	return innerChildAs[step1core.ProgressionElement](inst.domElement, step1core.TagProgression)
}

// SetProgression implements [step1core.SliceElement].
func (inst *domSliceElement) SetProgression(item step1core.ProgressionElement) error {
	return inst.innerSetContent(item, step1core.TagProgression)
}

// innerSetContent 设置 xs:choice 中的内容元素: 三者互斥, nil 表示清空.
func (inst *domSliceElement) innerSetContent(child step1core.DOMElement, tag string) error {

	if inst == nil || inst.domElement == nil {
		return fmt.Errorf("element is nil")
	}

	for _, item := range []string{step1core.TagWave, step1core.TagMidi, step1core.TagProgression} {
		inst.innerRemoveChildrenByTag(item)
	}

	if child == nil {
		return nil
	}

	node, err := innerCheckChild(child, tag)
	if err != nil {
		return err
	}

	inst.innerInsertNode(node, innerRankSliceChild)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// wave / midi : 片段 (ExtensionPointType)

type domWaveClipElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.WaveClipElement = (*domWaveClipElement)(nil)

// GetSrc implements [step1core.WaveClipElement].
func (inst *domWaveClipElement) GetSrc() string {
	return inst.innerGetAttribute(step1core.AttrSrc)
}

// SetSrc implements [step1core.WaveClipElement].
func (inst *domWaveClipElement) SetSrc(src string) error {
	return inst.SetAttribute(step1core.AttrSrc, src)
}

type domMidiClipElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.MidiClipElement = (*domMidiClipElement)(nil)

// GetSrc implements [step1core.MidiClipElement].
func (inst *domMidiClipElement) GetSrc() string {
	return inst.innerGetAttribute(step1core.AttrSrc)
}

// SetSrc implements [step1core.MidiClipElement].
func (inst *domMidiClipElement) SetSrc(src string) error {
	return inst.SetAttribute(step1core.AttrSrc, src)
}

////////////////////////////////////////////////////////////////////////////////
// progression : ProgressionElement

type domProgressionElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.ProgressionElement = (*domProgressionElement)(nil)

// ListChords implements [step1core.ProgressionElement].
func (inst *domProgressionElement) ListChords() []step1core.ChordElement {
	return innerChildrenAs[step1core.ChordElement](inst.domElement, step1core.TagChord)
}

// AddChord implements [step1core.ProgressionElement].
func (inst *domProgressionElement) AddChord(item step1core.ChordElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagChord)
}

////////////////////////////////////////////////////////////////////////////////
// chord : ChordElement (simpleContent)

type domChordElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.ChordElement = (*domChordElement)(nil)

// GetRoot implements [step1core.ChordElement].
//
// 属性值里若带有八度 (xsd 允许, 例如 "C4") 则无法用 ToneName 表示, 此时返回空串.
func (inst *domChordElement) GetRoot() step1core.ToneName {
	name, err := step1core.ParseToneNameText(inst.innerGetAttribute(step1core.AttrRoot))
	if err != nil {
		return ""
	}
	return name
}

// SetRoot implements [step1core.ChordElement].
func (inst *domChordElement) SetRoot(root step1core.ToneName) error {
	name, err := step1core.ParseToneNameText(string(root))
	if err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrRoot, string(name))
}

// GetOffset implements [step1core.ChordElement].
func (inst *domChordElement) GetOffset() *step1core.SampleTime {
	return innerGetOptionalSampleTime(inst.domElement, step1core.AttrOffset)
}

// SetOffset implements [step1core.ChordElement].
func (inst *domChordElement) SetOffset(value *step1core.SampleTime) error {
	return innerSetOptionalSampleTime(inst.domElement, step1core.AttrOffset, value)
}

// GetNotes implements [step1core.ChordElement].
//
// 组成音取自元素文本, 解析失败时返回 nil.
func (inst *domChordElement) GetNotes() []step1core.ToneName {
	notes, err := step1core.ParseChordNotesText(inst.GetText())
	if err != nil {
		return nil
	}
	return notes
}

// SetNotes implements [step1core.ChordElement].
func (inst *domChordElement) SetNotes(notes []step1core.ToneName) error {
	text := step1core.FormatChordNotesText(notes)
	if err := step1core.CheckChordNotesText(text); err != nil {
		return err
	}
	return inst.SetText(text)
}

////////////////////////////////////////////////////////////////////////////////
// devices : DevicesElement

type domDevicesElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.DevicesElement = (*domDevicesElement)(nil)

// ListDevices implements [step1core.DevicesElement].
func (inst *domDevicesElement) ListDevices() []step1core.DeviceElement {
	return innerChildrenAs[step1core.DeviceElement](inst.domElement, step1core.TagDevice)
}

// AddDevice implements [step1core.DevicesElement].
func (inst *domDevicesElement) AddDevice(item step1core.DeviceElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagDevice)
}

////////////////////////////////////////////////////////////////////////////////
// device : DeviceElement

type domDeviceElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.DeviceElement = (*domDeviceElement)(nil)

// GetName implements [step1core.DeviceElement].
func (inst *domDeviceElement) GetName() step1core.DeviceName {
	return step1core.DeviceName(inst.innerGetAttribute(step1core.AttrName))
}

// SetName implements [step1core.DeviceElement].
func (inst *domDeviceElement) SetName(name step1core.DeviceName) error {
	if err := step1core.CheckNCName(string(name)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrName, string(name))
}

// GetClass implements [step1core.DeviceElement].
func (inst *domDeviceElement) GetClass() step1core.DeviceClass {
	text := inst.innerGetAttribute(step1core.AttrClass)
	value, err := step1core.ParseDeviceClassText(text)
	if err != nil {
		return step1core.DeviceClass(text)
	}
	return value
}

// SetClass implements [step1core.DeviceElement].
func (inst *domDeviceElement) SetClass(class step1core.DeviceClass) error {
	value, err := step1core.ParseDeviceClassText(string(class))
	if err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrClass, step1core.FormatDeviceClassText(value))
}

// GetID implements [step1core.DeviceElement].
func (inst *domDeviceElement) GetID() step1core.DeviceID {
	return step1core.DeviceID(inst.innerGetAttribute(step1core.AttrID))
}

// SetID implements [step1core.DeviceElement].
func (inst *domDeviceElement) SetID(id step1core.DeviceID) error {
	if id == "" {
		return inst.RemoveAttribute(step1core.AttrID)
	}
	if err := step1core.CheckNCName(string(id)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrID, string(id))
}

// ListWires implements [step1core.DeviceElement].
func (inst *domDeviceElement) ListWires() []step1core.WireElement {
	return innerChildrenAs[step1core.WireElement](inst.domElement, step1core.TagWire)
}

// AddWire implements [step1core.DeviceElement].
func (inst *domDeviceElement) AddWire(item step1core.WireElement) error {
	return innerAppendTypedChild(inst.domElement, item, step1core.TagWire)
}

////////////////////////////////////////////////////////////////////////////////
// wire : WireElement (ExtensionPointType)

type domWireElement struct {
	*domElement
}

// 确保实现相关接口
var _ step1core.WireElement = (*domWireElement)(nil)

// GetPort implements [step1core.WireElement].
func (inst *domWireElement) GetPort() step1core.PortName {
	return step1core.PortName(inst.innerGetAttribute(step1core.AttrPort))
}

// SetPort implements [step1core.WireElement].
func (inst *domWireElement) SetPort(port step1core.PortName) error {
	if err := step1core.CheckNCName(string(port)); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrPort, string(port))
}

// GetSource implements [step1core.WireElement].
//
// 解析失败(属性缺失或格式不对)时返回零值.
func (inst *domWireElement) GetSource() step1core.PortRef {
	ref, err := step1core.ParseSourceRefText(inst.innerGetAttribute(step1core.AttrSource))
	if err != nil {
		return step1core.PortRef{}
	}
	return ref
}

// SetSource implements [step1core.WireElement].
func (inst *domWireElement) SetSource(source step1core.PortRef) error {
	text := step1core.FormatSourceRefText(source)
	if _, err := step1core.ParseSourceRefText(text); err != nil {
		return err
	}
	return inst.SetAttribute(step1core.AttrSource, text)
}

////////////////////////////////////////////////////////////////////////////////
// 可选的 SampleTime 属性 (offset / length)

// innerGetOptionalSampleTime 读取一个可选的 SampleTime 属性, 缺失或非法时返回 nil.
func innerGetOptionalSampleTime(inst *domElement, name string) *step1core.SampleTime {
	text, ok := inst.GetAttribute(name)
	if !ok {
		return nil
	}
	value, err := step1core.ParseSampleTimeText(text)
	if err != nil {
		return nil
	}
	return &value
}

// innerSetOptionalSampleTime 写入一个可选的 SampleTime 属性, nil 表示删除.
func innerSetOptionalSampleTime(inst *domElement, name string, value *step1core.SampleTime) error {
	if value == nil {
		return inst.RemoveAttribute(name)
	}
	if *value < 0 {
		return fmt.Errorf("sample time must not be negative: %d", int64(*value))
	}
	return inst.SetAttribute(name, step1core.FormatSampleTimeText(*value))
}

////////////////////////////////////////////////////////////////////////////////
// 文档 : innerDOMDocument

type innerDOMDocument struct {
	impl *innerDOMImplementation

	url string

	root *domElement
}

// 确保实现相关接口
var _ step1core.DOMDocument = (*innerDOMDocument)(nil)

// GetOwner implements [step1core.DOMNode].
//
// 文档节点的 owner 就是它自己.
func (inst *innerDOMDocument) GetOwner() step1core.DOMDocument {
	if inst == nil {
		return nil
	}
	return inst
}

// Implementation implements [step1core.DOMDocument].
func (inst *innerDOMDocument) Implementation() step1core.DOMImplementation {
	if inst == nil {
		return nil
	}
	return inst.impl
}

// GetURL implements [step1core.DOMDocument].
func (inst *innerDOMDocument) GetURL() string {
	if inst == nil {
		return ""
	}
	return inst.url
}

// GetRoot implements [step1core.DOMDocument].
func (inst *innerDOMDocument) GetRoot() (step1core.WorkElement, error) {

	if inst == nil || inst.root == nil {
		return nil, fmt.Errorf("work document has no root element")
	}
	if inst.root.tag != step1core.TagWork {
		return nil, fmt.Errorf("invalid root element: %s", inst.root.tag)
	}

	return innerWrapAs[step1core.WorkElement](inst.root)
}

// SetRoot implements [step1core.DOMDocument].
func (inst *innerDOMDocument) SetRoot(root step1core.WorkElement) error {

	if inst == nil {
		return fmt.Errorf("document is nil")
	}

	node, err := innerCheckChild(root, step1core.TagWork)
	if err != nil {
		return err
	}

	node.parent = nil
	innerSetOwnerRecursive(node, inst)
	inst.root = node
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 工厂 : innerDOMFactory

type innerDOMFactory struct {
	impl *innerDOMImplementation
}

// 确保实现相关接口
var _ step1core.DOMFactory = (*innerDOMFactory)(nil)

// CreateDocument implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateDocument(url string) step1core.DOMDocument {
	doc := new(innerDOMDocument)
	doc.impl = inst.impl
	doc.url = url
	return doc
}

// CreateElement implements [step1core.DOMFactory].
//
// 用于创建 title 以及其它命名空间的扩展元素; namespace / prefix 都允许为空串.
func (inst *innerDOMFactory) CreateElement(doc step1core.DOMDocument, namespace string, prefix string, tag string) (step1core.DOMElement, error) {
	if tag == "" {
		return nil, fmt.Errorf("element tag is empty")
	}
	return innerNewElement(doc, namespace, prefix, tag), nil
}

// CreateWorkElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateWorkElement(doc step1core.DOMDocument) (step1core.WorkElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagWork)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.WorkElement](node)
}

// CreateHeadElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateHeadElement(doc step1core.DOMDocument) (step1core.HeadElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagHead)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.HeadElement](node)
}

// CreateMetaElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateMetaElement(doc step1core.DOMDocument) (step1core.MetaElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagMeta)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.MetaElement](node)
}

// CreateTimelineElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateTimelineElement(doc step1core.DOMDocument) (step1core.TimelineElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagTimeline)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.TimelineElement](node)
}

// CreateTracksElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateTracksElement(doc step1core.DOMDocument) (step1core.TracksElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagTracks)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.TracksElement](node)
}

// CreateTrackElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateTrackElement(doc step1core.DOMDocument) (step1core.TrackElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagTrack)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.TrackElement](node)
}

// CreateSliceElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateSliceElement(doc step1core.DOMDocument) (step1core.SliceElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagSlice)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.SliceElement](node)
}

// CreateWaveClipElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateWaveClipElement(doc step1core.DOMDocument) (step1core.WaveClipElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagWave)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.WaveClipElement](node)
}

// CreateMidiClipElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateMidiClipElement(doc step1core.DOMDocument) (step1core.MidiClipElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagMidi)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.MidiClipElement](node)
}

// CreateProgressionElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateProgressionElement(doc step1core.DOMDocument) (step1core.ProgressionElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagProgression)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.ProgressionElement](node)
}

// CreateChordElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateChordElement(doc step1core.DOMDocument) (step1core.ChordElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagChord)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.ChordElement](node)
}

// CreateDevicesElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateDevicesElement(doc step1core.DOMDocument) (step1core.DevicesElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagDevices)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.DevicesElement](node)
}

// CreateDeviceElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateDeviceElement(doc step1core.DOMDocument) (step1core.DeviceElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagDevice)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.DeviceElement](node)
}

// CreateWireElement implements [step1core.DOMFactory].
func (inst *innerDOMFactory) CreateWireElement(doc step1core.DOMDocument) (step1core.WireElement, error) {
	node, err := inst.innerCreate(doc, step1core.TagWire)
	if err != nil {
		return nil, err
	}
	return innerWrapAs[step1core.WireElement](node)
}

// innerCreate 在本文档的命名空间里创建一个元素.
func (inst *innerDOMFactory) innerCreate(doc step1core.DOMDocument, tag string) (*domElement, error) {
	if doc == nil {
		return nil, fmt.Errorf("owner document is nil")
	}
	return innerNewElement(doc, step1core.Step1WorkNamespace, "", tag), nil
}

////////////////////////////////////////////////////////////////////////////////
// 编解码器 : innerDOMDocumentCodec

type innerDOMDocumentCodec struct {
	impl *innerDOMImplementation
}

// 确保实现相关接口
var _ step1core.DOMDocumentCodec = (*innerDOMDocumentCodec)(nil)

// Encode implements [step1core.DOMDocumentCodec].
//
// 输出带 XML 声明的格式化文本(4 个空格缩进); 没有子元素的元素写成单行,
// 没有文本也没有子元素时写成自闭合标签.
//
// 命名空间声明: 元素自身的命名空间若尚未在其祖先上声明, 会自动补上
// xmlns / xmlns:prefix, 因此由工厂新建的文档也能得到合法的 XML.
func (inst *innerDOMDocumentCodec) Encode(doc step1core.DOMDocument) ([]byte, error) {

	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	root, err := doc.GetRoot()
	if err != nil {
		return nil, err
	}

	node, ok := innerUnwrapElement(root)
	if !ok {
		return nil, fmt.Errorf("unsupported DOMDocument implementation")
	}

	buf := new(bytes.Buffer)
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteByte('\n')

	scope := make(map[string]string) // namespace -> prefix
	if err := innerEncodeElement(buf, node, scope, 0); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// Decode implements [step1core.DOMDocumentCodec].
//
// 注释 / 处理指令 / DOCTYPE 会被忽略; 根元素必须是 work.
func (inst *innerDOMDocumentCodec) Decode(data []byte, doc step1core.DOMDocument) error {

	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	uri2prefix := make(map[string]string)
	stack := make([]*domElement, 0, 8)
	nsStack := make([][]string, 0, 8) // 每层新增的 namespace, 用于退出时恢复

	var root *domElement

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid xml: %w", err)
		}

		switch item := token.(type) {

		case xml.StartElement:

			node := innerNewElement(doc, "", "", item.Name.Local)

			// 1. 先登记本层的命名空间声明, 后面的名字解析要用到
			added := make([]string, 0, 2)
			for _, attr := range item.Attr {
				prefix, uri, ok := innerSplitNSDeclaration(attr)
				if !ok {
					continue
				}
				if _, exists := uri2prefix[uri]; !exists {
					uri2prefix[uri] = prefix
					added = append(added, uri)
				}
			}

			node.namespace = item.Name.Space
			node.prefix = innerPrefixOf(uri2prefix, item.Name.Space)

			// 2. 加入父元素
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				node.parent = parent
				parent.children = append(parent.children, node)
			} else if root == nil {
				root = node
			} else {
				return fmt.Errorf("multiple root elements")
			}

			// 3. 属性
			for _, attr := range item.Attr {
				if out, ok := innerMakeNSDeclaration(attr); ok {
					node.attributes = append(node.attributes, out)
					continue
				}
				out := step1core.DOMAttribute{Value: attr.Value}
				out.Name = attr.Name.Local
				out.Prefix = innerPrefixOf(uri2prefix, attr.Name.Space)
				if innerIsNamespaceURI(attr.Name.Space) {
					out.Namespace = attr.Name.Space
				}
				node.attributes = append(node.attributes, out)
			}

			stack = append(stack, node)
			nsStack = append(nsStack, added)

		case xml.EndElement:

			if len(stack) == 0 {
				return fmt.Errorf("unexpected end element: %s", item.Name.Local)
			}

			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// 退出本层: 撤销本层新增的命名空间声明
			if len(nsStack) > 0 {
				added := nsStack[len(nsStack)-1]
				nsStack = nsStack[:len(nsStack)-1]
				for _, uri := range added {
					delete(uri2prefix, uri)
				}
			}

			// 有子元素时, 文本只是格式化用的空白, 丢弃
			if len(node.children) > 0 {
				node.text = ""
			} else {
				node.text = strings.TrimSpace(node.text)
			}

		case xml.CharData:

			if len(stack) > 0 {
				node := stack[len(stack)-1]
				node.text += string(item)
			}

		default:
			// 注释 / 处理指令 / directive : 忽略
		}
	}

	if len(stack) != 0 {
		return fmt.Errorf("unexpected end of xml document")
	}
	if root == nil {
		return fmt.Errorf("no root element")
	}
	if root.tag != step1core.TagWork {
		return fmt.Errorf("invalid root element: %s", root.tag)
	}

	root.parent = nil
	work, err := innerWrapAs[step1core.WorkElement](root)
	if err != nil {
		return err
	}
	return doc.SetRoot(work)
}

////////////////////////////////////////////////////////////////////////////////
// 编解码器的内部函数

// innerEncodeElement 写出一个元素 (含缩进).
func innerEncodeElement(buf *bytes.Buffer, node *domElement, scope map[string]string, level int) error {

	if node == nil {
		return fmt.Errorf("element is nil")
	}

	indent := strings.Repeat("    ", level)
	name := innerQualifiedName(node.prefix, node.tag)

	// 1. 登记元素上已经写明的 xmlns 声明
	added := make([]string, 0, 2)
	for index := range node.attributes {
		attr := &node.attributes[index]
		if !innerIsNSDeclarationAttr(attr) {
			continue
		}
		if _, exists := scope[attr.Value]; !exists {
			scope[attr.Value] = innerNSDeclarationPrefix(attr)
			added = append(added, attr.Value)
		}
	}

	// 2. 收集需要补写的命名空间声明
	type nsDecl struct {
		prefix string
		uri    string
	}
	generated := make([]nsDecl, 0, 2)

	// 2.1 元素自身的命名空间
	if node.namespace != "" {
		if _, exists := scope[node.namespace]; !exists {
			scope[node.namespace] = node.prefix
			added = append(added, node.namespace)
			generated = append(generated, nsDecl{node.prefix, node.namespace})
		}
	}

	// 2.2 属性用到的前缀 (只补 xml / xsi 这类众所周知的命名空间)
	for index := range node.attributes {
		attr := &node.attributes[index]
		if attr.Prefix == "" || innerIsNSDeclarationAttr(attr) {
			continue
		}
		if innerScopeHasPrefix(scope, attr.Prefix) {
			continue
		}
		uri := innerWellKnownNamespace(attr.Prefix)
		if uri == "" {
			continue
		}
		scope[uri] = attr.Prefix
		added = append(added, uri)
		generated = append(generated, nsDecl{attr.Prefix, uri})
	}

	// 3. 开始标签 + 属性
	buf.WriteString(indent)
	buf.WriteByte('<')
	buf.WriteString(name)

	for _, decl := range generated {
		innerWriteNSDeclaration(buf, decl.prefix, decl.uri)
	}
	for index := range node.attributes {
		innerWriteAttribute(buf, &node.attributes[index])
	}

	// 4. 内容
	if len(node.children) == 0 {
		if node.text == "" {
			buf.WriteString("/>")
		} else {
			buf.WriteByte('>')
			buf.WriteString(innerEscapeText(node.text))
			buf.WriteString("</")
			buf.WriteString(name)
			buf.WriteByte('>')
		}
	} else {
		buf.WriteByte('>')
		buf.WriteByte('\n')
		for _, child := range node.children {
			if err := innerEncodeElement(buf, child, scope, level+1); err != nil {
				return err
			}
			buf.WriteByte('\n')
		}
		buf.WriteString(indent)
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteByte('>')
	}

	// 5. 退出本层, 撤销本层新增的声明
	for _, uri := range added {
		delete(scope, uri)
	}

	return nil
}

// innerQualifiedName 拼出 prefix:tag 或 tag.
func innerQualifiedName(prefix string, tag string) string {
	if prefix == "" {
		return tag
	}
	return prefix + ":" + tag
}

// innerWriteAttribute 写出一个属性.
func innerWriteAttribute(buf *bytes.Buffer, attr *step1core.DOMAttribute) {
	if attr == nil {
		return
	}
	buf.WriteByte(' ')
	buf.WriteString(innerAttributeFullName(attr))
	buf.WriteString(`="`)
	buf.WriteString(innerEscapeAttribute(attr.Value))
	buf.WriteByte('"')
}

// innerWriteNSDeclaration 写出一个命名空间声明.
func innerWriteNSDeclaration(buf *bytes.Buffer, prefix string, namespace string) {
	buf.WriteByte(' ')
	buf.WriteString("xmlns")
	if prefix != "" {
		buf.WriteByte(':')
		buf.WriteString(prefix)
	}
	buf.WriteString(`="`)
	buf.WriteString(innerEscapeAttribute(namespace))
	buf.WriteByte('"')
}

// innerIsNSDeclarationAttr 判断一个属性是不是 xmlns 声明.
func innerIsNSDeclarationAttr(attr *step1core.DOMAttribute) bool {
	if attr == nil {
		return false
	}
	if attr.Name == "xmlns" && attr.Prefix == "" {
		return true
	}
	return attr.Prefix == "xmlns"
}

// innerScopeHasPrefix 判断某个前缀是否已经在命名空间作用域里声明过.
func innerScopeHasPrefix(scope map[string]string, prefix string) bool {
	for _, old := range scope {
		if old == prefix {
			return true
		}
	}
	return false
}

// 众所周知的前缀 (xml 与 xsi), 见 XML Namespaces / XML Schema 规范.
const (
	innerPrefixXMLNamespace = "xml"
	innerPrefixXSINamespace = "xsi"

	innerURIXMLNamespace = "http://www.w3.org/XML/1998/namespace"
	innerURIXSINamespace = "http://www.w3.org/2001/XMLSchema-instance"
)

// innerWellKnownNamespace 返回众所周知的前缀对应的命名空间 URI; 未知时返回空串.
func innerWellKnownNamespace(prefix string) string {
	switch prefix {
	case innerPrefixXMLNamespace:
		return innerURIXMLNamespace
	case innerPrefixXSINamespace:
		return innerURIXSINamespace
	default:
		return ""
	}
}

// innerNSDeclarationPrefix 取出 xmlns 声明对应的前缀 (默认命名空间时为空串).
func innerNSDeclarationPrefix(attr *step1core.DOMAttribute) string {
	if attr == nil || attr.Name == "xmlns" {
		return ""
	}
	return attr.Name
}

// innerSplitNSDeclaration 判断一个 xml 属性是不是 xmlns 声明, 并拆出前缀与 URI.
func innerSplitNSDeclaration(attr xml.Attr) (string, string, bool) {
	if attr.Name.Space == "xmlns" {
		return attr.Name.Local, attr.Value, true
	}
	if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
		return "", attr.Value, true
	}
	return "", "", false
}

// innerMakeNSDeclaration 把一个 xml 的 xmlns 声明转换成 DOMAttribute.
func innerMakeNSDeclaration(attr xml.Attr) (step1core.DOMAttribute, bool) {

	prefix, _, ok := innerSplitNSDeclaration(attr)
	if !ok {
		return step1core.DOMAttribute{}, false
	}

	out := step1core.DOMAttribute{Value: attr.Value}
	if prefix == "" {
		out.Name = "xmlns"
	} else {
		out.Prefix = "xmlns"
		out.Name = prefix
	}
	return out, true
}

// innerPrefixOf 由命名空间(URI)反查前缀.
func innerPrefixOf(uri2prefix map[string]string, space string) string {
	if space == "" {
		return ""
	}
	if prefix, ok := uri2prefix[space]; ok {
		return prefix
	}
	// 没有声明过: 若它本身就是一个前缀(而不是 URI), 直接使用
	if innerIsNamespaceURI(space) {
		return ""
	}
	return space
}

// innerIsNamespaceURI 粗略判断一个字符串是不是命名空间 URI.
func innerIsNamespaceURI(text string) bool {
	return strings.Contains(text, "://") || strings.HasPrefix(text, "urn:")
}

// innerEscapeText 转义元素文本.
func innerEscapeText(text string) string {

	var buf strings.Builder
	for _, ch := range text {
		switch ch {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '\r':
			buf.WriteString("&#13;")
		default:
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}

// innerEscapeAttribute 转义属性值.
func innerEscapeAttribute(value string) string {

	var buf strings.Builder
	for _, ch := range value {
		switch ch {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\n':
			buf.WriteString("&#10;")
		case '\r':
			buf.WriteString("&#13;")
		case '\t':
			buf.WriteString("&#9;")
		default:
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}

////////////////////////////////////////////////////////////////////////////////
// EOF
