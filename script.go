package sh

import (
	"fmt"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
)

type ShScriptConfig struct {
	Name          string            `mapstructure:"name" zen:"yes" desc:"Name for the target"`
	Description   string            `mapstructure:"desc" zen:"yes" desc:"Target description"`
	Labels        []string          `mapstructure:"labels" zen:"yes" desc:"Labels to apply to the targets"`
	Deps          []string          `mapstructure:"deps" zen:"yes" desc:"Build dependencies"`
	PassEnv       []string          `mapstructure:"pass_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	PassSecretEnv []string          `mapstructure:"secret_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env           map[string]string `mapstructure:"env" zen:"yes" desc:"Key-Value map of static environment variables to be used"`
	Tools         map[string]string `mapstructure:"tools" zen:"yes" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility    []string          `mapstructure:"visibility" zen:"yes" desc:"List of visibility for this target"`
	Script        string            `mapstructure:"script"`
	Shell         *string           `mapstructure:"shell"`
	Args          []string          `mapstructure:"args"`
}

func (ec ShScriptConfig) GetTargets(tcc *zen_targets.TargetConfigContext) ([]*zen_targets.TargetBuilder, error) {
	interpolatedStringName, err := tcc.Interpolate(ec.Script, nil)
	if err != nil {
		return nil, fmt.Errorf("interpolating script: %w", err)
	}

	if ec.Shell == nil {
		ec.Shell = utils.StringPtr("/bin/sh")
	}
	ec.Labels = append(ec.Labels, fmt.Sprintf("shell=%s", *ec.Shell))

	for _, a := range ec.Args {
		ec.Labels = append(ec.Labels, fmt.Sprintf("arg=%s", a))
	}

	t := zen_targets.ToTarget(ec)
	t.Srcs = map[string][]string{"_src": {interpolatedStringName}}
	t.Outs = []string{interpolatedStringName}
	t.Scripts["run"] = &zen_targets.TargetBuilderScript{}

	return []*zen_targets.TargetBuilder{t}, nil
}

func ScriptRun(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
	var shell string
	args := make([]string, 0)
	for _, l := range target.Labels {
		if strings.HasPrefix(l, "shell=") {
			shell = strings.TrimPrefix(l, "shell=")
		} else if strings.HasPrefix(l, "arg=") {
			args = append(args, strings.TrimPrefix(l, "arg="))
		}
	}

	fullCommand := append([]string{shell}, strings.Split(target.Srcs["_src"][0], " ")...)
	for _, a := range args {
		interpolatedArg, err := target.Interpolate(a)
		if err != nil {
			return fmt.Errorf("interpolating arg %s: %w", a, err)
		}

		fullCommand = append(fullCommand, interpolatedArg)
	}

	return target.Exec(fullCommand, "sh run")
}
