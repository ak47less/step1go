package step1core

////////////////////////////////////////////////////////////////////////////////

type TrackType int

const (
	TrackTypeUnknown TrackType = iota

	TrackTypeMaster
	TrackTypeChord
	TrackTypeAudio
	TrackTypeMIDI
)

////////////////////////////////////////////////////////////////////////////////

type TrackInfo struct {
	Name string

	Type TrackType

	Track Track
}

type Track interface {
	GetInfo(info *TrackInfo) *TrackInfo
}

type TrackManager interface {
	Count() uint

	GetTrack(index uint) Track
}
