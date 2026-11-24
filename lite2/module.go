package lite2

import "fmt"

const (
	theModuleName     = "github.com/ak47less/step1go/lite2"
	theModuleVersion  = "v0.0.0"
	theModuleRevision = 0
)

type ModuleInfo struct {
	Name     string
	Version  string
	Revision int
}

func (inst *ModuleInfo) String() string {
	if inst == nil {
		return ""
	}
	const f = "[lite2.module name:'%s' version:'%s' rev:%d ]"
	return fmt.Sprintf(f, inst.Name, inst.Version, inst.Revision)
}

func GetModuleInfo() *ModuleInfo {
	i := &ModuleInfo{
		Name:     theModuleName,
		Version:  theModuleVersion,
		Revision: theModuleRevision,
	}
	return i
}
