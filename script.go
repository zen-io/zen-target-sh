package own

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
)

type ShScriptConfig struct {
	zen_targets.BaseFields `mapstructure:",squash"`
	Script                 string   `mapstructure:"script"`
	Shell                  *string  `mapstructure:"shell"`
	Args                   []string `mapstructure:"args"`
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
