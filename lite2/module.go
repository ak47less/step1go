package lite2

import (
	"fmt"
)

const (
	theModuleName     = "github.com/ak47less/step1go/lite2"
	theModuleVersion  = "v0.0.0"
	theModuleRevision = 0
)

func GetModuleInfo() string {

	const f = "[lite2.module name:'%s' version:'%s' rev:%d ]"
	return fmt.Sprintf(f, theModuleName, theModuleVersion, theModuleRevision)

}
