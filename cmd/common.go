package cmd

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/cobra"
)

var localTypes = []engine.EngineType{
	engine.EngineTypeDockerCore,
	engine.EngineTypeDockerAll,
	engine.EngineTypeDockerDistroless,
	engine.EngineTypeJvmSingleJar,
	engine.EngineTypeGolang,
}

// allEngineTypes is the distinct set of engine types used by --all flags.
// Only one docker variant is included to avoid duplicate results.
var allEngineTypes = []engine.EngineType{
	engine.EngineTypeDockerCore,
	engine.EngineTypeJvmSingleJar,
	engine.EngineTypeGolang,
}

// runWithRecovery executes fn, recovering from logger.Fatal calls that
// would otherwise exit the process. Returns an error if the function panicked.
func runWithRecovery(fn func()) (err error) {
	oldExitFunc := logger.ExitFunc
	logger.ExitFunc = func(code int) {
		panic(fmt.Sprintf("fatal exit with code %d", code))
	}
	defer func() {
		logger.ExitFunc = oldExitFunc
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	fn()
	return nil
}

func registerEngineTypeCompletions(cmd *cobra.Command, additionalTypes ...engine.EngineType) {
	_ = cmd.RegisterFlagCompletionFunc("engine-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var types []string
		for _, t := range localTypes {
			types = append(types, string(t))
		}
		if len(additionalTypes) > 0 {
			for _, t := range additionalTypes {
				types = append(types, string(t))
			}
		}
		return types, cobra.ShellCompDirectiveNoFileComp
	})
}
