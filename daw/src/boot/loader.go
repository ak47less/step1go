package boot

import (
	"github.com/ak47less/step1go/step1core"
	"github.com/ak47less/step1go/step1core/src/core"
)

type Loader struct {
}

func (inst *Loader) Load(cfg *step1core.EngineConfiguration) (step1core.Engine, error) {
	factory := new(core.DefaultEngineFactory)
	return factory.CreateEngine(cfg)
}
