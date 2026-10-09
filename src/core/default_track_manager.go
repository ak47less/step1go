package core

import "github.com/ak47less/step1go"

type DefaultTrackManager struct {
}

// Count implements [step1go.TrackManager].
func (inst *DefaultTrackManager) Count() uint {
	panic("unimplemented")
}

// GetTrack implements [step1go.TrackManager].
func (inst *DefaultTrackManager) GetTrack(index uint) step1go.Track {
	panic("unimplemented")
}

func (inst *DefaultTrackManager) _impl() step1go.TrackManager {
	return inst
}
