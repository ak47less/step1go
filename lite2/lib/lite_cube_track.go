package lib

type TrackIndex uint

type Track struct {
	Cube *Cube
}

type TrackList []*Track

type TrackArray TrackList

type TrackSlice TrackList
