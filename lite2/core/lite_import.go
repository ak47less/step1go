package core

import "github.com/ak47less/step1go/common"

type Seconds = common.Seconds

type BPM = common.BPM

type Tempo = common.Tempo

// MIDI 按键编码, 例如:(value=69,A4,440Hz), (value=60,C4(middle_C),261.6Hz)
type PitchNumber = common.PitchNumber

type MidiNoteNumber = common.MidiNoteNumber

type MidiNoteVelocity = common.MidiNoteVelocity

type Logger = common.Logger

type GainLevel = common.GainDB

type GainDB = common.GainDB

type GainPercent = common.GainPercent

type GainRate = common.GainRate

// 表示 MIDI 通道号码, 值<0表示 omni
type MidiChannelNum = int8
