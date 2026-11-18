package lib

////////////////////////////////////////////////////////////////////////////////

type LifeStageCallbackFunc func() error

type LifeInitCallbackFunc func(ctx *ApplicationContext) error

////////////////////////////////////////////////////////////////////////////////

type LifeCycle struct {
	OnInit LifeInitCallbackFunc

	OnCreate  LifeStageCallbackFunc
	OnStart   LifeStageCallbackFunc
	OnResume  LifeStageCallbackFunc
	OnPause   LifeStageCallbackFunc
	OnStop    LifeStageCallbackFunc
	OnDestroy LifeStageCallbackFunc

	Order int
}

////////////////////////////////////////////////////////////////////////////////

type Life interface {
	GetLifeCycle() *LifeCycle
}

type LifeManager interface {
	AddCycle(lc *LifeCycle) error

	AddLife(l Life) error

	GetMain() Life
}

////////////////////////////////////////////////////////////////////////////////

func NewLifeManager() LifeManager {
	return new(innerLifeManager)
}

////////////////////////////////////////////////////////////////////////////////

type innerLifeManager struct {
}

// AddCycle implements [LifeManager].
func (inst *innerLifeManager) AddCycle(lc *LifeCycle) error {
	panic("unimplemented")
}

// AddLife implements [LifeManager].
func (inst *innerLifeManager) AddLife(l Life) error {
	panic("unimplemented")
}

// GetMain implements [LifeManager].
func (inst *innerLifeManager) GetMain() Life {
	panic("unimplemented")
}

////////////////////////////////////////////////////////////////////////////////
// EOF
