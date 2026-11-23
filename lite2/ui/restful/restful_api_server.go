package restful

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ak47less/step1go/lite2/lib"
)

////////////////////////////////////////////////////////////////////////////////

type ServerLoader struct {
	context *lib.ApplicationContext
	config  *lib.Configuration
	server  *innerWebServer
}

// GetLifeCycle implements [lib.Life].
func (inst *ServerLoader) GetLifeCycle() *lib.LifeCycle {
	lc := &lib.LifeCycle{
		OnInit:  inst.innerOnInit,
		OnStart: inst.innerOnStart,
		OnStop:  inst.innerOnStop,
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

	err = server.start()
	if err != nil {
		return err
	}

	inst.server = server
	return nil
}

func (inst *ServerLoader) innerOnStop() error {

	if inst.server == nil {
		return nil
	}
	return inst.server.stop()
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

	mu       sync.Mutex
	server   *http.Server
	listener net.Listener
	address  string
	started  bool

	cubeMu sync.RWMutex
}

func (inst *innerWebServer) init(ctx *lib.ApplicationContext, cfg *lib.WebServerConfig) error {

	if ctx == nil {
		return fmt.Errorf("application context is nil")
	}
	if cfg == nil {
		return fmt.Errorf("web server config is nil")
	}
	if cfg.Port < 0 || cfg.Port > 65535 {
		return fmt.Errorf("web server port %d is out of range", cfg.Port)
	}

	inst.context = ctx
	inst.host = cfg.Host
	inst.port = cfg.Port
	inst.controller = ctx.ActionController

	return nil
}

func (inst *innerWebServer) start() error {

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.started {
		return fmt.Errorf("web server is already started")
	}

	address := net.JoinHostPort(inst.host, strconv.Itoa(inst.port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}

	logger := inst.context.Logger
	logger.Info("serve HTTP RESTful API @%s", address)

	server := &http.Server{
		Handler:           inst.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	inst.server = server
	inst.listener = listener
	inst.address = listener.Addr().String()
	inst.started = true

	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			if inst.context != nil && inst.context.Logger != nil {
				inst.context.Logger.Error("REST server: %s", err)
			}
		}
	}()

	return nil
}

func (inst *innerWebServer) stop() error {

	inst.mu.Lock()
	server := inst.server
	inst.server = nil
	inst.listener = nil
	inst.address = ""
	inst.started = false
	inst.mu.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (inst *innerWebServer) serverAddress() string {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.address
}

////////////////////////////////////////////////////////////////////////////////
