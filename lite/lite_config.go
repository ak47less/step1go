package lite

import (
	"fmt"
	"net/url"
	"strconv"
)

type Configuration struct {
	Driver Driver

	InstrumentFactory InstrumentFactory

	SampleRate int

	Profile MusicProfile

	BankListSize  uint // the list default size
	StepListSize  uint // the list default size
	TrackListSize uint // the list default size

	BankListCapacity  uint // the list max size
	StepListCapacity  uint // the list max size
	TrackListCapacity uint // the list max size

	Parts []any // [Part|PartFactory|PartRegistration]

	Monitor *MonitorConfig // 如果这项配置为 nil, 表示不使用 Monitor
}

func DefaultConfig() *Configuration {

	cfg := new(Configuration)
	driver := NewMockDriver()

	cfg.Driver = driver
	cfg.SampleRate = 44100

	cfg.Profile = (new(SimpleProfile)).Init().SetSpeed(280)

	cfg.BankListCapacity = 8
	cfg.StepListCapacity = 128
	cfg.TrackListCapacity = 32

	cfg.BankListSize = 4
	cfg.StepListSize = 32
	cfg.TrackListSize = 8

	cfg.Monitor = nil

	return cfg
}

////////////////////////////////////////////////////////////////////////////////

type MonitorConfig struct {
	URL string

	Scheme string
	Host   string
	Port   int

	FlagLogSTC       bool
	FlagLogSeq       bool
	FlagLogTransport bool
	FlagLogEngine    bool
}

func (inst *MonitorConfig) Prepare() *MonitorConfig {

	if inst == nil {
		inst = new(MonitorConfig)
	}

	u1 := inst.URL

	if len(u1) > 0 {
		u2, err := url.Parse(u1)
		if err == nil && u2 != nil {
			portstr := u2.Port()
			portnum, _ := strconv.Atoi(portstr)
			inst.Scheme = u2.Scheme
			inst.Host = u2.Hostname()
			inst.Port = portnum
		}
	} else {

		s := inst.Scheme
		p := inst.Port
		h := inst.Host

		if s == "" {
			s = "std"
		}
		if h == "" {
			h = "127.0.0.1"
		}
		if p < 1 {
			p = 33309
		}

		u1 = fmt.Sprintf("%s://%s:%d", s, h, p)
	}

	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
