package core

import "github.com/ak47less/step1go/step1core"

type DefaultTrackManager struct {
}

// Count implements [step1core.TrackManager].
func (inst *DefaultTrackManager) Count() uint {
	panic("unimplemented")
}

// GetTrack implements [step1core.TrackManager].
func (inst *DefaultTrackManager) GetTrack(index uint) step1core.Track {
	panic("unimplemented")
}

func (inst *DefaultTrackManager) _impl() step1core.TrackManager {
	return inst
}
