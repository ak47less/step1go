package core

type TimePoint Seconds

type TimeLength Seconds

type TimelineType int

type Timeline struct {
	Type TimelineType

	Name string

	Parent *Timeline

	Offset TimePoint

	Length TimeLength

	Loop bool
}
