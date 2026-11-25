package lib

import "github.com/ak47less/step1go/lite2/webaudioapi"

type Mixer interface {
	Update(ctx *ApplicationContext)

	CountChannels() int

	GetChannel(index int) (MixerChannel, error)

	OpenChannel(index int) (MixerChannel, error)

	GetMasterChannel() MixerChannel
}

type MixerChannel interface {
	GetInfo(dst *MixerChannelInfo) *MixerChannelInfo

	SetGain(gain GainDB)

	SetSolo(sole bool)

	SetMute(mute bool)

	Enable(en bool)

	GetInput() webaudioapi.AudioNode

	GetOutput() webaudioapi.AudioNode
}

type MixerChannelInfo struct {
	Name string

	Label string

	Enabled bool

	Solo bool

	Mute bool

	Gain GainDB
}

type MixerChannelHolder interface {
	GetChannel() (MixerChannel, error)
}

type MixerChannelLoader interface {
	LoadChannel() (MixerChannel, error)
}
