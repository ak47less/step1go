package core

import (
	"os"
	"strings"
	"testing"

	"github.com/ak47less/step1go/step1core"
)

// dom_test.go : 对 dom_impl.go 中各个组件的测试.
//
// 覆盖:
//   - 编解码 : 用 docs/prototype.work.xml 做解码 / 编码 / 往返稳定性测试;
//   - 工厂   : 从零创建文档, 以及各 typed element 的创建与读写;
//   - 结构   : xs:sequence 的子元素顺序, xs:choice 的互斥;
//   - 名称   : 命名空间 / 前缀, 以及扩展元素;
//   - 校验   : 非法取值与非法结构的错误处理;
//   - 约束   : 与 CheckWorkConstraints() 的配合.

////////////////////////////////////////////////////////////////////////////////
// 测试辅助

const domPrototypeFile = "../../../docs/prototype.work.xml"

// newTestDOM 新建一个 DOM 门面.
func newTestDOM() *step1core.DOM {
	return step1core.NewDOM(GetDomImplementation())
}

// loadPrototype 读取原型文档 docs/prototype.work.xml.
func loadPrototype(t *testing.T) (*step1core.DOM, []byte) {
	t.Helper()
	data, err := os.ReadFile(domPrototypeFile)
	if err != nil {
		t.Fatalf("read prototype: %v", err)
	}
	dom := newTestDOM()
	if err := dom.Load(data); err != nil {
		t.Fatalf("load prototype: %v", err)
	}
	return dom, data
}

// requireRoot 返回文档的根元素.
func requireRoot(t *testing.T, dom *step1core.DOM) step1core.WorkElement {
	t.Helper()
	root, err := dom.GetRoot()
	if err != nil {
		t.Fatalf("GetRoot: %v", err)
	}
	if root == nil {
		t.Fatal("GetRoot: nil")
	}
	return root
}

// requireCreate 新建一篇空文档, 并返回它.
func requireCreate(t *testing.T, dom *step1core.DOM, url string) step1core.DOMDocument {
	t.Helper()
	doc, err := dom.Create(url)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if doc == nil {
		t.Fatal("Create: nil document")
	}
	return doc
}

// requireNoError 断言没有错误.
func requireNoError(t *testing.T, what string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

// requireEqual 断言两个可比较的值相等.
func requireEqual[T comparable](t *testing.T, what string, got T, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

// requireOptional 断言一个可选值存在且相等.
func requireOptional[T comparable](t *testing.T, what string, got *T, want T) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %v", what, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %v, want %v", what, *got, want)
	}
}

// requireNilOptional 断言一个可选值不存在.
func requireNilOptional[T comparable](t *testing.T, what string, got *T) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %v, want nil", what, *got)
	}
}

// requireSlice 断言两个切片相等.
func requireSlice[T comparable](t *testing.T, what string, got []T, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", what, got, want)
		return
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("%s = %v, want %v", what, got, want)
			return
		}
	}
}

// requireContains 断言文本里包含某个片段.
func requireContains(t *testing.T, what string, text string, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Errorf("%s: %q not found in:\n%s", what, want, text)
	}
}

// summary 生成一份文档结构摘要, 用于比较两篇文档是否等价.
func summary(root step1core.WorkElement) string {

	buf := new(strings.Builder)

	if head := root.GetHead(); head != nil {
		buf.WriteString("title=" + head.GetTitle() + "\n")
		for _, meta := range head.ListMeta() {
			buf.WriteString("meta:" + string(meta.GetName()) + "=" + meta.GetValue() + "\n")
		}
		buf.WriteString("description=" + head.GetDescription() + "\n")
	}

	if channels := root.GetChannels(); channels != nil {
		for _, channel := range channels.ListChannels() {
			buf.WriteString("channel:" + string(channel.GetName()) + "/" + step1core.FormatChannelTypeText(channel.GetType()) + "\n")
		}
	}

	var walk func(items []step1core.TimelineElement)
	walk = func(items []step1core.TimelineElement) {
		for _, item := range items {
			buf.WriteString("timeline:" + string(item.GetName()))
			if v := item.GetBPM(); v != nil {
				buf.WriteString(" bpm=" + step1core.FormatBPMText(*v))
			}
			if v := item.GetTempo(); v != nil {
				buf.WriteString(" tempo=" + step1core.FormatTempoText(*v))
			}
			if v := item.GetOffset(); v != nil {
				buf.WriteString(" offset=" + step1core.FormatSampleTimeText(*v))
			}
			if v := item.GetLength(); v != nil {
				buf.WriteString(" length=" + step1core.FormatSampleTimeText(*v))
			}
			buf.WriteString("\n")
			walk(item.ListTimelines())
		}
	}
	if top := root.GetTimeline(); top != nil {
		walk([]step1core.TimelineElement{top})
	}

	if tracks := root.GetTracks(); tracks != nil {
		for _, track := range tracks.ListTracks() {
			buf.WriteString("track:" + string(track.GetName()) + "/" + step1core.FormatTrackTypeText(track.GetType()) + "\n")
			for _, slice := range track.ListSlices() {
				buf.WriteString("  slice:" + string(slice.GetTimeline()))
				if v := slice.GetType(); v != nil {
					buf.WriteString("/" + step1core.FormatTrackTypeText(*v))
				}
				buf.WriteString("\n")
				if v := slice.GetWaveClip(); v != nil {
					buf.WriteString("    wave:" + v.GetSrc() + "\n")
				}
				if v := slice.GetMidiClip(); v != nil {
					buf.WriteString("    midi:" + v.GetSrc() + "\n")
				}
				if v := slice.GetProgression(); v != nil {
					for _, chord := range v.ListChords() {
						buf.WriteString("    chord:" + string(chord.GetRoot()))
						if o := chord.GetOffset(); o != nil {
							buf.WriteString("@" + step1core.FormatSampleTimeText(*o))
						}
						buf.WriteString("=" + step1core.FormatChordNotesText(chord.GetNotes()) + "\n")
					}
				}
			}
		}
	}

	if devices := root.GetDevices(); devices != nil {
		for _, device := range devices.ListDevices() {
			buf.WriteString("device:" + string(device.GetName()) + "/" + string(device.GetClass()) + "/" + string(device.GetID()) + "\n")
			for _, wire := range device.ListWires() {
				buf.WriteString("  wire:" + string(wire.GetPort()) + "<-" + step1core.FormatSourceRefText(wire.GetSource()) + "\n")
			}
		}
	}

	return buf.String()
}

// requireSameSummary 断言两篇文档的结构等价.
func requireSameSummary(t *testing.T, a step1core.WorkElement, b step1core.WorkElement) {
	t.Helper()
	text1 := summary(a)
	text2 := summary(b)
	if text1 != text2 {
		t.Errorf("documents are not equivalent:\n--- a ---\n%s--- b ---\n%s", text1, text2)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 解码 : prototype.work.xml

func TestDOMDecodePrototypeRoot(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	requireEqual(t, "root.tag", root.GetTag(), step1core.TagWork)
	requireEqual(t, "root.namespace", root.GetNamespace(), step1core.Step1WorkNamespace)
	requireEqual(t, "root.prefix", root.GetPrefix(), "")
	requireEqual(t, "root.orderid", string(root.GetOrderID()), "")
	if root.GetParent() != nil {
		t.Errorf("root.GetParent() = %v, want nil", root.GetParent())
	}
	if root.GetOwner() != dom.Document {
		t.Errorf("root.GetOwner() != document")
	}

	// 根元素上的命名空间声明与 xsi:schemaLocation 都要保留
	requireEqual(t, "root@xmlns", attr(t, root, "xmlns"), step1core.Step1WorkNamespace)
	requireContains(t, "root@xsi:schemaLocation", attr(t, root, "xsi:schemaLocation"), "step1work.xsd")
	// work 的子元素是 xs:all, 解码按文档中的实际顺序保留
	requireSlice(t, "root.ListChildren()", tagsOf(root.ListChildren()),
		[]string{step1core.TagHead, step1core.TagChannels, step1core.TagTimeline,
			step1core.TagTracks, step1core.TagDevices})
	requireEqual(t, "len(ListChildrenByTag(timeline))", len(root.ListChildrenByTag(step1core.TagTimeline)), 1)
	requireEqual(t, "len(ListChildrenByTag(channels))", len(root.ListChildrenByTag(step1core.TagChannels)), 1)
}

// attr 读取一个必填属性.
func attr(t *testing.T, element step1core.DOMElement, name string) string {
	t.Helper()
	value, ok := element.GetAttribute(name)
	if !ok {
		t.Fatalf("attribute not found: %s", name)
	}
	return value
}

// tagsOf 取出各个元素的标签名.
func tagsOf(list []step1core.DOMElement) []string {
	tags := make([]string, 0, len(list))
	for _, item := range list {
		if item == nil {
			tags = append(tags, "")
			continue
		}
		tags = append(tags, item.GetTag())
	}
	return tags
}

func TestDOMDecodePrototypeHead(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	head := root.GetHead()
	if head == nil {
		t.Fatal("GetHead() = nil")
	}
	requireEqual(t, "title", head.GetTitle(), "Demo Step1 Work Document")
	requireEqual(t, "description", head.GetDescription(), "这是一个 Step1 工程文件的示例")
	requireEqual(t, "head.tag", head.GetTag(), step1core.TagHead)
	if parent := head.GetParent(); parent == nil || parent.GetTag() != step1core.TagWork {
		t.Errorf("head.GetParent() = %v", parent)
	}
	// 子元素的 owner 是同一个文档
	if head.GetOwner() != dom.Document {
		t.Errorf("head.GetOwner() != document")
	}

	metas := head.ListMeta()
	requireEqual(t, "len(meta)", len(metas), 6)

	bpm, err := head.FindMeta(step1core.MetaNameBPM)
	requireNoError(t, "FindMeta(bpm)", err)
	requireEqual(t, "meta(bpm).value", bpm.GetValue(), "88")
	requireEqual(t, "meta(bpm).name", bpm.GetName(), step1core.MetaNameBPM)

	sampleRate, err := head.FindMeta(step1core.MetaNameSampleRate)
	requireNoError(t, "FindMeta(sample_rate)", err)
	requireEqual(t, "meta(sample_rate).value", sampleRate.GetValue(), "44100")

	if _, err := head.FindMeta("no_such_meta"); err == nil {
		t.Errorf("FindMeta(no_such_meta) should fail")
	}

	// 包装对象的身份是稳定的
	if head != root.GetHead() {
		t.Errorf("GetHead() should return the same object")
	}
}

func TestDOMDecodePrototypeTimelines(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	top := root.GetTimeline()
	if top == nil {
		t.Fatal("GetTimeline() = nil")
	}
	requireEqual(t, "timeline.name", top.GetName(), step1core.NameType("root"))
	requireNilOptional(t, "timeline.bpm", top.GetBPM())

	work := top.ListTimelines()
	requireEqual(t, "len(work timeline)", len(work), 1)
	requireEqual(t, "work.name", work[0].GetName(), step1core.NameType("work"))
	requireOptional(t, "work.bpm", work[0].GetBPM(), step1core.BPM(120))
	requireOptional(t, "work.tempo.upper", tempoUpper(work[0].GetTempo()), uint(4))
	requireOptional(t, "work.tempo.lower", tempoLower(work[0].GetTempo()), uint(4))

	parts := work[0].ListTimelines()
	requireEqual(t, "len(parts)", len(parts), 2)

	requireEqual(t, "part1.name", parts[0].GetName(), step1core.NameType("part1"))
	requireNilOptional(t, "part1.bpm", parts[0].GetBPM())
	requireOptional(t, "part1.tempo.upper", tempoUpper(parts[0].GetTempo()), uint(3))
	requireOptional(t, "part1.tempo.lower", tempoLower(parts[0].GetTempo()), uint(8))

	requireEqual(t, "part2.name", parts[1].GetName(), step1core.NameType("part2"))
	requireOptional(t, "part2.bpm", parts[1].GetBPM(), step1core.BPM(88))
	requireOptional(t, "part2.offset", parts[1].GetOffset(), step1core.SampleTime(0))
	requireOptional(t, "part2.length", parts[1].GetLength(), step1core.SampleTime(12345))

	// 递归查找 (timeline/@name 在整篇文档中唯一)
	found, err := root.FindTimeline("part2")
	requireNoError(t, "FindTimeline(part2)", err)
	requireEqual(t, "FindTimeline(part2)", found, parts[1])
	if _, err := root.FindTimeline("no-such-timeline"); err == nil {
		t.Errorf("FindTimeline(no-such-timeline) should fail")
	}
}

// tempoUpper / tempoLower 取出拍号的分子 / 分母.
func tempoUpper(tempo *step1core.Tempo) *uint {
	if tempo == nil {
		return nil
	}
	value := tempo.Upper
	return &value
}

func tempoLower(tempo *step1core.Tempo) *uint {
	if tempo == nil {
		return nil
	}
	value := tempo.Lower
	return &value
}

func TestDOMDecodePrototypeChannels(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	channels := root.GetChannels()
	if channels == nil {
		t.Fatal("GetChannels() = nil")
	}
	requireEqual(t, "channels.tag", channels.GetTag(), step1core.TagChannels)

	list := channels.ListChannels()
	requireEqual(t, "len(channel)", len(list), 4)

	requireEqual(t, "channel[0].name", list[0].GetName(), step1core.NameType("audio_left"))
	requireEqual(t, "channel[0].type", list[0].GetType(), step1core.ChannelTypeAudio)
	requireEqual(t, "channel[0].tag", list[0].GetTag(), step1core.TagChannel)
	requireEqual(t, "channel[0].text", list[0].GetText(), "")

	requireEqual(t, "channel[1].name", list[1].GetName(), step1core.NameType("audio_right"))
	requireEqual(t, "channel[1].type", list[1].GetType(), step1core.ChannelTypeAudio)
	requireEqual(t, "channel[2].name", list[2].GetName(), step1core.NameType("midi_notes"))
	requireEqual(t, "channel[2].type", list[2].GetType(), step1core.ChannelTypeMIDI)
	requireEqual(t, "channel[3].name", list[3].GetName(), step1core.NameType("midi_clock"))
	requireEqual(t, "channel[3].type", list[3].GetType(), step1core.ChannelTypeMIDI)

	// 包装对象的身份是稳定的
	if channels != root.GetChannels() {
		t.Errorf("GetChannels() should return the same object")
	}

	// 按名字查找 (ChannelNameKey)
	found, err := root.FindChannel("midi_clock")
	requireNoError(t, "FindChannel(midi_clock)", err)
	requireEqual(t, "FindChannel(midi_clock)", found, list[3])
	if _, err := root.FindChannel("no-such-channel"); err == nil {
		t.Errorf("FindChannel(no-such-channel) should fail")
	}
}

func TestDOMDecodePrototypeTracks(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	tracks := root.GetTracks()
	if tracks == nil {
		t.Fatal("GetTracks() = nil")
	}
	list := tracks.ListTracks()
	requireEqual(t, "len(track)", len(list), 5)

	// 第 1 条: 音频轨, 两个 slice, 其中第一个带 wave
	requireEqual(t, "track[0].name", list[0].GetName(), step1core.NameType("sound"))
	requireEqual(t, "track[0].type", list[0].GetType(), step1core.TrackTypeAudio)

	slices := list[0].ListSlices()
	requireEqual(t, "len(track[0].slice)", len(slices), 2)
	requireEqual(t, "slice[0].timeline", slices[0].GetTimeline(), step1core.NameType("part1"))
	requireNilOptional(t, "slice[0].type", slices[0].GetType())

	wave := slices[0].GetWaveClip()
	if wave == nil {
		t.Fatal("slice[0].wave = nil")
	}
	requireEqual(t, "wave.src", wave.GetSrc(), "./foo.wav")
	requireEqual(t, "wave.tag", wave.GetTag(), step1core.TagWave)
	requireEqual[step1core.DOMElement](t, "slice[0].parent", slices[0].GetParent(), list[0])
	if slices[0].GetMidiClip() != nil || slices[0].GetProgression() != nil {
		t.Errorf("slice[0] should only contain a wave clip")
	}
	requireOptional(t, "slice[1].type", slices[1].GetType(), step1core.TrackTypeAudio)
	requireEqual(t, "slice[1].wave.src", slices[1].GetWaveClip().GetSrc(), "./bar.wav")

	// 第 2 条: MIDI 轨, src 为空串但属性仍然存在
	midi := list[1].ListSlices()[0].GetMidiClip()
	if midi == nil {
		t.Fatal("track[1] midi = nil")
	}
	requireEqual(t, "midi.src", midi.GetSrc(), "")
	requireEqual(t, "midi.type", list[1].GetType(), step1core.TrackTypeMIDI)

	// 第 3 条: 两个空的 slice
	requireEqual(t, "len(track[2].slice)", len(list[2].ListSlices()), 2)
	if list[2].ListSlices()[0].GetWaveClip() != nil {
		t.Errorf("track[2] slice[0] should be empty")
	}

	// 第 4 条: 和弦轨
	requireEqual(t, "track[3].type", list[3].GetType(), step1core.TrackTypeChord)
	progression := list[3].ListSlices()[0].GetProgression()
	if progression == nil {
		t.Fatal("progression = nil")
	}
	chords := progression.ListChords()
	requireEqual(t, "len(chord)", len(chords), 5)

	requireEqual(t, "chord[0].root", chords[0].GetRoot(), step1core.ToneNameC)
	requireOptional(t, "chord[0].offset", chords[0].GetOffset(), step1core.SampleTime(0))
	requireEqual(t, "chord[0].text", chords[0].GetText(), "C,E,G")
	requireSlice(t, "chord[0].notes", chords[0].GetNotes(),
		[]step1core.ToneName{step1core.ToneNameC, step1core.ToneNameE, step1core.ToneNameG})

	// "D#, F , A" : 逗号两侧的空白在解析时被忽略
	requireEqual(t, "chord[1].root", chords[1].GetRoot(), step1core.ToneNameDS)
	requireOptional(t, "chord[1].offset", chords[1].GetOffset(), step1core.SampleTime(100))
	requireSlice(t, "chord[1].notes", chords[1].GetNotes(),
		[]step1core.ToneName{step1core.ToneNameDS, step1core.ToneNameF, step1core.ToneNameA})

	requireOptional(t, "chord[4].offset", chords[4].GetOffset(), step1core.SampleTime(1000009))
	requireSlice(t, "chord[4].notes", chords[4].GetNotes(),
		[]step1core.ToneName{step1core.ToneNameG, step1core.ToneNameB, step1core.ToneNameD})

	// 第 5 条: master 轨
	requireEqual(t, "track[4].type", list[4].GetType(), step1core.TrackTypeMaster)

	// 按名字查找
	found, err := root.FindTrack("ch1")
	requireNoError(t, "FindTrack(ch1)", err)
	requireEqual(t, "FindTrack(ch1)", found, list[3])
	if _, err := root.FindTrack("no-such-track"); err == nil {
		t.Errorf("FindTrack(no-such-track) should fail")
	}
}

func TestDOMDecodePrototypeDevices(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	devices := root.GetDevices()
	if devices == nil {
		t.Fatal("GetDevices() = nil")
	}
	list := devices.ListDevices()
	requireEqual(t, "len(device)", len(list), 5)

	requireEqual(t, "device[0].name", list[0].GetName(), step1core.DeviceName("engine1"))
	requireEqual(t, "device[0].class", list[0].GetClass(), step1core.DeviceClass("engine"))
	requireEqual(t, "device[0].id", string(list[0].GetID()), "")

	requireEqual(t, "device[2].name", list[2].GetName(), step1core.DeviceName("adda1"))
	requireEqual(t, "len(device[2].wire)", len(list[2].ListWires()), 1)
	requireEqual(t, "device[2].wire[0].port", list[2].ListWires()[0].GetPort(), step1core.PortName("audio_output"))
	requireEqual(t, "device[2].wire[0].source", step1core.FormatSourceRefText(list[2].ListWires()[0].GetSource()), "mix1#monitor1")

	requireEqual(t, "device[4].name", list[4].GetName(), step1core.DeviceName("mix1"))
	requireEqual(t, "device[4].class", list[4].GetClass(), step1core.DeviceClass("mixer"))
	requireEqual(t, "device[4].id", list[4].GetID(), step1core.DeviceID("mix-1"))

	wires := list[4].ListWires()
	requireEqual(t, "len(device[4].wire)", len(wires), 2)
	requireEqual(t, "wire[0].port", wires[0].GetPort(), step1core.PortName("ch1"))
	requireEqual(t, "wire[0].source", wires[0].GetSource(),
		step1core.PortRef{Device: step1core.DeviceName("gain1"), Port: step1core.PortName("output")})
	requireEqual(t, "wire[1].source", wires[1].GetSource(),
		step1core.PortRef{Device: step1core.DeviceName("adda1"), Port: step1core.PortName("ch1")})
	requireEqual(t, "wire[0].tag", wires[0].GetTag(), step1core.TagWire)

	found, err := root.FindDevice("mix1")
	requireNoError(t, "FindDevice(mix1)", err)
	requireEqual(t, "FindDevice(mix1)", found, list[4])

	byID, err := root.FindDeviceByID("mix-1")
	requireNoError(t, "FindDeviceByID(mix-1)", err)
	requireEqual(t, "FindDeviceByID(mix-1)", byID, list[4])

	if _, err := root.FindDevice("no-such-device"); err == nil {
		t.Errorf("FindDevice(no-such-device) should fail")
	}
	if _, err := root.FindDeviceByID("no-such-id"); err == nil {
		t.Errorf("FindDeviceByID(no-such-id) should fail")
	}
	if _, err := root.FindDeviceByID(""); err == nil {
		t.Errorf("FindDeviceByID(\"\") should fail")
	}
}

////////////////////////////////////////////////////////////////////////////////
// 编码 : 往返一致性

func TestDOMEncodePrototype(t *testing.T) {

	dom, _ := loadPrototype(t)
	data, err := dom.Save()
	requireNoError(t, "Save", err)
	text := string(data)

	// XML 声明与命名空间声明
	if !strings.HasPrefix(text, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("missing xml declaration")
	}
	requireEqual(t, `count(xmlns="...")`,
		strings.Count(text, `xmlns="`+step1core.Step1WorkNamespace+`"`), 1)
	requireContains(t, "xsi:schemaLocation", text,
		`xsi:schemaLocation="`+step1core.Step1WorkNamespace+` step1work.xsd"`)

	// 文本与空元素
	requireContains(t, "title", text, "<title>Demo Step1 Work Document</title>")
	requireContains(t, "description", text, "<description>这是一个 Step1 工程文件的示例</description>")
	requireContains(t, "channel", text, `<channel name="audio_left" type="audio"/>`)
	requireContains(t, "midi", text, `<midi src=""/>`)
	requireContains(t, "empty slice", text, `<slice timeline="part2"/>`)
	requireContains(t, "chord", text, `<chord root="D#" offset="100">D#, F , A</chord>`)
	requireContains(t, "wire", text, `<wire port="ch1" source="gain1#output"/>`)

	// 注释不保留 (DOM 里没有对应的节点类型)
	if strings.Contains(text, "<!--") {
		t.Errorf("comments should be dropped:\n%s", text)
	}
}

func TestDOMEncodeDecodeRoundTrip(t *testing.T) {

	dom, _ := loadPrototype(t)
	data1, err := dom.Save()
	requireNoError(t, "Save 1", err)

	dom2 := newTestDOM()
	requireNoError(t, "Load 2", dom2.Load(data1))
	data2, err := dom2.Save()
	requireNoError(t, "Save 2", err)

	requireEqual(t, "round-trip text", string(data2), string(data1))
	requireSameSummary(t, requireRoot(t, dom), requireRoot(t, dom2))

	// 前缀不能被丢掉
	requireEqual(t, "reloaded root.prefix", requireRoot(t, dom2).GetPrefix(), "")

	// 往返之后仍然满足语义约束
	if errs := step1core.CheckWorkConstraints(requireRoot(t, dom2)); len(errs) != 0 {
		t.Errorf("constraints = %v", errs)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 工厂 : 从零创建文档

func TestDOMFactoryCreateDocument(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "build.work.xml")

	requireEqual(t, "doc.url", doc.GetURL(), "build.work.xml")
	requireEqual(t, "doc.owner", doc.GetOwner(), doc)
	if doc.Implementation() != dom.Implementation {
		t.Errorf("doc.Implementation() != dom.Implementation")
	}

	root := requireRoot(t, dom)
	requireEqual(t, "root.tag", root.GetTag(), step1core.TagWork)
	requireEqual(t, "root.namespace", root.GetNamespace(), step1core.Step1WorkNamespace)

	// 新建的文档只有一个空的根元素
	if len(root.ListChildren()) != 0 {
		t.Errorf("new document should be empty")
	}
	data, err := dom.Save()
	requireNoError(t, "Save", err)
	requireContains(t, "empty work", string(data), `<work xmlns="`+step1core.Step1WorkNamespace+`"/>`)
}

func TestDOMFactoryBuildDocument(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// head : title + description + meta (meta 必须排在 description 之前)
	head, err := factory.CreateHeadElement(doc)
	requireNoError(t, "CreateHeadElement", err)
	requireNoError(t, "SetTitle", head.SetTitle("Hello & <World>"))
	requireNoError(t, "SetDescription", head.SetDescription("a demo work"))
	meta, err := factory.CreateMetaElement(doc)
	requireNoError(t, "CreateMetaElement", err)
	requireNoError(t, "SetName", meta.SetName(step1core.MetaNameBPM))
	requireNoError(t, "SetValue", meta.SetValue("120"))
	requireNoError(t, "AddMeta", head.AddMeta(meta))
	requireNoError(t, "SetHead", root.SetHead(head))

	// timeline : root > part1
	timeline, err := factory.CreateTimelineElement(doc)
	requireNoError(t, "CreateTimelineElement", err)
	requireNoError(t, "timeline.SetName", timeline.SetName("root"))
	part, err := factory.CreateTimelineElement(doc)
	requireNoError(t, "CreateTimelineElement 2", err)
	requireNoError(t, "part.SetName", part.SetName("part1"))
	bpm := step1core.BPM(99.5)
	requireNoError(t, "part.SetBPM", part.SetBPM(&bpm))
	tempo := step1core.Tempo{Upper: 3, Lower: 4}
	requireNoError(t, "part.SetTempo", part.SetTempo(&tempo))
	length := step1core.SampleTime(480)
	requireNoError(t, "part.SetLength", part.SetLength(&length))
	requireNoError(t, "AddTimeline(part)", timeline.AddTimeline(part))
	requireNoError(t, "SetTimeline", root.SetTimeline(timeline))

	// channels : audio_left
	channels, err := factory.CreateChannelsElement(doc)
	requireNoError(t, "CreateChannelsElement", err)
	channel, err := factory.CreateChannelElement(doc)
	requireNoError(t, "CreateChannelElement", err)
	requireNoError(t, "channel.SetName", channel.SetName("audio_left"))
	requireNoError(t, "channel.SetType", channel.SetType(step1core.ChannelTypeAudio))
	requireNoError(t, "channels.AddChannel", channels.AddChannel(channel))
	requireNoError(t, "root.SetChannels", root.SetChannels(channels))

	// tracks : sound > slice > wave
	tracks, err := factory.CreateTracksElement(doc)
	requireNoError(t, "CreateTracksElement", err)
	track, err := factory.CreateTrackElement(doc)
	requireNoError(t, "CreateTrackElement", err)
	requireNoError(t, "track.SetName", track.SetName("sound"))
	requireNoError(t, "track.SetType", track.SetType(step1core.TrackTypeAudio))
	slice, err := factory.CreateSliceElement(doc)
	requireNoError(t, "CreateSliceElement", err)
	requireNoError(t, "slice.SetTimeline", slice.SetTimeline("part1"))
	wave, err := factory.CreateWaveClipElement(doc)
	requireNoError(t, "CreateWaveClipElement", err)
	requireNoError(t, "wave.SetSrc", wave.SetSrc(`a&b"c.wav`))
	requireNoError(t, "slice.SetWaveClip", slice.SetWaveClip(wave))
	requireNoError(t, "track.AddSlice", track.AddSlice(slice))
	requireNoError(t, "tracks.AddTrack", tracks.AddTrack(track))
	requireNoError(t, "root.SetTracks", root.SetTracks(tracks))

	// devices : mix1 > wire
	devices, err := factory.CreateDevicesElement(doc)
	requireNoError(t, "CreateDevicesElement", err)
	device, err := factory.CreateDeviceElement(doc)
	requireNoError(t, "CreateDeviceElement", err)
	requireNoError(t, "device.SetName", device.SetName("mix1"))
	requireNoError(t, "device.SetClass", device.SetClass("mixer"))
	requireNoError(t, "device.SetID", device.SetID("mix-1"))
	wire, err := factory.CreateWireElement(doc)
	requireNoError(t, "CreateWireElement", err)
	requireNoError(t, "wire.SetPort", wire.SetPort("ch1"))
	requireNoError(t, "wire.SetSource", wire.SetSource(step1core.PortRef{Device: "gain1", Port: "output"}))
	requireNoError(t, "device.AddWire", device.AddWire(wire))
	requireNoError(t, "devices.AddDevice", devices.AddDevice(device))
	requireNoError(t, "root.SetDevices", root.SetDevices(devices))

	// work 的子元素按规范顺序写出 : head, timeline, channels, tracks, devices
	// (xsd 用 xs:all, 顺序自由; 这里只是要求输出稳定)
	requireSlice(t, "root.ListChildren()", tagsOf(root.ListChildren()),
		[]string{step1core.TagHead, step1core.TagTimeline, step1core.TagChannels,
			step1core.TagTracks, step1core.TagDevices})
	// head 内部 : title, meta*, description
	requireSlice(t, "head.ListChildren()", tagsOf(head.ListChildren()),
		[]string{step1core.TagTitle, step1core.TagMeta, step1core.TagDescription})

	// 语义约束
	if errs := step1core.CheckWorkConstraints(root); len(errs) != 0 {
		t.Errorf("constraints = %v", errs)
	}

	// 保存
	data, err := dom.Save()
	requireNoError(t, "Save", err)
	text := string(data)
	requireContains(t, "xmlns", text, `xmlns="`+step1core.Step1WorkNamespace+`"`)
	requireContains(t, "title escaped", text, "<title>Hello &amp; &lt;World&gt;</title>")
	requireContains(t, "description", text, "<description>a demo work</description>")
	requireContains(t, "channel", text, `<channel name="audio_left" type="audio"/>`)
	requireContains(t, "src escaped", text, `src="a&amp;b&quot;c.wav"`)
	requireContains(t, "tempo", text, `tempo="3/4"`)
	requireContains(t, "bpm", text, `bpm="99.5"`)
	requireContains(t, "wire", text, `<wire port="ch1" source="gain1#output"/>`)

	// 读回来结构一致, 转义也要还原
	dom2 := newTestDOM()
	requireNoError(t, "Load", dom2.Load(data))
	root2 := requireRoot(t, dom2)
	requireSameSummary(t, root, root2)
	requireEqual(t, "reloaded title", root2.GetHead().GetTitle(), "Hello & <World>")
	requireEqual(t, "reloaded description", root2.GetHead().GetDescription(), "a demo work")
	requireEqual(t, "reloaded src", root2.GetTracks().ListTracks()[0].ListSlices()[0].GetWaveClip().GetSrc(), `a&b"c.wav`)
	requireEqual(t, "reloaded length", step1core.FormatSampleTimeText(*root2.GetTimeline().ListTimelines()[0].GetLength()), "480")
	requireEqual(t, "reloaded channel", root2.GetChannels().ListChannels()[0].GetType(), step1core.ChannelTypeAudio)
}

func TestDOMSliceChoice(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	factory := dom.Factory

	slice, err := factory.CreateSliceElement(doc)
	requireNoError(t, "CreateSliceElement", err)
	requireNoError(t, "SetTimeline", slice.SetTimeline("part1"))

	wave, _ := factory.CreateWaveClipElement(doc)
	midi, _ := factory.CreateMidiClipElement(doc)
	progression, _ := factory.CreateProgressionElement(doc)

	requireNoError(t, "wave.SetSrc", wave.SetSrc("./a.wav"))
	requireEqual(t, "wave.src", wave.GetSrc(), "./a.wav")
	requireNoError(t, "midi.SetSrc", midi.SetSrc("./a.mid"))
	requireEqual(t, "midi.src", midi.GetSrc(), "./a.mid")

	// xs:choice : wave / midi / progression 三选一
	requireNoError(t, "SetWaveClip", slice.SetWaveClip(wave))
	if slice.GetWaveClip() == nil {
		t.Errorf("wave clip missing")
	}
	requireNoError(t, "SetMidiClip", slice.SetMidiClip(midi))
	if slice.GetWaveClip() != nil || slice.GetMidiClip() == nil {
		t.Errorf("SetMidiClip() should replace the wave clip")
	}
	requireNoError(t, "SetProgression", slice.SetProgression(progression))
	if slice.GetWaveClip() != nil || slice.GetMidiClip() != nil || slice.GetProgression() == nil {
		t.Errorf("SetProgression() should replace the midi clip")
	}
	requireEqual(t, "len(slice.children)", len(slice.ListChildren()), 1)

	// nil 表示清空
	requireNoError(t, "SetProgression(nil)", slice.SetProgression(nil))
	requireEqual(t, "len(slice.children)", len(slice.ListChildren()), 0)
}

func TestDOMProgressionChords(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	factory := dom.Factory

	progression, err := factory.CreateProgressionElement(doc)
	requireNoError(t, "CreateProgressionElement", err)

	chord, err := factory.CreateChordElement(doc)
	requireNoError(t, "CreateChordElement", err)
	requireNoError(t, "SetRoot", chord.SetRoot("db")) // 会被规范化成 "C#"
	requireEqual(t, "chord.root", chord.GetRoot(), step1core.ToneNameCS)
	offset := step1core.SampleTime(10086)
	requireNoError(t, "SetOffset", chord.SetOffset(&offset))
	requireNoError(t, "SetNotes", chord.SetNotes([]step1core.ToneName{
		step1core.ToneNameC, step1core.ToneNameE, step1core.ToneNameG,
	}))
	requireEqual(t, "chord.text", chord.GetText(), "C,E,G")
	requireNoError(t, "AddChord", progression.AddChord(chord))

	requireEqual(t, "len(chord)", len(progression.ListChords()), 1)
	requireEqual[step1core.DOMElement](t, "chord.parent", chord.GetParent(), progression)

	// 没有 offset 的和弦
	chord2, _ := factory.CreateChordElement(doc)
	requireNoError(t, "SetRoot 2", chord2.SetRoot(step1core.ToneNameA))
	requireNilOptional(t, "chord2.offset", chord2.GetOffset())
	requireNoError(t, "AddChord 2", progression.AddChord(chord2))

	// 文本可以被直接读写
	requireNoError(t, "SetText", chord2.SetText("A, C#, E"))
	requireSlice(t, "chord2.notes", chord2.GetNotes(),
		[]step1core.ToneName{step1core.ToneNameA, step1core.ToneNameCS, step1core.ToneNameE})
}

func TestDOMOptionalAttributes(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	factory := dom.Factory

	// timeline 的可选属性: 设置 -> 读取 -> 再删除
	timeline, _ := factory.CreateTimelineElement(doc)
	requireNoError(t, "SetName", timeline.SetName("root"))
	requireNilOptional(t, "bpm(before)", timeline.GetBPM())
	requireNilOptional(t, "tempo(before)", timeline.GetTempo())

	bpm := step1core.BPM(88)
	requireNoError(t, "SetBPM", timeline.SetBPM(&bpm))
	requireOptional(t, "bpm(after)", timeline.GetBPM(), bpm)
	requireNoError(t, "SetBPM(nil)", timeline.SetBPM(nil))
	requireNilOptional(t, "bpm(nil)", timeline.GetBPM())

	offset := step1core.SampleTime(0)
	requireNoError(t, "SetOffset", timeline.SetOffset(&offset))
	requireOptional(t, "offset", timeline.GetOffset(), step1core.SampleTime(0))
	requireNoError(t, "SetOffset(nil)", timeline.SetOffset(nil))
	requireNilOptional(t, "offset(nil)", timeline.GetOffset())

	// slice 的可选 type
	slice, _ := factory.CreateSliceElement(doc)
	trackType := step1core.TrackTypeMIDI
	requireNoError(t, "slice.SetType", slice.SetType(&trackType))
	requireOptional(t, "slice.type", slice.GetType(), step1core.TrackTypeMIDI)
	requireNoError(t, "slice.SetType(nil)", slice.SetType(nil))
	requireNilOptional(t, "slice.type(nil)", slice.GetType())

	// head 的可选 description : 空串表示删除
	head, _ := factory.CreateHeadElement(doc)
	requireEqual(t, "description(before)", head.GetDescription(), "")
	requireNoError(t, "SetDescription", head.SetDescription("a demo"))
	requireEqual(t, "description", head.GetDescription(), "a demo")
	requireNoError(t, "SetDescription 2", head.SetDescription("another"))
	requireEqual(t, "len(description)", len(head.ListChildrenByTag(step1core.TagDescription)), 1)
	requireNoError(t, "SetDescription(\"\")", head.SetDescription(""))
	requireEqual(t, "description(\"\")", head.GetDescription(), "")
	requireEqual(t, "len(description)", len(head.ListChildrenByTag(step1core.TagDescription)), 0)

	// device 的可选 id
	device, _ := factory.CreateDeviceElement(doc)
	requireNoError(t, "device.SetID", device.SetID("id-1"))
	requireEqual(t, "device.id", device.GetID(), step1core.DeviceID("id-1"))
	requireNoError(t, "device.SetID(\"\")", device.SetID(""))
	requireEqual(t, "device.id(\"\")", string(device.GetID()), "")

	// work 的 orderid 以及"空串表示删除"
	root := requireRoot(t, dom)
	requireNoError(t, "SetOrderID", root.SetOrderID("889923"))
	requireEqual(t, "orderid", root.GetOrderID(), step1core.OrderID("889923"))
	requireNoError(t, "SetOrderID(\"\")", root.SetOrderID(""))
	requireEqual(t, "orderid(\"\")", string(root.GetOrderID()), "")

	// 属性: 空串仍然写出, RemoveAttribute 之后不再写出
	requireNoError(t, "device.SetName", device.SetName("mix1"))
	requireNoError(t, "device.SetClass", device.SetClass("mixer"))
	requireNoError(t, "SetAttribute", device.SetAttribute("src", ""))
	if value, ok := device.GetAttribute("src"); !ok || value != "" {
		t.Errorf("GetAttribute(src) = %q, %v", value, ok)
	}
	requireNoError(t, "RemoveAttribute", device.RemoveAttribute("src"))
	if _, ok := device.GetAttribute("src"); ok {
		t.Errorf("src should be removed")
	}
}

////////////////////////////////////////////////////////////////////////////////
// 命名空间 : 前缀与扩展元素

func TestDOMNamespacePrefix(t *testing.T) {

	// 1. 解码时保留 xsi: 前缀与它的 URI
	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	found := false
	for _, item := range root.ListAttributes() {
		if item.Name == "schemaLocation" {
			found = true
			requireEqual(t, "attr.prefix", item.Prefix, "xsi")
			requireEqual(t, "attr.namespace", item.Namespace, "http://www.w3.org/2001/XMLSchema-instance")
		}
	}
	if !found {
		t.Errorf("xsi:schemaLocation not found in %+v", root.ListAttributes())
	}

	// 按 "前缀:名字" 或者 "名字" 都能读到属性
	if _, ok := root.GetAttribute("schemaLocation"); !ok {
		t.Errorf("GetAttribute(schemaLocation) should work")
	}

	// 2. 新建文档时, xsi: 前缀会自动补上声明
	dom2 := newTestDOM()
	requireCreate(t, dom2, "")
	root2 := requireRoot(t, dom2)
	requireNoError(t, "SetAttribute", root2.SetAttribute("xsi:schemaLocation", step1core.Step1WorkSchemaLocation))

	data, err := dom2.Save()
	requireNoError(t, "Save", err)
	text := string(data)
	requireContains(t, "xmlns:xsi", text, `xmlns:xsi="`+innerURIXSINamespace+`"`)
	requireContains(t, "xsi:schemaLocation", text, `xsi:schemaLocation="`+step1core.Step1WorkSchemaLocation+`"`)
	if strings.Contains(text, "xmlns:xmlns") {
		t.Errorf("broken namespace declaration:\n%s", text)
	}

	// 3. 读回来之后前缀与 URI 都不变
	dom3 := newTestDOM()
	requireNoError(t, "Load", dom3.Load(data))
	root3 := requireRoot(t, dom3)
	requireEqual(t, "reloaded xsi:schemaLocation", attr(t, root3, "xsi:schemaLocation"), step1core.Step1WorkSchemaLocation)
}

func TestDOMExtensionElement(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// 扩展点 (ExtensionPointType) 允许插入其它命名空间的子元素
	head, _ := factory.CreateHeadElement(doc)
	_ = head.SetTitle("t")
	requireNoError(t, "SetHead", root.SetHead(head))

	tracks, _ := factory.CreateTracksElement(doc)
	track, _ := factory.CreateTrackElement(doc)
	_ = track.SetName("sound")
	_ = track.SetType(step1core.TrackTypeAudio)
	slice, _ := factory.CreateSliceElement(doc)
	_ = slice.SetTimeline("root")
	wave, _ := factory.CreateWaveClipElement(doc)
	_ = wave.SetSrc("./a.wav")

	ext, err := factory.CreateElement(doc, "http://example.com/ext", "ext", "note")
	requireNoError(t, "CreateElement", err)
	requireEqual(t, "ext.tag", ext.GetTag(), "note")
	requireEqual(t, "ext.namespace", ext.GetNamespace(), "http://example.com/ext")
	requireEqual(t, "ext.prefix", ext.GetPrefix(), "ext")
	requireNoError(t, "ext.SetAttribute", ext.SetAttribute("kind", "demo"))
	requireNoError(t, "wave.AddChild", wave.AddChild(ext))
	if wave.GetParent() != nil {
		t.Errorf("wave.GetParent() = %v, want nil", wave.GetParent())
	}

	requireNoError(t, "slice.SetWaveClip", slice.SetWaveClip(wave))
	requireNoError(t, "track.AddSlice", track.AddSlice(slice))
	requireNoError(t, "tracks.AddTrack", tracks.AddTrack(track))
	requireNoError(t, "root.SetTracks", root.SetTracks(tracks))

	data, err := dom.Save()
	requireNoError(t, "Save", err)
	text := string(data)
	requireContains(t, "xmlns:ext", text, `xmlns:ext="http://example.com/ext"`)
	requireContains(t, "ext:note", text, `<ext:note xmlns:ext="http://example.com/ext" kind="demo"/>`)

	// 读回来: 扩展元素保持命名空间, 且不会被包装成 typed element
	dom2 := newTestDOM()
	requireNoError(t, "Load", dom2.Load(data))
	root2 := requireRoot(t, dom2)
	wave2 := root2.GetTracks().ListTracks()[0].ListSlices()[0].GetWaveClip()
	children := wave2.ListChildren()
	requireEqual(t, "len(wave.children)", len(children), 1)
	requireEqual(t, "ext.tag", children[0].GetTag(), "note")
	requireEqual(t, "ext.namespace", children[0].GetNamespace(), "http://example.com/ext")
	requireEqual(t, "ext.prefix", children[0].GetPrefix(), "ext")
	requireEqual(t, "ext@kind", attr(t, children[0], "kind"), "demo")
	if children[0].GetOwner() != dom2.Document {
		t.Errorf("ext.GetOwner() != document")
	}
}

////////////////////////////////////////////////////////////////////////////////
// 通用节点操作

func TestDOMNodeOperations(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// 通用元素: 文本 + 属性
	title, err := factory.CreateElement(doc, step1core.Step1WorkNamespace, "", step1core.TagTitle)
	requireNoError(t, "CreateElement", err)
	requireNoError(t, "SetText", title.SetText("Hello"))
	requireEqual(t, "title.text", title.GetText(), "Hello")
	requireNoError(t, "title.SetAttribute", title.SetAttribute("lang", "zh"))
	requireEqual(t, "title@lang", attr(t, title, "lang"), "zh")

	// 属性快照是副本
	attrs := title.ListAttributes()
	requireEqual(t, "len(attrs)", len(attrs), 1)
	attrs[0].Value = "changed"
	requireEqual(t, "title@lang(after)", attr(t, title, "lang"), "zh")

	// 子元素列表也是副本
	if list := root.ListChildren(); len(list) != 0 {
		t.Errorf("new root should have no children")
	}

	// AddChild / GetParent / RemoveChild
	requireNoError(t, "AddChild", root.AddChild(title))
	requireEqual[step1core.DOMElement](t, "title.parent", title.GetParent(), root)
	requireEqual(t, "len(root.children)", len(root.ListChildren()), 1)
	requireEqual(t, "root.children[0]", root.ListChildren()[0], title)
	if err := root.AddChild(title); err == nil {
		t.Errorf("AddChild() with the same child twice should fail")
	}
	requireNoError(t, "RemoveChild", root.RemoveChild(title))
	if title.GetParent() != nil {
		t.Errorf("title.GetParent() should be nil after RemoveChild")
	}
	if err := root.RemoveChild(title); err == nil {
		t.Errorf("RemoveChild() twice should fail")
	}

	// 加入新文档时会改属新的 owner
	docOfOther := newTestDOM()
	requireCreate(t, docOfOther, "")
	otherRoot := requireRoot(t, docOfOther)
	requireNoError(t, "otherRoot.AddChild", otherRoot.AddChild(title))
	if title.GetOwner() != docOfOther.Document {
		t.Errorf("child should be adopted by the new document")
	}

	// 列表按标签筛选
	head, _ := factory.CreateHeadElement(doc)
	_ = head.SetTitle("t")
	requireNoError(t, "SetHead", root.SetHead(head))
	requireEqual(t, "ListChildrenByTag(head)", len(root.ListChildrenByTag(step1core.TagHead)), 1)
	requireEqual(t, "ListChildrenByTag(tracks)", len(root.ListChildrenByTag(step1core.TagTracks)), 0)

	// ListChildren() 返回的已经是 typed element
	child := root.ListChildren()[0]
	typed, ok := child.(step1core.HeadElement)
	if !ok {
		t.Fatalf("child %T is not a HeadElement", child)
	}
	requireEqual(t, "typed head.title", typed.GetTitle(), "t")
}

func TestDOMReplaceAndRemoveSections(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	head, _ := factory.CreateHeadElement(doc)
	_ = head.SetTitle("first")
	timeline, _ := factory.CreateTimelineElement(doc)
	_ = timeline.SetName("root")
	channels, _ := factory.CreateChannelsElement(doc)
	tracks, _ := factory.CreateTracksElement(doc)
	devices, _ := factory.CreateDevicesElement(doc)
	requireNoError(t, "SetHead", root.SetHead(head))
	requireNoError(t, "SetTimeline", root.SetTimeline(timeline))
	requireNoError(t, "SetChannels", root.SetChannels(channels))
	requireNoError(t, "SetTracks", root.SetTracks(tracks))
	requireNoError(t, "SetDevices", root.SetDevices(devices))
	requireSlice(t, "children", tagsOf(root.ListChildren()),
		[]string{step1core.TagHead, step1core.TagTimeline, step1core.TagChannels,
			step1core.TagTracks, step1core.TagDevices})

	// 替换 head : 旧的会被移除, 新的插入在原来的位置 (最前面)
	head2, _ := factory.CreateHeadElement(doc)
	_ = head2.SetTitle("second")
	requireNoError(t, "SetHead 2", root.SetHead(head2))
	requireSlice(t, "children(after replace)", tagsOf(root.ListChildren()),
		[]string{step1core.TagHead, step1core.TagTimeline, step1core.TagChannels,
			step1core.TagTracks, step1core.TagDevices})
	requireEqual(t, "head.title", root.GetHead().GetTitle(), "second")
	requireEqual(t, "len(head)", len(root.ListChildrenByTag(step1core.TagHead)), 1)
	if head.GetParent() != nil {
		t.Errorf("replaced head should be detached")
	}

	// 替换 timeline : 在 work 下最多一个 (xs:all)
	timeline2, _ := factory.CreateTimelineElement(doc)
	_ = timeline2.SetName("root2")
	requireNoError(t, "SetTimeline 2", root.SetTimeline(timeline2))
	requireEqual(t, "len(timeline)", len(root.ListChildrenByTag(step1core.TagTimeline)), 1)
	requireEqual(t, "timeline.name", root.GetTimeline().GetName(), step1core.NameType("root2"))
	if timeline.GetParent() != nil {
		t.Errorf("replaced timeline should be detached")
	}
	found, err := root.FindTimeline("root2")
	requireNoError(t, "FindTimeline(root2)", err)
	requireEqual(t, "FindTimeline(root2)", found, timeline2)

	// nil 表示删除
	requireNoError(t, "SetHead(nil)", root.SetHead(nil))
	requireNoError(t, "SetTimeline(nil)", root.SetTimeline(nil))
	requireNoError(t, "SetChannels(nil)", root.SetChannels(nil))
	requireNoError(t, "SetTracks(nil)", root.SetTracks(nil))
	requireNoError(t, "SetDevices(nil)", root.SetDevices(nil))
	requireEqual(t, "len(children)", len(root.ListChildren()), 0)
	if root.GetHead() != nil || root.GetTimeline() != nil || root.GetChannels() != nil ||
		root.GetTracks() != nil || root.GetDevices() != nil {
		t.Errorf("sections should be removed")
	}
	if _, err := root.FindTimeline("root2"); err == nil {
		t.Errorf("FindTimeline(root2) after removal should fail")
	}
}

func TestDOMSetRoot(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)

	// 同一个根元素可以重复设置
	requireNoError(t, "SetRoot(same)", doc.SetRoot(root))
	requireEqual(t, "root", requireRoot(t, dom), root)

	// 新的根元素会替换旧的
	other := newTestDOM()
	requireCreate(t, other, "")
	newRoot := requireRoot(t, other)
	requireNoError(t, "SetRoot(other)", doc.SetRoot(newRoot))
	requireEqual(t, "new root", requireRoot(t, dom), newRoot)
	requireEqual(t, "new root.owner", newRoot.GetOwner(), doc)

	// 换成"属于另一个文档"的元素之后, 它的 owner 会被改成当前文档
	if got := requireRoot(t, dom).GetOwner(); got != doc {
		t.Errorf("root.GetOwner() = %v, want %v", got, doc)
	}
	if root.GetParent() != nil || newRoot.GetParent() != nil {
		t.Errorf("root elements should have no parent")
	}
}

////////////////////////////////////////////////////////////////////////////////
// 语义约束 : 与 CheckWorkConstraints() 配合

func TestDOMConstraintsOK(t *testing.T) {

	dom, _ := loadPrototype(t)
	root := requireRoot(t, dom)

	if errs := step1core.CheckWorkConstraints(root); len(errs) != 0 {
		t.Errorf("CheckWorkConstraints() = %v", errs)
	}
	errs, err := dom.CheckConstraints()
	requireNoError(t, "CheckConstraints", err)
	if len(errs) != 0 {
		t.Errorf("DOM.CheckConstraints() = %v", errs)
	}
}

func TestDOMConstraintsBroken(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// 两条同名的 timeline : work 下只能有一个 (xs:all), 用子时间轴制造重复
	rootTimeline, _ := factory.CreateTimelineElement(doc)
	_ = rootTimeline.SetName("dup")
	childTimeline, _ := factory.CreateTimelineElement(doc)
	_ = childTimeline.SetName("dup")
	_ = rootTimeline.AddTimeline(childTimeline)
	_ = root.SetTimeline(rootTimeline)

	// 两条同名的 channel
	channels, _ := factory.CreateChannelsElement(doc)
	for i := 0; i < 2; i++ {
		channel, _ := factory.CreateChannelElement(doc)
		_ = channel.SetName("c1")
		_ = channel.SetType(step1core.ChannelTypeAudio)
		_ = channels.AddChannel(channel)
	}
	_ = root.SetChannels(channels)

	// 两条同名的 track, 其中第一条的 slice 引用了不存在的 timeline
	tracks, _ := factory.CreateTracksElement(doc)
	for i := 0; i < 2; i++ {
		track, _ := factory.CreateTrackElement(doc)
		_ = track.SetName("t1")
		_ = track.SetType(step1core.TrackTypeAudio)
		if i == 0 {
			slice, _ := factory.CreateSliceElement(doc)
			_ = slice.SetTimeline("no-such-timeline")
			_ = track.AddSlice(slice)
		}
		_ = tracks.AddTrack(track)
	}
	_ = root.SetTracks(tracks)

	// 两台同名 (且同 id) 的 device
	devices, _ := factory.CreateDevicesElement(doc)
	for i := 0; i < 2; i++ {
		device, _ := factory.CreateDeviceElement(doc)
		_ = device.SetName("d1")
		_ = device.SetClass("gain")
		_ = device.SetID("id-1")
		_ = devices.AddDevice(device)
	}
	_ = root.SetDevices(devices)

	// 两条同名的 meta
	head, _ := factory.CreateHeadElement(doc)
	_ = head.SetTitle("t")
	for i := 0; i < 2; i++ {
		meta, _ := factory.CreateMetaElement(doc)
		_ = meta.SetName(step1core.MetaNameBPM)
		_ = meta.SetValue("120")
		_ = head.AddMeta(meta)
	}
	_ = root.SetHead(head)

	// 每条约束各有一处冲突
	errs := step1core.CheckWorkConstraints(root)
	requireEqual(t, "len(errors)", len(errs), 7)

	found := make(map[step1core.WorkConstraint]string)
	for _, item := range errs {
		found[item.Constraint] = item.Path
		if item.Value == "" {
			t.Errorf("constraint error should carry a value: %v", item)
		}
		requireContains(t, "error string", item.Error(), string(item.Constraint))
	}

	want := map[step1core.WorkConstraint]string{
		step1core.WorkConstraintTimelineNameKey:  "work/timeline[1]/timeline[1]",
		step1core.WorkConstraintSliceTimelineRef: "work/tracks/track[1]/slice[1]",
		step1core.WorkConstraintChannelNameKey:   "work/channels/channel[2]",
		step1core.WorkConstraintTrackNameKey:     "work/tracks/track[2]",
		step1core.WorkConstraintDeviceNameKey:    "work/devices/device[2]",
		step1core.WorkConstraintDeviceIdUnique:   "work/devices/device[2]",
		step1core.WorkConstraintMetaNameKey:      "work/head/meta[2]",
	}
	for constraint, path := range want {
		requireEqual(t, "constraint "+string(constraint), found[constraint], path)
	}
	requireEqual(t, "len(errors)", len(found), len(want))

	// 通过 DOM 门面也能检查
	errs, err := dom.CheckConstraints()
	requireNoError(t, "CheckConstraints", err)
	requireEqual(t, "len(DOM.CheckConstraints)", len(errs), len(want))
}

////////////////////////////////////////////////////////////////////////////////
// 错误处理

func TestDOMValidationErrors(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// 各 simpleType 的取值范围
	if err := root.SetOrderID("abc"); err == nil {
		t.Errorf("SetOrderID(abc) should fail")
	}

	timeline, _ := factory.CreateTimelineElement(doc)
	if err := timeline.SetName("1bad"); err == nil {
		t.Errorf("SetName(1bad) should fail")
	}
	if err := timeline.SetName(""); err == nil {
		t.Errorf("SetName(\"\") should fail")
	}
	tooBig := step1core.BPM(9999)
	if err := timeline.SetBPM(&tooBig); err == nil {
		t.Errorf("SetBPM(9999) should fail")
	}
	if err := timeline.SetTempo(&step1core.Tempo{Upper: 0, Lower: 4}); err == nil {
		t.Errorf("SetTempo(0/4) should fail")
	}
	negative := step1core.SampleTime(-1)
	if err := timeline.SetOffset(&negative); err == nil {
		t.Errorf("SetOffset(-1) should fail")
	}

	track, _ := factory.CreateTrackElement(doc)
	if err := track.SetType(step1core.TrackTypeUnknown); err == nil {
		t.Errorf("SetType(unknown) should fail")
	}

	slice, _ := factory.CreateSliceElement(doc)
	master := step1core.TrackTypeMaster
	if err := slice.SetType(&master); err == nil {
		t.Errorf("SetType(master) should fail")
	}
	if err := slice.SetTimeline("bad name"); err == nil {
		t.Errorf("SetTimeline(\"bad name\") should fail")
	}

	channel, _ := factory.CreateChannelElement(doc)
	if err := channel.SetName("bad name"); err == nil {
		t.Errorf("channel.SetName(\"bad name\") should fail")
	}
	if err := channel.SetName(""); err == nil {
		t.Errorf("channel.SetName(\"\") should fail")
	}
	if err := channel.SetType(step1core.ChannelTypeUnknown); err == nil {
		t.Errorf("channel.SetType(unknown) should fail")
	}

	wire, _ := factory.CreateWireElement(doc)
	if err := wire.SetSource(step1core.PortRef{}); err == nil {
		t.Errorf("SetSource(empty) should fail")
	}
	if err := wire.SetPort(""); err == nil {
		t.Errorf("SetPort(\"\") should fail")
	}

	chord, _ := factory.CreateChordElement(doc)
	if err := chord.SetRoot("H"); err == nil {
		t.Errorf("SetRoot(H) should fail")
	}
	if err := chord.SetNotes([]step1core.ToneName{"X"}); err == nil {
		t.Errorf("SetNotes(X) should fail")
	}

	meta, _ := factory.CreateMetaElement(doc)
	if err := meta.SetName("bad name"); err == nil {
		t.Errorf("SetName(\"bad name\") should fail")
	}

	device, _ := factory.CreateDeviceElement(doc)
	if err := device.SetName("bad name"); err == nil {
		t.Errorf("SetName(\"bad name\") should fail")
	}
	if err := device.SetClass("bad class"); err == nil {
		t.Errorf("SetClass(\"bad class\") should fail")
	}
	if err := device.SetID("bad id"); err == nil {
		t.Errorf("SetID(\"bad id\") should fail")
	}
}

func TestDOMStructuralErrors(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	// 类型不匹配的子元素
	requireNoError(t, "SetTimeline(nil)", root.SetTimeline(nil)) // nil 表示删除
	track, _ := factory.CreateTrackElement(doc)
	wrongTimeline := &domTimelineElement{track.(*domTrackElement).domElement} // 标签其实是 track
	if err := root.SetTimeline(wrongTimeline); err == nil {
		t.Errorf("SetTimeline(track) should fail")
	}
	wrongChannels := &domChannelsElement{track.(*domTrackElement).domElement} // 标签其实是 track
	if err := root.SetChannels(wrongChannels); err == nil {
		t.Errorf("SetChannels(track) should fail")
	}

	// 根元素必须是 work
	head, _ := factory.CreateHeadElement(doc)
	if err := doc.SetRoot(nil); err == nil {
		t.Errorf("SetRoot(nil) should fail")
	}
	wrongRoot := &domWorkElement{head.(*domHeadElement).domElement} // 标签其实是 head
	if err := doc.SetRoot(wrongRoot); err == nil {
		t.Errorf("SetRoot(head) should fail")
	}

	// 工厂参数
	if _, err := factory.CreateElement(doc, "", "", ""); err == nil {
		t.Errorf("CreateElement(empty tag) should fail")
	}
	if _, err := factory.CreateWorkElement(nil); err == nil {
		t.Errorf("CreateWorkElement(nil) should fail")
	}
	if _, err := factory.CreateTrackElement(nil); err == nil {
		t.Errorf("CreateTrackElement(nil) should fail")
	}

	// typed-nil 不能被当成子元素 (包一层接口以模拟调用方)
	var typedNil *domTrackElement
	var asElement step1core.DOMElement = typedNil
	if err := root.AddChild(asElement); err == nil {
		t.Errorf("AddChild(typed nil) should fail")
	}
	if _, ok := innerUnwrapElement(nil); ok {
		t.Errorf("innerUnwrapElement(nil) should be false")
	}
	if _, ok := innerUnwrapElement(asElement); ok {
		t.Errorf("innerUnwrapElement(typed nil) should be false")
	}
}

func TestDOMCodecErrors(t *testing.T) {

	// 空的 / 非法的 XML
	dom := newTestDOM()
	if err := dom.Load([]byte("   ")); err == nil {
		t.Errorf("Load(blank) should fail")
	}
	if err := dom.Load(nil); err == nil {
		t.Errorf("Load(nil) should fail")
	}
	if err := dom.Load([]byte("<html></html>")); err == nil {
		t.Errorf("Load(html) should fail")
	}
	if err := dom.Load([]byte("<work><a></work>")); err == nil {
		t.Errorf("Load(malformed) should fail")
	}
	if err := dom.Load([]byte("<work/><work/>")); err == nil {
		t.Errorf("Load(two roots) should fail")
	}

	// 没有实现时的错误
	empty := step1core.NewDOM(nil)
	if empty.Factory != nil || empty.Codec != nil {
		t.Errorf("NewDOM(nil) should have no factory/codec")
	}
	if _, err := empty.Create("x"); err == nil {
		t.Errorf("Create() without factory should fail")
	}
	if _, err := empty.Save(); err == nil {
		t.Errorf("Save() without codec should fail")
	}
	if err := empty.Load([]byte("<work/>")); err == nil {
		t.Errorf("Load() without codec should fail")
	}

	// 文档不存在时的错误
	dom2 := newTestDOM()
	if _, err := dom2.Save(); err == nil {
		t.Errorf("Save() without document should fail")
	}
	if _, err := dom2.GetRoot(); err == nil {
		t.Errorf("GetRoot() without document should fail")
	}
	if _, err := dom2.CheckConstraints(); err == nil {
		t.Errorf("CheckConstraints() without document should fail")
	}

	// 空文档: 有 document 但没有根元素
	var nilDoc *innerDOMDocument
	if _, err := nilDoc.GetRoot(); err == nil {
		t.Errorf("nil document.GetRoot() should fail")
	}
	if err := nilDoc.SetRoot(nil); err == nil {
		t.Errorf("nil document.SetRoot() should fail")
	}
	requireEqual(t, "nil document.GetURL()", nilDoc.GetURL(), "")
	if nilDoc.GetOwner() != nil {
		t.Errorf("(*innerDOMDocument)(nil).GetOwner() = %v, want nil", nilDoc.GetOwner())
	}

	// 文档节点的 owner 是它自己
	dom3 := newTestDOM()
	doc3 := requireCreate(t, dom3, "")
	requireEqual(t, "document.GetOwner()", doc3.GetOwner(), doc3)

	// 编解码器需要非空文档
	codec := GetDomImplementation().Codec()
	if _, err := codec.Encode(nil); err == nil {
		t.Errorf("Encode(nil) should fail")
	}
	if err := codec.Decode([]byte("<work/>"), nil); err == nil {
		t.Errorf("Decode(nil) should fail")
	}
}

func TestDOMLoadIntoExistingDocument(t *testing.T) {

	// DOM.Load() 时若还没有 document, 会自动新建一个
	data, err := os.ReadFile(domPrototypeFile)
	requireNoError(t, "read prototype", err)

	dom := newTestDOM()
	requireNoError(t, "Load", dom.Load(data))
	if dom.Document == nil {
		t.Fatal("Load() should create a document")
	}
	requireEqual(t, "title", requireRoot(t, dom).GetHead().GetTitle(), "Demo Step1 Work Document")

	// 向已经有内容的文档再 Load 一次: 根元素被整体替换
	small := []byte(`<work xmlns="` + step1core.Step1WorkNamespace + `"><head><title>small</title></head></work>`)
	requireNoError(t, "Load 2", dom.Load(small))
	root := requireRoot(t, dom)
	requireEqual(t, "title(2)", root.GetHead().GetTitle(), "small")
	if root.GetTimeline() != nil {
		t.Errorf("GetTimeline() = %v, want nil", root.GetTimeline())
	}
	requireEqual(t, "len(children)", len(root.ListChildren()), 1)
}

////////////////////////////////////////////////////////////////////////////////
// 转义 与 nil 保护

func TestDOMTextEscaping(t *testing.T) {

	dom := newTestDOM()
	doc := requireCreate(t, dom, "")
	root := requireRoot(t, dom)
	factory := dom.Factory

	head, err := factory.CreateHeadElement(doc)
	requireNoError(t, "CreateHeadElement", err)

	title := "a<b & c>d \"e\" 'f'\n\tg\rh"
	requireNoError(t, "SetTitle", head.SetTitle(title))
	requireNoError(t, "SetHead", root.SetHead(head))

	// xml: 是众所周知的前缀, 会自动补声明
	requireNoError(t, "SetAttribute(xml:lang)", root.SetAttribute("xml:lang", "zh"))

	// 普通元素的属性值也要转义
	note, err := factory.CreateElement(doc, step1core.Step1WorkNamespace, "", "note")
	requireNoError(t, "CreateElement", err)
	value := `x<y & z" w` + "\n\tv\r"
	requireNoError(t, "note.SetAttribute", note.SetAttribute("data", value))
	requireNoError(t, "AddChild", root.AddChild(note))

	data, err := dom.Save()
	requireNoError(t, "Save", err)
	text := string(data)

	requireContains(t, "xmlns:xml", text, `xmlns:xml="`+innerURIXMLNamespace+`"`)
	requireContains(t, "xml:lang", text, `xml:lang="zh"`)
	requireContains(t, "escaped text", text, "<title>a&lt;b &amp; c&gt;d \"e\" 'f'\n\tg&#13;h</title>")
	requireContains(t, "escaped attribute", text, `data="x&lt;y &amp; z&quot; w&#10;&#9;v&#13;"`)

	// 读回来必须是原文
	dom2 := newTestDOM()
	requireNoError(t, "Load", dom2.Load(data))
	root2 := requireRoot(t, dom2)
	requireEqual(t, "reloaded title", root2.GetHead().GetTitle(), title)
	requireEqual(t, "reloaded xml:lang", attr(t, root2, "xml:lang"), "zh")
	requireEqual(t, "reloaded data", attr(t, root2.ListChildren()[1], "data"), value)
}

func TestDOMNillable(t *testing.T) {

	// 内部节点: 所有只读方法都要能接受 nil 接收者
	var node *domElement
	requireEqual(t, "node.GetTag()", node.GetTag(), "")
	requireEqual(t, "node.GetPrefix()", node.GetPrefix(), "")
	requireEqual(t, "node.GetNamespace()", node.GetNamespace(), "")
	requireEqual(t, "node.GetText()", node.GetText(), "")
	if node.GetOwner() != nil {
		t.Errorf("node.GetOwner() = %v, want nil", node.GetOwner())
	}
	if node.GetParent() != nil {
		t.Errorf("node.GetParent() = %v, want nil", node.GetParent())
	}
	if node.ListAttributes() != nil {
		t.Errorf("node.ListAttributes() = %v, want nil", node.ListAttributes())
	}
	if node.ListChildren() != nil {
		t.Errorf("node.ListChildren() = %v, want nil", node.ListChildren())
	}
	if node.ListChildrenByTag(step1core.TagWork) != nil {
		t.Errorf("node.ListChildrenByTag() != nil")
	}
	if _, ok := node.GetAttribute("x"); ok {
		t.Errorf("node.GetAttribute() should be false")
	}
	if err := node.SetText("x"); err == nil {
		t.Errorf("node.SetText() should fail")
	}
	if err := node.SetAttribute("x", "y"); err == nil {
		t.Errorf("node.SetAttribute() should fail")
	}
	if err := node.RemoveAttribute("x"); err == nil {
		t.Errorf("node.RemoveAttribute() should fail")
	}
	if err := node.AddChild(nil); err == nil {
		t.Errorf("node.AddChild() should fail")
	}
	if err := node.RemoveChild(nil); err == nil {
		t.Errorf("node.RemoveChild() should fail")
	}

	// 文档: 只读方法也要能接受 nil 接收者
	var doc *innerDOMDocument
	requireEqual(t, "doc.GetURL()", doc.GetURL(), "")
	if doc.GetOwner() != nil {
		t.Errorf("doc.GetOwner() = %v, want nil", doc.GetOwner())
	}
	if doc.Implementation() != nil {
		t.Errorf("doc.Implementation() = %v, want nil", doc.Implementation())
	}

	// 包装: nil 节点的包装结果也是 nil
	if innerWrapElement(nil) != nil {
		t.Errorf("innerWrapElement(nil) != nil")
	}
	if _, err := innerWrapAs[step1core.WorkElement](nil); err == nil {
		t.Errorf("innerWrapAs(nil) should fail")
	}
	var typedNil *domTrackElement
	if err := innerAppendTypedChild(nil, typedNil, step1core.TagTrack); err == nil {
		t.Errorf("innerAppendTypedChild(nil, typed nil) should fail")
	}
}
