package own

import (
	"fmt"
	"os"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
)

type ShScriptConfig struct {
	Name        string            `mapstructure:"name" desc:"Name for the target"`
	Description string            `mapstructure:"desc" desc:"Target description"`
	Labels      []string          `mapstructure:"labels" desc:"Labels to apply to the targets"` //
	Deps        []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv     []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv   []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env         map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools       map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility  []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Script      string            `mapstructure:"script"`
	Shell       *string           `mapstructure:"shell"`
	Args        []string          `mapstructure:"args"`
}

func (ec ShScriptConfig) GetTargets(tcc *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	interpolatedStringName, err := tcc.Interpolate(ec.Script, nil)
	if err != nil {
		return nil, fmt.Errorf("interpolating script: %w", err)
	}

	if ec.Shell == nil {
		ec.Shell = utils.StringPtr("/bin/sh")
	}

	pass_env := map[string]string{}
	for _, e := range ec.PassEnv {
		pass_env[e] = os.Getenv(e)
	}

	return []*zen_targets.Target{
		zen_targets.NewTarget(
			ec.Name,
			zen_targets.WithSrcs(map[string][]string{"_src": {interpolatedStringName}}),
			zen_targets.WithOuts([]string{interpolatedStringName}),
			zen_targets.WithVisibility(ec.Visibility),
			zen_targets.WithEnvVars(pass_env),
			zen_targets.WithBinary(),
			zen_targets.WithSecretEnvVars(ec.SecretEnv),
			zen_targets.WithTargetScript("run", &zen_targets.TargetScript{
				Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
					fullCommand := append([]string{*ec.Shell}, strings.Split(target.Srcs["_src"][0], " ")...)
					for _, a := range ec.Args {
						interpolatedArg, err := target.Interpolate(a)
						if err != nil {
							return fmt.Errorf("interpolating arg %s: %w", a, err)
						}

						fullCommand = append(fullCommand, interpolatedArg)
					}

					return target.Exec(fullCommand, "sh run")
				},
			}),
		),
	}, nil
}
