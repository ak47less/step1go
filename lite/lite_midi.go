package lite

type MidiEventType = int

const (
	METypeMin MidiEventType = iota

	METypeNoteOn
	METypeNoteOff

	METypeMax
)

////////////////////////////////////////////////////////////////////////////////

type MidiNote struct {

	// 事件类型
	EventType MidiEventType

	//  如果 Position<0.001 , 表示这是个即时消息
	Position AbsolutePosition

	Channel  byte             // midi 通道
	Key      MidiNoteNumber   // 音高
	Velocity MidiNoteVelocity // 力度(midi.按键速度)
}

type MidiMessage struct {
	NoteOn  *MidiNote
	NoteOff *MidiNote
}
