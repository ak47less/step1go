package lite

import (
	"fmt"
	"net/url"
	"strconv"
)

type Configuration struct {
	Driver Driver

	SampleRate int

	BPM   BPM
	Tempo Tempo

	CountBankMax  uint
	CountStepMax  uint
	CountTrackMax uint

	Monitor *MonitorConfig // 如果这项配置为 nil, 表示不使用 Monitor
}

func DefaultConfig() *Configuration {

	cfg := new(Configuration)
	driver := NewMockDriver()

	cfg.Driver = driver
	cfg.SampleRate = 44100
	cfg.BPM = 100
	cfg.Tempo.Upper = 4
	cfg.Tempo.Lower = 4

	cfg.CountBankMax = 16
	cfg.CountStepMax = 32
	cfg.CountTrackMax = 16

	cfg.Monitor = nil

	return cfg
}

////////////////////////////////////////////////////////////////////////////////

type MonitorConfig struct {
	URL string

	Scheme string
	Host   string
	Port   int
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
