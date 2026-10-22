package boot

import (
	"github.com/ak47less/step1go/step1core"
	"github.com/ak47less/step1go/step1core/src/core"

	_ "embed"
)

func GetDefaultConfig() *step1core.EngineConfiguration {

	cfg := new(step1core.EngineConfiguration)
	ldr := new(innerDefaultConfigLoader)

	cfg, err := ldr.load(cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

////////////////////////////////////////////////////////////////////////////////

//go:embed "config/default_step1.xml"
var theDefaultDomXml []byte

////////////////////////////////////////////////////////////////////////////////

type innerDefaultConfigLoader struct {
}

func (inst *innerDefaultConfigLoader) load(cfg *step1core.EngineConfiguration) (*step1core.EngineConfiguration, error) {

	if cfg == nil {
		cfg = new(step1core.EngineConfiguration)
	}

	dp := inst.innerGetDefaultProviders()
	dom, err := inst.innerGetDefaultDOM()

	if err != nil {
		return nil, err
	}

	cfg.DeviceProviders = dp
	cfg.DOM = dom
	cfg.MetaDefault = make(map[string]string)
	cfg.MetaFinal = make(map[string]string)
	cfg.MetaDoc = make(map[string]string)

	return cfg, nil
}

func (inst *innerDefaultConfigLoader) innerGetDefaultDOM() (*step1core.DOM, error) {

	imp := core.GetDomImplementation()
	dom := new(step1core.DOM)
	codec := imp.Codec()
	factory := imp.Factory()
	doc := factory.CreateDocument("embed://app/config/default_step1.xml")
	bin := theDefaultDomXml

	dom.Codec = codec
	dom.Factory = factory
	dom.Document = doc
	dom.Implementation = imp

	err := codec.Decode(bin, doc)

	return dom, err
}

func (inst *innerDefaultConfigLoader) innerGetDefaultProviders() step1core.DeviceProviderManager {
	dp := new(core.MainDeviceFactory)
	return dp
}

////////////////////////////////////////////////////////////////////////////////
// EOF
