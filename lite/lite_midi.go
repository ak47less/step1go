package lite

type MidiNote struct {

	// At+Position 共同决定了该消息出现在哪个时刻; 如果 At==nil, 表示这是个即时消息

	At       *Timeline
	Position Seconds
	Key      MidiNoteNumber   // 音高
	Velocity MidiNoteVelocity // 力度(midi.按键速度)
}

type MidiMessage struct {
	NoteOn  *MidiNote
	NoteOff *MidiNote
}
