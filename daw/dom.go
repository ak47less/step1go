package step1core

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// dom.go : Step1-DAW 工程文档 (work document) 的文档对象模型 (DOM).
//
// 本文件按 docs/step1work.xsd 的定义, 给出与工程文档 XML 一一对应的 Go 类型,
// 以及 xsd 中 simpleType 的文本形式转换; 具体的实现 (元素工厂 / 编解码器) 见 src/core.
//
// 文档结构 (与 xsd 中的 xs:element 一致):
//
//	work                                     WorkElement
//	├── head                                 HeadElement
//	│   ├── title                            (元素文本)
//	│   ├── meta*                            MetaElement
//	│   └── description                      (元素文本)
//	├── timeline                             TimelineElement (可递归嵌套)
//	├── channels                             ChannelsElement
//	│   └── channel*                         ChannelElement
//	├── tracks                               TracksElement
//	│   └── track*                           TrackElement
//	│       └── slice*                       SliceElement
//	│           ├── wave                     WaveClipElement
//	│           ├── midi                     MidiClipElement
//	│           └── progression              ProgressionElement
//	│               └── chord*               ChordElement
//	└── devices                              DevicesElement
//	    └── device*                          DeviceElement
//	        └── wire*                        WireElement
//
// xsd 中的 simpleType 与 Go 类型的对应关系 (与 xsd 头部注释一致):
//
//	NameType        -> NameType                  (timeline / track / channel 的名字, 以及 slice/@timeline)
//	PortNameType    -> PortName                  (device 的端口名)
//	IDType          -> DeviceID                  (device/@id)
//	OrderIDType     -> OrderID                   (work/@orderid)
//	TrackKindType   -> TrackType                 (track/@type)
//	SliceKindType   -> TrackType (不含 master)    (slice/@type)
//	ChannelKindType -> ChannelType               (channel/@type, audio / midi)
//	DeviceClassType -> DeviceClass               (device/@class)
//	SourceRefType   -> PortRef                   (wire/@source, 形如 "设备名#端口名")
//	MetaNameType    -> MetaName                  (meta/@name)
//	NoteNameType    -> ToneName                  (chord/@root 以及 chord 的元素文本)
//	ChordNotesType  -> []ToneName                (chord 的元素文本)
//	BPMType         -> BPM                        (timeline/@bpm, 取值区间 [10,500])
//	TempoType       -> Tempo                      (timeline/@tempo, 形如 "4/4")
//	SampleTimeType  -> SampleTime                 (offset / length / chord@offset, 非负)
//
// 约定:
//  1. work 的直接子元素 (head / timeline / channels / tracks / devices) 用 xs:all 约束:
//     顺序自由, 但各自最多出现一次; work 内部其余元素仍用 xs:sequence 约束顺序
//     (例如 head 的子元素顺序依次为 title, meta*, description), 因此 typed 的
//     AddXxx() / SetXxx() 方法应当把新元素插入到符合 schema 的位置, 而不是简单追加.
//  2. xsd 在 work 上还声明了若干 xs:key / xs:keyref 语义约束, 它们不能在 schema 结构里
//     表达完整, 需要应用层保证; 见本文件末尾的 CheckWorkConstraints().
//  3. 可选属性在接口上用指针表示 (nil 表示 XML 中没有该属性); 可选子元素用 nil 表示.
//  4. 基于 ExtensionPointType 的元素 (meta / wave / midi / wire) 只允许空白文本内容,
//     但允许插入其它命名空间的子元素 (扩展点), 参见 DOMElement 的通用子元素方法.

////////////////////////////////////////////////////////////////////////////////
// 命名空间 / 标签名 / 属性名

const (
	// Step1WorkNamespace 是 xsd 的 targetNamespace; 原型文档在 work 上以默认命名空间声明.
	Step1WorkNamespace = "http://bitwormhole.com/step1/xml/ns"

	// Step1WorkSchemaFileName 是 schema 的文件名.
	Step1WorkSchemaFileName = "step1work.xsd"

	// Step1WorkSchemaLocation 是 xsi:schemaLocation 的推荐取值 (见 xsd 头部注释).
	Step1WorkSchemaLocation = Step1WorkNamespace + " " + Step1WorkSchemaFileName
)

// 元素名 (xsd 中 xs:element/@name).
const (
	TagWork        = "work"
	TagHead        = "head"
	TagTitle       = "title"
	TagMeta        = "meta"
	TagDescription = "description"
	TagTimeline    = "timeline"
	TagChannels    = "channels"
	TagChannel     = "channel"
	TagTracks      = "tracks"
	TagTrack       = "track"
	TagSlice       = "slice"
	TagWave        = "wave"
	TagMidi        = "midi"
	TagProgression = "progression"
	TagChord       = "chord"
	TagDevices     = "devices"
	TagDevice      = "device"
	TagWire        = "wire"
)

// 属性名 (xsd 中 xs:attribute/@name).
const (
	AttrName     = "name"
	AttrValue    = "value"
	AttrBPM      = "bpm"
	AttrTempo    = "tempo"
	AttrOffset   = "offset"
	AttrLength   = "length"
	AttrType     = "type"
	AttrTimeline = "timeline"
	AttrSrc      = "src"
	AttrRoot     = "root"
	AttrClass    = "class"
	AttrID       = "id"
	AttrPort     = "port"
	AttrSource   = "source"
	AttrOrderID  = "orderid"
)

////////////////////////////////////////////////////////////////////////////////
// simpleType 对应的 Go 类型

// NameType 对应 xsd 的 NameType (基于 xs:NCName): timeline / track 的名称,
// 同时也是 slice/@timeline 引用时间轴时使用的名字.
//
// 约束: 不能含空格, 不能以数字、'-'、'.' 开头, 允许中文等 Unicode 字母.
// 校验方法见 CheckNCName().
type NameType string

// MetaName 对应 xsd 的 MetaNameType: head/meta 的名称, 不含空白字符.
// 校验方法见 CheckMetaNameText().
type MetaName string

// OrderID 对应 xsd 的 OrderIDType: 由纯数字组成的工程(订单)编号.
// 校验方法见 CheckOrderIDText().
type OrderID string

// ChannelType 对应 xsd 的 ChannelKindType: channel/@type 的取值 (audio / midi).
//
// 它与 TrackType 的取值风格一致, 但不含 chord / master; 由于 step1core 目前还没有
// 独立的通道领域模型, 该枚举暂定义在 DOM 层. 文本形式的解析与格式化分别见
// ParseChannelTypeText() / FormatChannelTypeText().
type ChannelType int

const (
	ChannelTypeUnknown ChannelType = iota

	// ChannelTypeAudio 是音频通道, 对应 channel/@type="audio".
	ChannelTypeAudio

	// ChannelTypeMIDI 是 MIDI 通道, 对应 channel/@type="midi".
	ChannelTypeMIDI
)

// 原型文档 (docs/prototype.work.xml) 中出现过的 meta 名称.
const (
	MetaNameSampleRate   MetaName = "sample_rate"   // 采样率, 例如 44100
	MetaNameBufferLength MetaName = "buffer_length" // 每缓冲区样本数, 例如 128
	MetaNameTempo        MetaName = "tempo"         // 拍号, 例如 "3/4"
	MetaNameBPM          MetaName = "bpm"           // 节拍速度, 例如 88
	MetaNamePrimary      MetaName = "primary"       // 主音, 例如 "F#"
	MetaNameScale        MetaName = "scale"         // 调式, 例如 "major"
)

// BPM 的合法区间 (xsd BPMType: minInclusive=10, maxInclusive=500).
const (
	BPMRangeMin BPM = 10
	BPMRangeMax BPM = 500
)

////////////////////////////////////////////////////////////////////////////////
// simpleType 的文本形式 : 校验 / 解析 / 格式化
//
// xsd 中的 pattern / minInclusive / maxInclusive 都在这里用等价的 Go 正则与判断实现,
// 供编解码器在读写 XML 属性时使用. 名称后缀 "Text" 表示处理的是 XML 中的文本形式
// (与 music.go 等文件中可能存在的领域类型转换相区分).

var (
	// NameType / PortNameType / IDType 的基类型 xs:NCName.
	theNCNamePattern = regexp.MustCompile(`^[_\p{L}][_\p{L}\p{N}\p{M}.\-\x{B7}]*$`)

	// MetaNameType: [^\s]+
	theMetaNamePattern = regexp.MustCompile(`^\S+$`)

	// OrderIDType: [0-9]+
	theOrderIDPattern = regexp.MustCompile(`^[0-9]+$`)

	// DeviceClassType: [^\s/]+(/[^\s/]+)*
	theDeviceClassPattern = regexp.MustCompile(`^[^\s/]+(/[^\s/]+)*$`)

	// SourceRefType: [^\s#]+(#[^\s#]+)?
	theSourceRefPattern = regexp.MustCompile(`^[^\s#]+(#[^\s#]+)?$`)

	// NoteNameType: [A-Ga-g](#|b)?(-?[0-9]+)?  (允许八度, 例如 "C4"、"D-1")
	theNoteNamePattern = regexp.MustCompile(`^[A-Ga-g](#|b)?(-?[0-9]+)?$`)

	// 音级: 只含音级与升降号, 不含八度, 例如 "C"、" F#"(去空白后)、"Db".
	theToneNamePattern = regexp.MustCompile(`^[A-Ga-g](#|b)?$`)

	// ChordNotesType: 以逗号分隔的音名列表, 逗号两侧允许空白.
	theChordNotesPattern = regexp.MustCompile(`^[A-Ga-g](#|b)?(-?[0-9]+)?(\s*,\s*[A-Ga-g](#|b)?(-?[0-9]+)?)*$`)

	// TempoType: [1-9][0-9]*\s*/\s*[1-9][0-9]*
	theTempoPattern = regexp.MustCompile(`^[1-9][0-9]*\s*/\s*[1-9][0-9]*$`)
)

// CheckNCName 校验 xs:NCName 形式的名称 (xsd 的 NameType / PortNameType / IDType).
func CheckNCName(text string) error {
	if !theNCNamePattern.MatchString(text) {
		return fmt.Errorf("invalid NCName: %q", text)
	}
	return nil
}

// CheckMetaNameText 校验 xsd 的 MetaNameType (不含空白字符, 且非空).
func CheckMetaNameText(text string) error {
	if !theMetaNamePattern.MatchString(text) {
		return fmt.Errorf("invalid meta name: %q", text)
	}
	return nil
}

// CheckOrderIDText 校验 xsd 的 OrderIDType (纯数字, 且非空).
func CheckOrderIDText(text string) error {
	if !theOrderIDPattern.MatchString(text) {
		return fmt.Errorf("invalid order id: %q", text)
	}
	return nil
}

// CheckNoteNameText 校验 xsd 的 NoteNameType.
//
// 注意: 该类型允许带八度 (例如 "C4"、"D-1"); 若需要转换成 music.go 的音级类型 ToneName,
// 请使用 ParseToneNameText() (它不接受八度).
func CheckNoteNameText(text string) error {
	if !theNoteNamePattern.MatchString(text) {
		return fmt.Errorf("invalid note name: %q", text)
	}
	return nil
}

// CheckChordNotesText 校验 xsd 的 ChordNotesType (逗号分隔的音名列表).
//
// 与 xsd 一致, 该方法按 xs:token 处理: 先去掉首尾空白, 逗号两侧允许空白.
func CheckChordNotesText(text string) error {
	s := strings.TrimSpace(text)
	if !theChordNotesPattern.MatchString(s) {
		return fmt.Errorf("invalid chord notes: %q", text)
	}
	return nil
}

// ParseTrackTypeText 解析 xsd 的 TrackKindType (audio / midi / chord / master).
func ParseTrackTypeText(text string) (TrackType, error) {
	switch strings.TrimSpace(text) {
	case "audio":
		return TrackTypeAudio, nil
	case "midi":
		return TrackTypeMIDI, nil
	case "chord":
		return TrackTypeChord, nil
	case "master":
		return TrackTypeMaster, nil
	default:
		return TrackTypeUnknown, fmt.Errorf("invalid track type: %q", text)
	}
}

// FormatTrackTypeText 把 TrackType 转换为 xsd 的 TrackKindType 文本.
// 未知类型返回空串, 调用方应当在写出之前先行判断.
func FormatTrackTypeText(t TrackType) string {
	switch t {
	case TrackTypeAudio:
		return "audio"
	case TrackTypeMIDI:
		return "midi"
	case TrackTypeChord:
		return "chord"
	case TrackTypeMaster:
		return "master"
	default:
		return ""
	}
}

// CheckSliceType 校验 slice/@type 的取值 (xsd 的 SliceKindType, 不含 master).
func CheckSliceType(t TrackType) error {
	switch t {
	case TrackTypeAudio, TrackTypeMIDI, TrackTypeChord:
		return nil
	default:
		return fmt.Errorf("invalid slice type: %s", FormatTrackTypeText(t))
	}
}

// ParseChannelTypeText 解析 xsd 的 ChannelKindType (audio / midi).
func ParseChannelTypeText(text string) (ChannelType, error) {
	switch strings.TrimSpace(text) {
	case "audio":
		return ChannelTypeAudio, nil
	case "midi":
		return ChannelTypeMIDI, nil
	default:
		return ChannelTypeUnknown, fmt.Errorf("invalid channel type: %q", text)
	}
}

// FormatChannelTypeText 把 ChannelType 转换为 xsd 的 ChannelKindType 文本.
// 未知类型返回空串, 调用方应当在写出之前先行判断.
func FormatChannelTypeText(t ChannelType) string {
	switch t {
	case ChannelTypeAudio:
		return "audio"
	case ChannelTypeMIDI:
		return "midi"
	default:
		return ""
	}
}

// ParseTempoText 解析 xsd 的 TempoType, 例如 "4/4"、" 3 / 8 ".
func ParseTempoText(text string) (Tempo, error) {
	s := strings.TrimSpace(text)
	if !theTempoPattern.MatchString(s) {
		return Tempo{}, fmt.Errorf("invalid tempo: %q", text)
	}
	parts := strings.SplitN(s, "/", 2)
	upper, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return Tempo{}, fmt.Errorf("invalid tempo: %q", text)
	}
	lower, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return Tempo{}, fmt.Errorf("invalid tempo: %q", text)
	}
	return Tempo{Upper: uint(upper), Lower: uint(lower)}, nil
}

// FormatTempoText 把 Tempo 转换为 xsd 的 TempoType 文本, 例如 "4/4".
func FormatTempoText(t Tempo) string {
	return fmt.Sprintf("%d/%d", t.Upper, t.Lower)
}

// ParseBPMText 解析 xsd 的 BPMType, 取值范围为 [BPMRangeMin, BPMRangeMax], 允许小数.
func ParseBPMText(text string) (BPM, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(text), 32)
	if err != nil {
		return 0, fmt.Errorf("invalid bpm: %q", text)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("invalid bpm: %q", text)
	}
	b := BPM(f)
	if b < BPMRangeMin || b > BPMRangeMax {
		return 0, fmt.Errorf("bpm out of range [%v,%v]: %q", BPMRangeMin, BPMRangeMax, text)
	}
	return b, nil
}

// FormatBPMText 把 BPM 转换为 xsd 的 BPMType 文本, 例如 88、78.9.
func FormatBPMText(b BPM) string {
	return strconv.FormatFloat(float64(b), 'f', -1, 32)
}

// ParseSampleTimeText 解析 xsd 的 SampleTimeType (非负整数).
func ParseSampleTimeText(text string) (SampleTime, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid sample time: %q", text)
	}
	if n < 0 {
		return 0, fmt.Errorf("sample time must not be negative: %q", text)
	}
	return SampleTime(n), nil
}

// FormatSampleTimeText 把 SampleTime 转换为 xsd 的 SampleTimeType 文本.
func FormatSampleTimeText(t SampleTime) string {
	return strconv.FormatInt(int64(t), 10)
}

// ParseDeviceClassText 解析 xsd 的 DeviceClassType, 例如 "gain"、"step1/classes/gain".
func ParseDeviceClassText(text string) (DeviceClass, error) {
	s := strings.TrimSpace(text)
	if !theDeviceClassPattern.MatchString(s) {
		return "", fmt.Errorf("invalid device class: %q", text)
	}
	return DeviceClass(s), nil
}

// FormatDeviceClassText 把 DeviceClass 转换为 xsd 的 DeviceClassType 文本.
func FormatDeviceClassText(c DeviceClass) string {
	return string(c)
}

// ParseSourceRefText 解析 xsd 的 SourceRefType, 即 "设备名#端口名".
//
// 端口名可以省略 (此时 PortRef.Port 为空串), 表示引用整个设备.
func ParseSourceRefText(text string) (PortRef, error) {
	s := strings.TrimSpace(text)
	if !theSourceRefPattern.MatchString(s) {
		return PortRef{}, fmt.Errorf("invalid source ref: %q", text)
	}
	ref := PortRef{}
	index := strings.IndexByte(s, '#')
	if index < 0 {
		ref.Device = DeviceName(s)
	} else {
		ref.Device = DeviceName(s[:index])
		ref.Port = PortName(s[index+1:])
	}
	return ref, nil
}

// FormatSourceRefText 把 PortRef 转换为 xsd 的 SourceRefType 文本.
func FormatSourceRefText(ref PortRef) string {
	if ref.Port == "" {
		return string(ref.Device)
	}
	return string(ref.Device) + "#" + string(ref.Port)
}

// ParseToneNameText 解析只有音级的音名 (xsd NoteNameType 去掉可选八度的那一部分).
//
// 大小写与升降号都会被规范化: 返回值的写法与 music.go 中 ToneName 的常量一致,
// 即大写音级 + '#' (例如 "db" -> "C#", "B#" -> "C", "Cb" -> "B").
//
// 带八度的形式 (例如 "C4") 会返回错误, 因为 ToneName 只能表示音级;
// 若需要保留八度, 请使用 CheckNoteNameText() 或 music.go 的 PitchName.
func ParseToneNameText(text string) (ToneName, error) {
	s := strings.TrimSpace(text)
	if !theToneNamePattern.MatchString(s) {
		return "", fmt.Errorf("invalid tone name: %q", text)
	}
	num := ToneName(s).Number()
	return num.Name(), nil
}

// ParseChordNotesText 解析 xsd 的 ChordNotesType, 例如 "C,E,G"、" D#, F , A ".
//
// 空文本返回空列表 (注意: xsd 的 pattern 要求至少一个音名, 编码时不应写出空文本);
// 带八度的音名 (xsd 允许, 例如 "C4") 会返回错误, 原因同 ParseToneNameText().
func ParseChordNotesText(text string) ([]ToneName, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return nil, nil
	}
	if !theChordNotesPattern.MatchString(s) {
		return nil, fmt.Errorf("invalid chord notes: %q", text)
	}
	parts := strings.Split(s, ",")
	notes := make([]ToneName, 0, len(parts))
	for _, part := range parts {
		name, err := ParseToneNameText(part)
		if err != nil {
			return nil, fmt.Errorf("chord note %q: %w", strings.TrimSpace(part), err)
		}
		notes = append(notes, name)
	}
	return notes, nil
}

// FormatChordNotesText 把音名列表转换为 xsd 的 ChordNotesType 文本, 例如 "C,E,G".
func FormatChordNotesText(notes []ToneName) string {
	buf := make([]string, 0, len(notes))
	for _, item := range notes {
		buf = append(buf, string(item))
	}
	return strings.Join(buf, ",")
}

// ResolveSliceType 按 xsd 的约定返回片段的有效类型:
// slice/@type 可以省略, 省略时与父 track/@type 一致.
func ResolveSliceType(track TrackElement, slice SliceElement) TrackType {
	if slice != nil {
		if t := slice.GetType(); t != nil {
			return *t
		}
	}
	if track != nil {
		return track.GetType()
	}
	return TrackTypeUnknown
}

////////////////////////////////////////////////////////////////////////////////
// 基础接口

// DOMNode 是所有 DOM 节点 (文档 / 元素) 的公共接口.
type DOMNode interface {
	// GetOwner 返回节点所属的文档.
	GetOwner() DOMDocument
}

// DOMAttribute 是一个 XML 属性.
type DOMAttribute struct {
	Namespace string
	Prefix    string
	Name      string
	Value     string
}

// DOMElement 是一个 XML 元素, 也是下文所有具体元素接口的基类型.
//
// 它提供最基础的读写能力 (标签 / 文本 / 属性 / 子元素), 供通用的遍历与编解码使用;
// 具体元素接口 (WorkElement 等) 在此之上给出与 xsd 一一对应的 typed 访问方法.
type DOMElement interface {
	DOMNode

	// GetParent 返回父元素; 根元素没有父元素, 返回 nil.
	GetParent() DOMElement

	// GetNamespace / GetPrefix / GetTag 返回元素的限定名.
	GetNamespace() string
	GetPrefix() string
	GetTag() string

	// GetText / SetText 读写元素的文本内容 (例如 title 的标题, chord 的组成音).
	GetText() string
	SetText(text string) error

	// GetAttribute 读取属性; 属性不存在时返回 ("", false).
	GetAttribute(name string) (string, bool)

	// SetAttribute 写入属性; 取值为空串时仍然写出该属性.
	SetAttribute(name string, value string) error

	// RemoveAttribute 删除属性.
	RemoveAttribute(name string) error

	// ListAttributes 返回属性的快照.
	ListAttributes() []DOMAttribute

	// AddChild 追加子元素; RemoveChild 删除直接子元素.
	AddChild(child DOMElement) error
	RemoveChild(child DOMElement) error

	// ListChildren 返回直接子元素的快照.
	ListChildren() []DOMElement

	// ListChildrenByTag 按标签名筛选直接子元素 (不限定命名空间).
	ListChildrenByTag(tag string) []DOMElement
}

// DOMDocument 是一篇工程文档 (work document).
type DOMDocument interface {
	DOMNode

	Implementation() DOMImplementation

	// GetURL 返回文档的位置 (通常是文件路径); 新建的空文档可以为空串.
	GetURL() string

	// GetRoot 返回根元素 work.
	GetRoot() (WorkElement, error)

	// SetRoot 设置(替换)根元素 work.
	SetRoot(root WorkElement) error
}

// DOMDocumentCodec 是工程文档的编解码器 (XML 文本 <-> DOM).
type DOMDocumentCodec interface {
	Encode(doc DOMDocument) ([]byte, error)

	Decode(data []byte, doc DOMDocument) error
}

// DOMImplementation 是 DOM 的实现 (提供工厂与编解码器).
type DOMImplementation interface {
	Codec() DOMDocumentCodec

	Factory() DOMFactory
}

// DOMFactory 按 xsd 中定义的元素创建对应的 typed element.
//
// 除了通用的 CreateElement() 之外, 每一种具体的元素都有专门的创建方法;
// 由这些方法创建出来的元素, 其命名空间/标签已经符合 Step1WorkNamespace 的定义.
type DOMFactory interface {
	CreateDocument(url string) DOMDocument

	// CreateElement 创建一个通用元素 (用于 title、扩展点等没有专门接口的元素).
	CreateElement(doc DOMDocument, namespace string, prefix string, tag string) (DOMElement, error)

	CreateWorkElement(doc DOMDocument) (WorkElement, error)

	CreateHeadElement(doc DOMDocument) (HeadElement, error)

	CreateMetaElement(doc DOMDocument) (MetaElement, error)

	CreateTimelineElement(doc DOMDocument) (TimelineElement, error)

	CreateChannelsElement(doc DOMDocument) (ChannelsElement, error)

	CreateChannelElement(doc DOMDocument) (ChannelElement, error)

	CreateTracksElement(doc DOMDocument) (TracksElement, error)

	CreateTrackElement(doc DOMDocument) (TrackElement, error)

	CreateSliceElement(doc DOMDocument) (SliceElement, error)

	CreateWaveClipElement(doc DOMDocument) (WaveClipElement, error)

	CreateMidiClipElement(doc DOMDocument) (MidiClipElement, error)

	CreateProgressionElement(doc DOMDocument) (ProgressionElement, error)

	CreateChordElement(doc DOMDocument) (ChordElement, error)

	CreateDevicesElement(doc DOMDocument) (DevicesElement, error)

	CreateDeviceElement(doc DOMDocument) (DeviceElement, error)

	CreateWireElement(doc DOMDocument) (WireElement, error)
}

// DOM 是工程文档 DOM 的门面: 把实现、工厂与编解码器组合在一起.
type DOM struct {
	Document DOMDocument

	Factory DOMFactory

	Codec DOMDocumentCodec

	Implementation DOMImplementation
}

// NewDOM 基于给定的实现装配一个 DOM 门面.
func NewDOM(impl DOMImplementation) *DOM {
	inst := new(DOM)
	inst.SetImplementation(impl)
	return inst
}

// SetImplementation 设置 DOM 的实现, 并同步更新 Factory 与 Codec.
func (inst *DOM) SetImplementation(impl DOMImplementation) *DOM {
	inst.Implementation = impl
	if impl == nil {
		inst.Factory = nil
		inst.Codec = nil
	} else {
		inst.Factory = impl.Factory()
		inst.Codec = impl.Codec()
	}
	return inst
}

// Create 新建一篇空文档, 并创建空的 work 根元素.
func (inst *DOM) Create(url string) (DOMDocument, error) {

	factory := inst.Factory
	if factory == nil {
		return nil, fmt.Errorf("DOM.Factory is nil")
	}

	doc := factory.CreateDocument(url)
	if doc == nil {
		return nil, fmt.Errorf("DOM.Factory.CreateDocument() returns nil")
	}

	root, err := factory.CreateWorkElement(doc)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("DOM.Factory.CreateWorkElement() returns nil")
	}

	err = doc.SetRoot(root)
	if err != nil {
		return nil, err
	}

	inst.Document = doc
	return doc, nil
}

// Load 把 XML 数据解析到 DOM.Document; 文档不存在时先按空 url 新建.
func (inst *DOM) Load(data []byte) error {

	codec := inst.Codec
	if codec == nil {
		return fmt.Errorf("DOM.Codec is nil")
	}

	doc := inst.Document
	if doc == nil {
		created, err := inst.Create("")
		if err != nil {
			return err
		}
		doc = created
	}

	return codec.Decode(data, doc)
}

// Save 把 DOM.Document 编码为 XML 数据.
func (inst *DOM) Save() ([]byte, error) {

	codec := inst.Codec
	if codec == nil {
		return nil, fmt.Errorf("DOM.Codec is nil")
	}

	doc := inst.Document
	if doc == nil {
		return nil, fmt.Errorf("DOM.Document is nil")
	}

	return codec.Encode(doc)
}

// GetRoot 返回当前文档的根元素 work.
func (inst *DOM) GetRoot() (WorkElement, error) {
	doc := inst.Document
	if doc == nil {
		return nil, fmt.Errorf("DOM.Document is nil")
	}
	return doc.GetRoot()
}

// CheckConstraints 检查当前文档中 xsd 声明的语义约束 (xs:key / xs:keyref).
func (inst *DOM) CheckConstraints() ([]*WorkConstraintError, error) {
	root, err := inst.GetRoot()
	if err != nil {
		return nil, err
	}
	return CheckWorkConstraints(root), nil
}

////////////////////////////////////////////////////////////////////////////////
// 各元素的接口 (与 xsd 的 complexType 一一对应)

// WorkElement 对应 work / WorkType: 工程文档的根元素.
//
// 子元素用 xs:all 约束: head? / timeline? / channels? / tracks? / devices?,
// 顺序自由, 但各自最多出现一次 (timeline 的子时间轴嵌套由 TimelineElement 递归表达).
// 传给 SetHead / SetTimeline / SetChannels / SetTracks / SetDevices 的实参为 nil 时
// 表示删除该子元素.
type WorkElement interface {
	DOMElement

	// GetOrderID 对应 work/@orderid (OrderIDType); 未设置时返回空串.
	GetOrderID() OrderID
	SetOrderID(id OrderID) error

	// GetHead 对应 head 子元素; 不存在时返回 nil.
	GetHead() HeadElement
	SetHead(head HeadElement) error

	// GetTimeline 对应 timeline 子元素 (最多一个); 不存在时返回 nil.
	GetTimeline() TimelineElement
	SetTimeline(item TimelineElement) error

	// GetChannels 对应 channels 子元素; 不存在时返回 nil.
	GetChannels() ChannelsElement
	SetChannels(channels ChannelsElement) error

	// GetTracks 对应 tracks 子元素; 不存在时返回 nil.
	GetTracks() TracksElement
	SetTracks(tracks TracksElement) error

	// GetDevices 对应 devices 子元素; 不存在时返回 nil.
	GetDevices() DevicesElement
	SetDevices(devices DevicesElement) error

	// 以下查找方法对应 xsd 在 work 上声明的 key / keyref:
	// timeline/@name 全文唯一 (含各级子时间轴), channel/@name 全文唯一,
	// track/@name 全文唯一, device/@name 全文唯一, device/@id 若存在则唯一.
	// 找不到时返回错误.

	FindTimeline(name NameType) (TimelineElement, error)
	FindChannel(name NameType) (ChannelElement, error)
	FindTrack(name NameType) (TrackElement, error)
	FindDevice(name DeviceName) (DeviceElement, error)
	FindDeviceByID(id DeviceID) (DeviceElement, error)
}

// HeadElement 对应 head / HeadType: title, 之后是任意多条 meta, 最后是可选的一段
// description (顺序由 xs:sequence 约束).
type HeadElement interface {
	DOMElement

	// GetTitle / SetTitle 对应 head/title 的文本.
	GetTitle() string
	SetTitle(title string) error

	// GetDescription / SetDescription 对应 head/description 的文本 (可选, 最多一个);
	// 不存在时 GetDescription 返回空串, SetDescription("") 表示删除该元素.
	GetDescription() string
	SetDescription(text string) error

	// ListMeta 返回全部 meta (对应 MetaNameKey: 同名 meta 不允许重复).
	ListMeta() []MetaElement
	AddMeta(item MetaElement) error

	// FindMeta 按名字查找 meta, 找不到时返回错误.
	FindMeta(name MetaName) (MetaElement, error)
}

// MetaElement 对应 meta / MetaType (基于 ExtensionPointType): 一条元信息.
//
// 取值一律以文本形式给出 (例如 value="44100"), 元素本身只允许空白文本内容,
// 但允许插入其它命名空间的扩展子元素.
type MetaElement interface {
	DOMElement

	GetName() MetaName
	SetName(name MetaName) error

	GetValue() string
	SetValue(value string) error
}

// TimelineElement 对应 timeline / TimelineType: 时间轴节点, 可递归嵌套.
//
// name 为必填; bpm / tempo / offset / length 为可选, 返回 nil 表示 XML 中没有该属性.
// offset 与 length 都是相对父时间轴的样本数.
type TimelineElement interface {
	DOMElement

	GetName() NameType
	SetName(name NameType) error

	GetBPM() *BPM
	SetBPM(value *BPM) error

	GetTempo() *Tempo
	SetTempo(value *Tempo) error

	GetOffset() *SampleTime
	SetOffset(value *SampleTime) error

	GetLength() *SampleTime
	SetLength(value *SampleTime) error

	ListTimelines() []TimelineElement
	AddTimeline(item TimelineElement) error
}

// ChannelsElement 对应 channels / ChannelsType: 通道集合.
type ChannelsElement interface {
	DOMElement

	ListChannels() []ChannelElement
	AddChannel(item ChannelElement) error
}

// ChannelElement 对应 channel / ChannelType (基于 ExtensionPointType): 一条通道,
// 例如音频输出通道 audio_left / audio_right, MIDI 通道 midi_notes / midi_clock.
//
// name 与 type 都是必填 (对应 ChannelNameKey: channel/@name 全文唯一).
// 元素本身只允许空白文本内容, 但允许插入其它命名空间的扩展子元素;
// 通道与轨道 / 端口之间的对应关系由应用层决定.
type ChannelElement interface {
	DOMElement

	GetName() NameType
	SetName(name NameType) error

	GetType() ChannelType
	SetType(t ChannelType) error
}

// TracksElement 对应 tracks / TracksType: 轨道集合.
type TracksElement interface {
	DOMElement

	ListTracks() []TrackElement
	AddTrack(item TrackElement) error
}

// TrackElement 对应 track / TrackType: 一条轨道.
//
// name 与 type 都是必填. audio / midi / chord 轨道可以包含 slice,
// master 轨道不放片段 (xsd 的 SliceTimelineRef 与结构无关, 由应用层保证).
type TrackElement interface {
	DOMElement

	GetName() NameType
	SetName(name NameType) error

	GetType() TrackType
	SetType(t TrackType) error

	ListSlices() []SliceElement
	AddSlice(item SliceElement) error
}

// SliceElement 对应 slice / SliceType: 轨道在某个时间轴上的一个片段.
//
// timeline 为必填, 必须指向已存在的 timeline/@name (SliceTimelineRef);
// type 可选, 缺省时与父 track/@type 一致, 可用 ResolveSliceType() 计算有效值.
//
// 片段内容三选一 (xs:choice): wave / midi / progression, 也可以都没有;
// Set 其中之一时应当清除另外两个, 未设置的那个返回 nil.
type SliceElement interface {
	DOMElement

	GetTimeline() NameType
	SetTimeline(name NameType) error

	GetType() *TrackType
	SetType(value *TrackType) error

	GetWaveClip() WaveClipElement
	SetWaveClip(item WaveClipElement) error

	GetMidiClip() MidiClipElement
	SetMidiClip(item MidiClipElement) error

	GetProgression() ProgressionElement
	SetProgression(item ProgressionElement) error
}

// WaveClipElement 对应 wave / WaveClipType (基于 ExtensionPointType): 音频片段.
type WaveClipElement interface {
	DOMElement

	// GetSrc 对应 wave/@src, 可选; 相对路径以工程文档所在目录为基准.
	GetSrc() string
	SetSrc(src string) error
}

// MidiClipElement 对应 midi / MidiClipType (基于 ExtensionPointType): MIDI 片段.
type MidiClipElement interface {
	DOMElement

	// GetSrc 对应 midi/@src, 可选, 允许为空串 (表示内联或尚未绑定文件).
	GetSrc() string
	SetSrc(src string) error
}

// ProgressionElement 对应 progression / ProgressionType: 和弦进行.
type ProgressionElement interface {
	DOMElement

	ListChords() []ChordElement
	AddChord(item ChordElement) error
}

// ChordElement 对应 chord / ChordType: 一个和弦.
//
// 它使用 simpleContent: root 属性给出根音, offset 给出位置(样本数, 缺省视为 0),
// 元素文本给出组成音, 例如 "C,E,G" —— 即 DOMElement 的 GetText / SetText,
// 也可以用便于使用的 GetNotes / SetNotes 读写解析后的音级列表.
type ChordElement interface {
	DOMElement

	GetRoot() ToneName
	SetRoot(root ToneName) error

	GetOffset() *SampleTime
	SetOffset(value *SampleTime) error

	GetNotes() []ToneName
	SetNotes(notes []ToneName) error
}

// DevicesElement 对应 devices / DevicesType: 设备集合.
type DevicesElement interface {
	DOMElement

	ListDevices() []DeviceElement
	AddDevice(item DeviceElement) error
}

// DeviceElement 对应 device / DeviceType: 一台设备 (音频引擎、时钟、ADDA、增益、混音器等).
//
// name 与 class 必填, id 可选 (未设置时返回空串).
type DeviceElement interface {
	DOMElement

	GetName() DeviceName
	SetName(name DeviceName) error

	GetClass() DeviceClass
	SetClass(class DeviceClass) error

	GetID() DeviceID
	SetID(id DeviceID) error

	ListWires() []WireElement
	AddWire(item WireElement) error
}

// WireElement 对应 wire / WireType (基于 ExtensionPointType): 一条连线.
//
// 数据方向: source(源) -> port(本设备上的目标端口), 例如
// <wire port="ch1" source="gain1#output"/>.
// 源引用在 XML 中是 "设备名#端口名" 文本, 这里用 wire.go 的 PortRef 表示;
// 若只写设备名 (端口名省略), 则 PortRef.Port 为空串.
// 运行时的 EngineConfiguration 需要的是 WireConfig, 可按下面的方式组装:
//
//	WireConfig{
//	    Source:      wire.GetSource(),                                   // 来自 wire/@source
//	    Destination: PortRef{Device: device.GetName(), Port: wire.GetPort()},
//	}
type WireElement interface {
	DOMElement

	GetPort() PortName
	SetPort(port PortName) error

	GetSource() PortRef
	SetSource(source PortRef) error
}

////////////////////////////////////////////////////////////////////////////////
// 语义约束 (对应 xsd 中 work 上的 xs:key / xs:keyref)

// WorkConstraint 是 xsd 在 work 元素上声明的一条语义约束的名字.
type WorkConstraint string

const (
	// WorkConstraintTimelineNameKey : timeline/@name 全文唯一.
	WorkConstraintTimelineNameKey WorkConstraint = "TimelineNameKey"

	// WorkConstraintSliceTimelineRef : slice/@timeline 必须指向已存在的 timeline.
	WorkConstraintSliceTimelineRef WorkConstraint = "SliceTimelineRef"

	// WorkConstraintChannelNameKey : channel/@name 全文唯一.
	WorkConstraintChannelNameKey WorkConstraint = "ChannelNameKey"

	// WorkConstraintTrackNameKey : track/@name 全文唯一.
	WorkConstraintTrackNameKey WorkConstraint = "TrackNameKey"

	// WorkConstraintDeviceNameKey : device/@name 全文唯一 (wire/@source 按设备名寻址).
	WorkConstraintDeviceNameKey WorkConstraint = "DeviceNameKey"

	// WorkConstraintDeviceIdUnique : device/@id 若存在则必须唯一.
	WorkConstraintDeviceIdUnique WorkConstraint = "DeviceIdUnique"

	// WorkConstraintMetaNameKey : 同一个 head 下的 meta/@name 唯一.
	WorkConstraintMetaNameKey WorkConstraint = "MetaNameKey"
)

// WorkConstraintError 描述一处被破坏的语义约束.
type WorkConstraintError struct {
	// Constraint 是约束名 (对应 xsd 中 xs:key / xs:keyref 的 name).
	Constraint WorkConstraint

	// Path 是出错元素的位置, 下标从 1 开始 (与 XPath 一致),
	// 例如 "work/tracks/track[2]/slice[1]".
	Path string

	// Value 是引起冲突的取值 (名称或 id).
	Value string

	// Message 是可读的说明.
	Message string
}

func (inst *WorkConstraintError) Error() string {
	if inst == nil {
		return ""
	}
	return fmt.Sprintf("[%s] %s: %s", inst.Constraint, inst.Path, inst.Message)
}

// CheckWorkConstraints 按 xsd 中声明的 key / keyref 检查一篇工程文档, 返回全部错误.
//
// 检查内容:
//
//	TimelineNameKey  - timeline/@name 全文唯一 (含各级子时间轴)
//	SliceTimelineRef - slice/@timeline 必须指向已存在的 timeline
//	ChannelNameKey   - channel/@name 全文唯一
//	TrackNameKey     - track/@name 全文唯一
//	DeviceNameKey    - device/@name 全文唯一
//	DeviceIdUnique   - device/@id 若存在则必须唯一
//	MetaNameKey      - 同一个 head 下的 meta/@name 唯一
//
// 结构约束 (work 的 xs:all, 其余元素的 xs:sequence, 以及 simpleType 的取值范围)
// 由 xsd 描述, 请在读写时使用 CheckNCName / ParseBPMText / CheckSliceType 等函数逐一校验.
func CheckWorkConstraints(work WorkElement) []*WorkConstraintError {

	list := make([]*WorkConstraintError, 0)
	if work == nil {
		return list
	}

	add := func(c WorkConstraint, path string, value string, message string) {
		item := new(WorkConstraintError)
		item.Constraint = c
		item.Path = path
		item.Value = value
		item.Message = message
		list = append(list, item)
	}

	// 1. TimelineNameKey : timeline/@name 全文唯一 (含各级子时间轴)
	timelineIndex := make(map[NameType]string)
	var checkTimelines func(items []TimelineElement, path string)
	checkTimelines = func(items []TimelineElement, path string) {
		for index, item := range items {
			if item == nil {
				continue
			}
			here := fmt.Sprintf("%s[%d]", path, index+1)
			name := item.GetName()
			old, exists := timelineIndex[name]
			if exists {
				add(WorkConstraintTimelineNameKey, here, string(name), fmt.Sprintf("时间轴名称重复, 已出现在 %s", old))
			} else {
				timelineIndex[name] = here
			}
			checkTimelines(item.ListTimelines(), here+"/timeline")
		}
	}
	// work 下的 timeline 最多一个 (xs:all), 子时间轴的嵌套由 TimelineElement 表达
	if top := work.GetTimeline(); top != nil {
		checkTimelines([]TimelineElement{top}, "work/timeline")
	}

	// 2. TrackNameKey : track/@name 全文唯一
	// 3. SliceTimelineRef : slice/@timeline 必须指向已存在的 timeline
	tracks := work.GetTracks()
	if tracks != nil {
		trackIndex := make(map[NameType]string)
		for index, item := range tracks.ListTracks() {
			if item == nil {
				continue
			}
			here := fmt.Sprintf("work/tracks/track[%d]", index+1)
			name := item.GetName()
			old, exists := trackIndex[name]
			if exists {
				add(WorkConstraintTrackNameKey, here, string(name), fmt.Sprintf("轨道名称重复, 已出现在 %s", old))
			} else {
				trackIndex[name] = here
			}
			for si, slice := range item.ListSlices() {
				if slice == nil {
					continue
				}
				ref := slice.GetTimeline()
				if _, ok := timelineIndex[ref]; !ok {
					path := fmt.Sprintf("%s/slice[%d]", here, si+1)
					add(WorkConstraintSliceTimelineRef, path, string(ref), "片段引用了不存在的 timeline")
				}
			}
		}
	}

	// 4. ChannelNameKey : channel/@name 全文唯一
	channels := work.GetChannels()
	if channels != nil {
		channelIndex := make(map[NameType]string)
		for index, item := range channels.ListChannels() {
			if item == nil {
				continue
			}
			here := fmt.Sprintf("work/channels/channel[%d]", index+1)
			name := item.GetName()
			old, exists := channelIndex[name]
			if exists {
				add(WorkConstraintChannelNameKey, here, string(name), fmt.Sprintf("通道名称重复, 已出现在 %s", old))
			} else {
				channelIndex[name] = here
			}
		}
	}

	// 5. DeviceNameKey  : device/@name 全文唯一
	// 6. DeviceIdUnique : device/@id 若存在则唯一
	devices := work.GetDevices()
	if devices != nil {
		deviceNameIndex := make(map[DeviceName]string)
		deviceIDIndex := make(map[DeviceID]string)
		for index, item := range devices.ListDevices() {
			if item == nil {
				continue
			}
			here := fmt.Sprintf("work/devices/device[%d]", index+1)
			name := item.GetName()
			old, exists := deviceNameIndex[name]
			if exists {
				add(WorkConstraintDeviceNameKey, here, string(name), fmt.Sprintf("设备名称重复, 已出现在 %s", old))
			} else {
				deviceNameIndex[name] = here
			}
			id := item.GetID()
			if id != "" {
				old, exists = deviceIDIndex[id]
				if exists {
					add(WorkConstraintDeviceIdUnique, here, string(id), fmt.Sprintf("设备 id 重复, 已出现在 %s", old))
				} else {
					deviceIDIndex[id] = here
				}
			}
		}
	}

	// 7. MetaNameKey : 同一个 head 下的 meta/@name 唯一
	head := work.GetHead()
	if head != nil {
		metaIndex := make(map[MetaName]string)
		for index, item := range head.ListMeta() {
			if item == nil {
				continue
			}
			here := fmt.Sprintf("work/head/meta[%d]", index+1)
			name := item.GetName()
			old, exists := metaIndex[name]
			if exists {
				add(WorkConstraintMetaNameKey, here, string(name), fmt.Sprintf("meta 名称重复, 已出现在 %s", old))
			} else {
				metaIndex[name] = here
			}
		}
	}

	return list
}

////////////////////////////////////////////////////////////////////////////////
// EOF
