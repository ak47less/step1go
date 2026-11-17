package lib

type Life interface {
	GetLifeCycle() *LifeCycle
}

type LifeCallbackFunc func() error

type LifeCycle struct {
	OnInit func(ctx *AppContext) error

	OnCreate  LifeCallbackFunc
	OnStart   LifeCallbackFunc
	OnResume  LifeCallbackFunc
	OnPause   LifeCallbackFunc
	OnStop    LifeCallbackFunc
	OnDestroy LifeCallbackFunc

	Order int
}

type LifeManager interface {
	AddCycle(lc *LifeCycle) error

	AddLife(l Life) error

	GetMain() Life
}
