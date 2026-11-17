package lib

import "github.com/ak47less/step1go/lite2/core"

type EngineContext struct {
	Engine Engine

	FrameTimeline    *core.Timeline
	LoopTimeline     *core.Timeline
	EngineTimeline   *core.Timeline
	DocumentTimeline *core.Timeline
	RootTimeline     *core.Timeline
}

type Engine interface {
	GetEngineContext() *EngineContext
}
