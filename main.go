package own

import (
	zen_targets "github.com/zen-io/zen-core/target"
)

var KnownTargets = zen_targets.TargetCreatorMap{
	"sh_script": ShScriptConfig{},
}
