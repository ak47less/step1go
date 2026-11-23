package restful

import (
	"fmt"

	"github.com/ak47less/step1go/lite2/lib"
)

////////////////////////////////////////////////////////////////////////////////

type ServerLoader struct {
	context *lib.ApplicationContext
	config  *lib.Configuration
}

// GetLifeCycle implements [lib.Life].
func (inst *ServerLoader) GetLifeCycle() *lib.LifeCycle {
	lc := &lib.LifeCycle{
		OnInit:  inst.innerOnInit,
		OnStart: inst.innerOnStart,
	}
	return lc
}

func (inst *ServerLoader) innerOnInit(ac *lib.ApplicationContext) error {

	cfg := ac.Configuration

	inst.context = ac
	inst.config = cfg

	return nil
}

func (inst *ServerLoader) innerOnStart() error {
	cfg := inst.config
	web := cfg.WebServer
	if web != nil {
		if web.Enabled {
			return inst.innerStartServer(web)
		}
	}
	return nil
}

func (inst *ServerLoader) innerStartServer(conf *lib.WebServerConfig) error {

	ctx := inst.context
	server := new(innerWebServer)
	err := server.init(ctx, conf)

	if err != nil {
		return err
	}

	return server.start()
}

func (inst *ServerLoader) _impl() lib.Life {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type innerWebServer struct {
	host       string
	port       int
	context    *lib.ApplicationContext
	controller lib.ActionController
}

func (inst *innerWebServer) init(ctx *lib.ApplicationContext, cfg *lib.WebServerConfig) error {

	inst.context = ctx
	inst.host = cfg.Host
	inst.port = cfg.Port
	inst.controller = ctx.ActionController

	return nil
}

func (inst *innerWebServer) start() error {

	return fmt.Errorf("no impl")
}

////////////////////////////////////////////////////////////////////////////////
