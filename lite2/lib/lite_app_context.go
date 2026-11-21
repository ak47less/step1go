package lib

import (
	"sort"

	"github.com/ak47less/step1go/common"
	"github.com/ak47less/step1go/lite2/core"
)

////////////////////////////////////////////////////////////////////////////////

type ApplicationContext struct {

	// config & app

	Configuration *Configuration
	Properties    core.Properties
	Factory       ApplicationContextFactory
	Runner        ApplicationRunner

	// contexts

	EngineContext    *EngineContext
	SequencerContext *SequencerContext
	SessionContext   *SessionContext
	TransportContext *TransportContext

	// detail api

	Logger          core.Logger
	Parts           PartManager
	LifeManager     LifeManager
	Session         Session
	Engine          Engine
	Sequencer       Sequencer
	EventDispatcher EventDispatcher
	Transport       Transport
	Clock           Clock
}

////////////////////////////////////////////////////////////////////////////////

type ApplicationContextFactory interface {
	CreateApplicationContext(cfg *Configuration) (*ApplicationContext, error)
}

func NewApplicationContextFactory() ApplicationContextFactory {
	return new(innerApplicationContextFactory)
}

////////////////////////////////////////////////////////////////////////////////

type innerApplicationContextFactory struct {
}

// CreateApplicationContext implements [ApplicationContextFactory].
func (inst *innerApplicationContextFactory) CreateApplicationContext(cfg *Configuration) (*ApplicationContext, error) {

	ac := new(ApplicationContext)

	ac.Configuration = cfg
	ac.Factory = inst

	ac.EngineContext = new(EngineContext)
	ac.SessionContext = new(SessionContext)
	ac.SequencerContext = new(SequencerContext)
	ac.TransportContext = new(TransportContext)

	ac.Parts = NewPartManager()
	ac.LifeManager = NewLifeManager()
	ac.Logger = common.Log()

	ac.EngineContext.AC = ac
	ac.SessionContext.AC = ac
	ac.SequencerContext.AC = ac
	ac.TransportContext.AC = ac

	err := inst.innerApplyAll(cfg, ac)
	if err != nil {
		return nil, err
	}

	return ac, nil
}

func (inst *innerApplicationContextFactory) innerApplyAll(cfg *Configuration, ctx *ApplicationContext) error {

	var err error
	list := make([]func(ctx *ApplicationContext, cfg *Configuration) error, 0)

	list = append(list, inst.innerApplyConfigurators)
	list = append(list, inst.innerApplyProperties)
	list = append(list, inst.innerApplyParts)
	list = append(list, inst.innerInitLifeCycle)

	for _, fn := range list {
		err = fn(ctx, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerApplicationContextFactory) innerApplyProperties(ctx *ApplicationContext, cfg *Configuration) error {
	ctx.Properties = cfg.Properties
	return nil
}

func (inst *innerApplicationContextFactory) innerApplyConfigurators(ctx *ApplicationContext, cfg *Configuration) error {

	src := cfg.Configurators
	tmp := make([]*ConfiguratorRegistration, 0)

	for _, reg := range src {
		if reg == nil {
			continue
		}
		if reg.Configurator == nil {
			continue
		}
		tmp = append(tmp, reg)
	}

	sort.Slice(tmp, func(i, j int) bool {
		o1 := tmp[i]
		o2 := tmp[j]
		return (o1.Order < o2.Order)
	})

	for _, reg := range tmp {
		reg.Configurator.Configure(cfg)
	}

	return nil
}

func (inst *innerApplicationContextFactory) innerApplyParts(ctx *ApplicationContext, cfg *Configuration) error {

	var err error
	src := cfg.Parts
	dst1 := ctx.Parts
	dst2 := dst1.GetRegistry()

	for _, item := range src {
		err = dst2.RegisterAny(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerApplicationContextFactory) innerInitLifeCycle(ctx *ApplicationContext, cfg *Configuration) error {

	src := ctx.Parts
	dst := ctx.LifeManager
	namelist := src.ListNames()

	for _, name := range namelist {

		info, err := src.GetInfoByName(name)
		if err != nil {
			return err
		}

		life, ok := info.Part.(Life)
		if ok && (life != nil) {
			dst.AddLife(life)
		}
	}

	ml := dst.GetMain()
	mlc := ml.GetLifeCycle()
	return mlc.OnInit(ctx)
}

func (inst *innerApplicationContextFactory) _impl() ApplicationContextFactory {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
