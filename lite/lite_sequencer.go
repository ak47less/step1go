package lite

type Step struct {
	Index  int
	Weight byte
}

type StepList struct {
	list     []*Step
	count    int // 节拍(步奏)数量(有效值)
	capacity int // 节拍(步奏)数量(最大值)
}

type Track struct {
	Index int
	Name  string
	Label string

	Instrument Instrument
}

type TrackList struct {
	list     []*Track
	count    int // 轨道数量(有效值)
	capacity int // 轨道数量(最大值)
}

type Sequencer interface {
	GetSteps() []*Step
	GetTracks() []*Track
}

type SequencerContext struct {
	Sequencer Sequencer

	Tracks TrackList

	Steps StepList
}
