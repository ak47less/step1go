package lib

////////////////////////////////////////////////////////////////////////////////

type ApplicationRunner interface {
	Run(ctx *ApplicationContext) error
}

////////////////////////////////////////////////////////////////////////////////

type DefaultAppRunner struct {
}

// Run implements [ApplicationRunner].
func (inst *DefaultAppRunner) Run(ctx *ApplicationContext) error {

	var err error
	life := ctx.LifeManager.GetMain()
	lc := life.GetLifeCycle()

	// create/destroy

	err = inst.invokeStage(ctx, "create", lc.OnCreate)
	if err != nil {
		return err
	}
	defer inst.invokeStage(ctx, "destroy", lc.OnDestroy)

	// start/stop

	err = inst.invokeStage(ctx, "start", lc.OnStart)
	if err != nil {
		return err
	}
	defer inst.invokeStage(ctx, "stop", lc.OnStop)

	// resume/pause

	err = inst.invokeStage(ctx, "resume", lc.OnResume)
	if err != nil {
		return err
	}
	defer inst.invokeStage(ctx, "pause", lc.OnPause)

	// loop

	return inst.invokeStage(ctx, "loop", lc.OnLoop)
}

func (inst *DefaultAppRunner) invokeStage(ctx *ApplicationContext, stageName string, fn LifeStageCallbackFunc) error {
	if inst == nil || ctx == nil || fn == nil {
		return nil
	}

	logger := ctx.Logger
	logger.Info("DefaultAppRunner: on stage '%s'", stageName)

	return fn()
}

// GetLifeCycle implements [Life].
func (inst *DefaultAppRunner) GetLifeCycle() *LifeCycle {
	lc := &LifeCycle{
		OnInit: inst.onInit,
	}
	return lc
}

func (inst *DefaultAppRunner) onInit(c *ApplicationContext) error {
	c.Runner = inst
	return nil
}

func (inst *DefaultAppRunner) _impl() (Life, ApplicationRunner) {
	return inst, inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
