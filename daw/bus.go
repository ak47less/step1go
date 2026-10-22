package step1core

type Bus struct {
	Context *EngineContext

	Current TimePoint

	Router Router

	Channels []*Channel
}
