package lib

import "github.com/ak47less/step1go/lite2/core"

////////////////////////////////////////////////////////////////////////////////

type MidiEventType int

const (
	MET_NONE MidiEventType = iota

	MET_NOTE_ON
	MET_NOTE_OFF
	MET_CC
)

////////////////////////////////////////////////////////////////////////////////

type MidiEvent struct {

	// 事件类型
	Type MidiEventType

	// 该事件的触发时刻, 如果这个值 <0 , 表示立即触发的实时事件
	Time core.AbsolutePosition

	// 该事件所属的 MIDI 通道
	Channel core.MidiChannelNum

	//  扩展数据

	Note *MidiNote
	CC   *MidiCC
}

////////////////////////////////////////////////////////////////////////////////

type MidiNote struct {
	Number core.MidiNoteNumber

	Velocity core.MidiNoteVelocity
}

type MidiCC struct {
	Key int

	Value int
}

////////////////////////////////////////////////////////////////////////////////
