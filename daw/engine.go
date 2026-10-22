package step1core

import (
	"fmt"
	"math"
	"time"
)

type Engine interface {
	GetContext() *EngineContext

	GetConfiguration() *EngineConfiguration

	GetTransport() Transport

	GetDeviceManager() DeviceManager

	Reset() error

	Start() error

	Stop() error

	Join() error

	//////////////////

	// Init(cfg *EngineConfiguration) error
	// Load() error
	// Create() error
	// Destroy() error
	// Run() error

}

type EngineFactory interface {
	CreateEngine(cfg *EngineConfiguration) (Engine, error)
}

type EngineLS interface {

	// 根据 DOM 加载引擎中的各个组件
	Load(ctx *EngineContext) error

	// 按照引擎中的各个组件, 生成相应的DOM
	Save(ctx *EngineContext) error
}

////////////////////////////////////////////////////////////////////////////////

type EngineState struct {
	StartedAt UnixTime
	StoppedAt UnixTime

	Starting bool
	Stopping bool

	Started bool
	Stopped bool
}

func (inst *EngineState) innerWaitForFn(timeout time.Duration, fn func() bool) error {

	const step = time.Second
	const max = math.MaxInt64

	if timeout < 0 {
		timeout = max
	}

	ttl := timeout

	for ttl > 0 {
		if fn() {
			return nil
		}
		ttl -= step
		time.Sleep(step)
	}

	return fmt.Errorf("Timeout")
}

func (inst *EngineState) WaitForStopping(timeout time.Duration) error {
	fn := func() bool {
		return inst.Stopping
	}
	return inst.innerWaitForFn(timeout, fn)
}

func (inst *EngineState) WaitForStopped(timeout time.Duration) error {
	fn := func() bool {
		return inst.Stopped
	}
	return inst.innerWaitForFn(timeout, fn)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
