package own

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
	"gitlab.com/hidothealth/platform/ahoy/src/utils"
)

type ShScriptConfig struct {
	ahoy_targets.BaseFields `mapstructure:",squash"`
	Script                  string   `mapstructure:"script"`
	Shell                   *string  `mapstructure:"shell"`
	Args                    []string `mapstructure:"args"`
}

func (ec ShScriptConfig) GetTargets(tcc *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
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

	return []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			ec.Name,
			ahoy_targets.WithSrcs(map[string][]string{"_src": {interpolatedStringName}}),
			ahoy_targets.WithOuts([]string{interpolatedStringName}),
			ahoy_targets.WithVisibility(ec.Visibility),
			ahoy_targets.WithEnvVars(pass_env),
			ahoy_targets.WithBinary(),
			ahoy_targets.WithSecretEnvVars(ec.SecretEnv),
			ahoy_targets.WithTargetScript("run", &ahoy_targets.TargetScript{
				Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
					env_vars := target.GetEnvironmentVariablesList()

					fullCommand := strings.Split(target.Srcs["_src"][0], " ")
					for _, a := range ec.Args {
						interpolatedArg, err := target.Interpolate(a)
						if err != nil {
							return fmt.Errorf("interpolating arg %s: %w", a, err)
						}

						fullCommand = append(fullCommand, interpolatedArg)
					}

					cmd := exec.Command(*ec.Shell, fullCommand...)
					cmd.Dir = target.Cwd
					cmd.Env = env_vars
					cmd.Stdout = target
					cmd.Stderr = target
					return cmd.Run()
				},
			}),
		),
	}, nil
}
